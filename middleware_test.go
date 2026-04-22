package maxrouter

import (
	"context"
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

var startUpdate = &schemes.MessageCreatedUpdate{
	Message: schemes.Message{Body: schemes.MessageBody{Text: "/start"}},
}

func TestMiddlewareChain(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	log := []string{}

	r.Use(Middleware(func(next RouteHandler) RouteHandler {
		return HandlerFunc(func(ctx Context) error {
			log = append(log, "mw1")
			return next.ServeContext(ctx)
		})
	}))
	r.Use(Middleware(func(next RouteHandler) RouteHandler {
		return HandlerFunc(func(ctx Context) error {
			log = append(log, "mw2")
			return next.ServeContext(ctx)
		})
	}))
	r.HandleCommand("/start", func(ctx Context) error {
		log = append(log, "handler")
		return nil
	})

	_ = r.HandleCtx(context.Background(), startUpdate)

	want := []string{"mw1", "mw2", "handler"}
	if len(log) != len(want) {
		t.Fatalf("ожидался порядок %v, получен %v", want, log)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Errorf("позиция %d: ожидалось %q, получено %q", i, want[i], log[i])
		}
	}
}

func TestGroupMiddleware(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	groupMwCalled := false

	r.Group(func(sub *Router) {
		sub.Use(Middleware(func(next RouteHandler) RouteHandler {
			return HandlerFunc(func(ctx Context) error {
				groupMwCalled = true
				return next.ServeContext(ctx)
			})
		}))
		sub.HandleCommand("/inside", func(ctx Context) error { return nil })
	})

	r.HandleCommand("/outside", func(ctx Context) error { return nil })

	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/inside"}},
	})

	if !groupMwCalled {
		t.Error("миддлварь группы должна была сработать для /inside")
	}

	groupMwCalled = false
	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/outside"}},
	})

	if groupMwCalled {
		t.Error("миддлварь группы не должна срабатывать для /outside")
	}
}

func TestNotFoundReceivesGlobalMiddleware(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	mwCalled := false
	notFoundCalled := false

	r.Use(Middleware(func(next RouteHandler) RouteHandler {
		return HandlerFunc(func(ctx Context) error {
			mwCalled = true
			return next.ServeContext(ctx)
		})
	}))
	r.NotFound(func(ctx Context) error {
		notFoundCalled = true
		return nil
	})

	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/nonexistent"}},
	})

	if !notFoundCalled {
		t.Error("NotFound-хэндлер должен был вызваться")
	}
	if !mwCalled {
		t.Error("глобальная миддлварь должна применяться и к NotFound-хэндлеру")
	}
}
