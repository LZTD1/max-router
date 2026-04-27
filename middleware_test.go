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

// --- Тесты на With ---

func TestWithAppliesMiddlewareOnlyToRoute(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	withMwCalled := false

	r.With(Middleware(func(next RouteHandler) RouteHandler {
		return HandlerFunc(func(ctx Context) error {
			withMwCalled = true
			return next.ServeContext(ctx)
		})
	})).HandleCommand("/scoped", func(ctx Context) error { return nil })

	r.HandleCommand("/plain", func(ctx Context) error { return nil })

	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/scoped"}},
	})
	if !withMwCalled {
		t.Error("миддлварь With должна сработать для /scoped")
	}

	withMwCalled = false
	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/plain"}},
	})
	if withMwCalled {
		t.Error("миддлварь With не должна срабатывать для /plain")
	}
}

func TestWithPreservesGlobalMiddlewareOrder(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	var log []string

	r.Use(Middleware(func(next RouteHandler) RouteHandler {
		return HandlerFunc(func(ctx Context) error {
			log = append(log, "global")
			return next.ServeContext(ctx)
		})
	}))

	r.With(Middleware(func(next RouteHandler) RouteHandler {
		return HandlerFunc(func(ctx Context) error {
			log = append(log, "with")
			return next.ServeContext(ctx)
		})
	})).HandleCommand("/ordered", func(ctx Context) error {
		log = append(log, "handler")
		return nil
	})

	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/ordered"}},
	})

	want := []string{"global", "with", "handler"}
	if len(log) != len(want) {
		t.Fatalf("ожидался порядок %v, получен %v", want, log)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Errorf("позиция %d: ожидалось %q, получено %q", i, want[i], log[i])
		}
	}
}

func TestWithDoesNotMutateRouter(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	withMwCalled := false

	r.With(Middleware(func(next RouteHandler) RouteHandler {
		return HandlerFunc(func(ctx Context) error {
			withMwCalled = true
			return next.ServeContext(ctx)
		})
	})).HandleCommand("/with-route", func(ctx Context) error { return nil })

	r.HandleCommand("/after-with", func(ctx Context) error { return nil })

	withMwCalled = false
	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/after-with"}},
	})
	if withMwCalled {
		t.Error("миддлварь With не должна просочиться в маршруты, зарегистрированные после неё на оригинальном роутере")
	}
}

// --- Тесты на изоляцию скопов между группами ---
func TestGroupsDoNotShareMiddlewareScope(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	groupAMwCalled := false
	groupBMwCalled := false

	r.Group(func(sub *Router) {
		sub.Use(Middleware(func(next RouteHandler) RouteHandler {
			return HandlerFunc(func(ctx Context) error {
				groupAMwCalled = true
				return next.ServeContext(ctx)
			})
		}))
		sub.HandleCommand("/group-a", func(ctx Context) error { return nil })
	})

	r.Group(func(sub *Router) {
		sub.Use(Middleware(func(next RouteHandler) RouteHandler {
			return HandlerFunc(func(ctx Context) error {
				groupBMwCalled = true
				return next.ServeContext(ctx)
			})
		}))
		sub.HandleCommand("/group-b", func(ctx Context) error { return nil })
	})

	// Запрос к группе A — только миддлварь A должна сработать
	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/group-a"}},
	})
	if !groupAMwCalled {
		t.Error("миддлварь группы A должна сработать для /group-a")
	}
	if groupBMwCalled {
		t.Error("миддлварь группы B не должна срабатывать для /group-a")
	}

	groupAMwCalled = false
	groupBMwCalled = false
	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/group-b"}},
	})
	if !groupBMwCalled {
		t.Error("миддлварь группы B должна сработать для /group-b")
	}
	if groupAMwCalled {
		t.Error("миддлварь группы A не должна срабатывать для /group-b")
	}
}

func TestGroupDoesNotLeakMiddlewareToParent(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	groupMwCalled := false

	r.Group(func(sub *Router) {
		sub.Use(Middleware(func(next RouteHandler) RouteHandler {
			return HandlerFunc(func(ctx Context) error {
				groupMwCalled = true
				return next.ServeContext(ctx)
			})
		}))
		sub.HandleCommand("/group-only", func(ctx Context) error { return nil })
	})

	r.HandleCommand("/root", func(ctx Context) error { return nil })

	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/root"}},
	})
	if groupMwCalled {
		t.Error("миддлварь группы не должна утекать в корневой роутер")
	}
}

func TestGroupInheritsParentMiddleware(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	globalMwCalled := false
	groupMwCalled := false

	r.Use(Middleware(func(next RouteHandler) RouteHandler {
		return HandlerFunc(func(ctx Context) error {
			globalMwCalled = true
			return next.ServeContext(ctx)
		})
	}))

	r.Group(func(sub *Router) {
		sub.Use(Middleware(func(next RouteHandler) RouteHandler {
			return HandlerFunc(func(ctx Context) error {
				groupMwCalled = true
				return next.ServeContext(ctx)
			})
		}))
		sub.HandleCommand("/child", func(ctx Context) error { return nil })
	})

	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/child"}},
	})

	if !globalMwCalled {
		t.Error("глобальная миддлварь родителя должна применяться к маршрутам группы")
	}
	if !groupMwCalled {
		t.Error("миддлварь группы должна сработать для /child")
	}
}

func TestTwoGroupsWithSharedParentMiddleware(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	var log []string

	r.Use(Middleware(func(next RouteHandler) RouteHandler {
		return HandlerFunc(func(ctx Context) error {
			log = append(log, "global")
			return next.ServeContext(ctx)
		})
	}))

	r.Group(func(sub *Router) {
		sub.Use(Middleware(func(next RouteHandler) RouteHandler {
			return HandlerFunc(func(ctx Context) error {
				log = append(log, "group-a-mw")
				return next.ServeContext(ctx)
			})
		}))
		sub.HandleCommand("/a", func(ctx Context) error {
			log = append(log, "handler-a")
			return nil
		})
	})

	r.Group(func(sub *Router) {
		sub.Use(Middleware(func(next RouteHandler) RouteHandler {
			return HandlerFunc(func(ctx Context) error {
				log = append(log, "group-b-mw")
				return next.ServeContext(ctx)
			})
		}))
		sub.HandleCommand("/b", func(ctx Context) error {
			log = append(log, "handler-b")
			return nil
		})
	})

	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/a"}},
	})
	wantA := []string{"global", "group-a-mw", "handler-a"}
	if len(log) != len(wantA) {
		t.Fatalf("/a: ожидался порядок %v, получен %v", wantA, log)
	}
	for i := range wantA {
		if log[i] != wantA[i] {
			t.Errorf("/a позиция %d: ожидалось %q, получено %q", i, wantA[i], log[i])
		}
	}

	log = nil
	_ = r.HandleCtx(context.Background(), &schemes.MessageCreatedUpdate{
		Message: schemes.Message{Body: schemes.MessageBody{Text: "/b"}},
	})
	wantB := []string{"global", "group-b-mw", "handler-b"}
	if len(log) != len(wantB) {
		t.Fatalf("/b: ожидался порядок %v, получен %v", wantB, log)
	}
	for i := range wantB {
		if log[i] != wantB[i] {
			t.Errorf("/b позиция %d: ожидалось %q, получено %q", i, wantB[i], log[i])
		}
	}
}
