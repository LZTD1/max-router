package _examples

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	maxrouter "github.com/LZTD1/max-router/v2"
	"github.com/LZTD1/max-router/v2/middleware"
	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
)

var ADMIN_IDS = map[int64]struct{}{
	111: {},
	222: {},
}

// RandomAccessMw предоставляет случайный доступ к команде
func RandomAccessMw(next maxrouter.RouteHandler) maxrouter.RouteHandler {
	return maxrouter.HandlerFunc(func(c maxrouter.Context) error {
		if rand.Intn(100) > 50 {
			return next.ServeContext(c)
		}
		log.Printf("[RandomAccessMw] access denied for user %d", c.UserID())
		return nil
	})
}

// LoggerMw логирует запросы пользователей
func LoggerMw(next maxrouter.RouteHandler) maxrouter.RouteHandler {
	return maxrouter.HandlerFunc(func(c maxrouter.Context) error {
		log.Printf("[LoggerMw] received update from user %d", c.UserID())
		return next.ServeContext(c)
	})
}

// AdminFilterMw фильтрует доступ только для администраторов
func AdminFilterMw(next maxrouter.RouteHandler) maxrouter.RouteHandler {
	return maxrouter.HandlerFunc(func(c maxrouter.Context) error {
		if _, ok := ADMIN_IDS[c.UserID()]; ok {
			return next.ServeContext(c)
		}
		log.Printf("[AdminFilterMw] access denied for user %d", c.UserID())
		return nil
	})
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	token := os.Getenv("MAX_TOKEN")
	if token == "" {
		log.Fatal("MAX_TOKEN не задан")
	}
	api, err := maxbot.NewApi(token)
	if err != nil {
		log.Fatal(err)
	}

	r := maxrouter.NewRouter(api)

	// --- Глобальный middleware ---
	r.Use(middleware.Recover())
	r.Use(LoggerMw)

	// --- Публичные маршруты ---
	r.HandleCommand("/start", func(c maxrouter.Context) error {
		return c.Send(fmt.Sprintf("Привет!\nTry:\n/rnd - случайный шанс получить доступ\n/ap - административная панель"))
	})

	// Маршрут с middleware случайного доступа
	r.With(RandomAccessMw).HandleCommand("/rnd", func(c maxrouter.Context) error {
		log.Printf("user %d accessed /rnd endpoint", c.UserID())
		return c.Send("Доступ выдан!")
	})

	// --- Группа для админских маршрутов ---
	r.Group(func(sub *maxrouter.Router) {
		sub.Use(AdminFilterMw) // Использование миддлвари на конкретную группу команд

		sub.HandleCommand("/ap", func(c maxrouter.Context) error {
			return c.Send("Вы можете использовать команды:\n\n/ap - административная панель\n/restart - рестарт сервера")
		})

		sub.HandleCommand("/restart", func(c maxrouter.Context) error {
			return c.Send("Сервер перезапускается...")
		})
	})

	// --- Обработчик неизвестных команд ---
	r.NotFound(func(c maxrouter.Context) error {
		return c.Send("Хмм... я не понимаю эту команду")
	})

	if err := r.RunPolling(ctx, maxrouter.WithPollingErrorHandler(func(err error) {
		log.Printf("polling error: %v", err)
	})); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}
