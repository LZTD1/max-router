package maxrouter

import (
	"context"
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func TestContextData(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	var gotCommand, gotText string

	r.HandleCommand("/help", func(ctx Context) error {
		gotCommand = ctx.Command()
		gotText = ctx.Text()
		return nil
	})

	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/help"}},
	})

	if gotCommand != "/help" {
		t.Errorf("Command(): ожидалось '/help', получено '%s'", gotCommand)
	}
	if gotText != "/help" {
		t.Errorf("Text(): ожидалось '/help', получено '%s'", gotText)
	}
}

func TestContextLocals(t *testing.T) {
	t.Run("set and get", func(t *testing.T) {
		r := NewRouter(nil, WithAsync(false))

		r.HandleCommand("/start", func(ctx Context) error {
			ctx.Set("key", "val")

			v, ok := ctx.Get("key")
			if !ok {
				t.Error("Get: ключ должен существовать после Set")
			}
			if v != "val" {
				t.Errorf("Get: ожидалось 'val', получено '%v'", v)
			}
			return nil
		})

		_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
			Message: schemes.Message{Body: schemes.MessageBody{Text: "/start"}},
		})
	})

	t.Run("get missing key", func(t *testing.T) {
		r := NewRouter(nil, WithAsync(false))

		r.HandleCommand("/start", func(ctx Context) error {
			_, ok := ctx.Get("ghost")
			if ok {
				t.Error("Get: несуществующий ключ не должен возвращать ok=true")
			}
			return nil
		})

		_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
			Message: schemes.Message{Body: schemes.MessageBody{Text: "/start"}},
		})
	})
}
