package _examples

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
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

	// -- Регистрация хендлеров --
	r.HandleCommand("/start", func(c maxrouter.Context) error {

		kb := maxbot.InlineKeyboard(
			maxbot.Row(
				maxbot.BtnLink("Репозиторий", "https://github.com/LZTD1/max-router"),
				maxbot.Btn("Помощь", "show_help_callback"),
			),
		)

		return c.Send(
			fmt.Sprintf("Привет!\n\nДобро пожаловать в бот."),
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
	for update := range api.GetUpdates(ctx) {
		r.Handle(update, ctx)
	}
}
