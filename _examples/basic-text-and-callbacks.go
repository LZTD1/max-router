package _examples

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	maxrouter "github.com/LZTD1/max-router/v2"
	"github.com/LZTD1/max-router/v2/middleware"
	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

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
	r.Use(middleware.Recover())

	// -- Регистрация хендлеров --
	r.HandleCommand("/start", func(c maxrouter.Context) error {

		kb := model.NewKeyboard()
		kb.AddRow().
			AddLink("Репозиторий", "https://github.com/LZTD1/max-router").
			AddCallBack("Помощь", "show_help_callback")

		return c.Send(
			fmt.Sprintf("Привет %s!\n\nДобро пожаловать в бот.", c.FullName()),
			maxrouter.WithKeyboard(kb),
		)
	})
	r.HandleText("Помогите", func(c maxrouter.Context) error {
		return c.Send("Доступные команды:\n/start — Показать стартовое меню\nПомогите — показать это сообщение")
	})

	r.HandleCallback("show_help_callback", func(c maxrouter.Context) error {
		if err := c.Answer("Показываю помощь!"); err != nil {
			log.Printf("err: %v", err)
		}

		return c.Send("Доступные команды:\n/start — Показать стартовое меню\nПомогите — показать это сообщение")
	})

	// Fallback
	r.NotFound(func(c maxrouter.Context) error {
		return c.Send("Хмм... я не понимаю эту команду")
	})
	r.OnError(func(err error, c maxrouter.Context) {
		log.Printf("err: %v", err)
		_ = c.Send("К сожалению, я не смог обработать вашу команду, подробности смотри в консоли.")
	})

	log.Println("Bot starting...")
	if err := r.RunPolling(ctx, maxrouter.WithPollingErrorHandler(func(err error) {
		log.Printf("polling error: %v", err)
	})); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}
