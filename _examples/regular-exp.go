package _examples

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	maxrouter "github.com/LZTD1/max-router"
	"github.com/LZTD1/max-router/middleware"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	api, err := maxbot.New(os.Getenv("MAX_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	r := maxrouter.NewRouter(api)
	r.Use(middleware.Recover())

	// --- Предкомпиляция регулярных выражений ---
	userCommandRegex := regexp.MustCompile(`^/user (\d+)$`)
	viewItemCallbackRegex := regexp.MustCompile(`^view_item:(.+)$`)

	// --- Обработчик текста по регулярному выражению ---
	r.HandleRegexpText(userCommandRegex, func(c maxrouter.Context) error {
		matches := userCommandRegex.FindStringSubmatch(c.Text())
		userID := "неизвестен"
		if len(matches) > 1 {
			userID = matches[1]
		}

		return c.Send(fmt.Sprintf("Обрабатываю запрос для пользователя с ID: %s", userID))
	})

	// --- Обработчик callback по регулярному выражению ---
	r.HandleRegexpCallback(viewItemCallbackRegex, func(c maxrouter.Context) error {
		callbackData := c.Data()

		if err := c.Answer("Загружаю детали товара..."); err != nil {
			log.Printf("Ошибка при ответе на callback %q: %v", callbackData, err)
		}

		matches := viewItemCallbackRegex.FindStringSubmatch(callbackData)
		itemID := "неизвестен"
		if len(matches) > 1 {
			itemID = matches[1]
		}

		return c.Send(fmt.Sprintf("Показываю детали товара: %s", itemID))
	})

	// --- Обработчик команды /start с инлайн-кнопками ---
	r.HandleCommand("/start", func(c maxrouter.Context) error {
		kb := maxbot.InlineKeyboard(
			maxbot.Row(
				maxbot.Btn("Товар ABC", "view_item:ABC"),
				maxbot.Btn("Товар 123", "view_item:123"),
			),
		)

		return c.Send(
			"Добро пожаловать! Попробуй команды вида `/user 123` или нажми на кнопку ниже.",
			maxrouter.WithKeyboard(kb),
		)
	})

	log.Println("Бот запускается...")
	for update := range api.GetUpdates(ctx) {
		r.Handle(update, ctx)
	}
}
