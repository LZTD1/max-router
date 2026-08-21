package maxrouter

import (
	"context"
	"regexp"
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

func TestExactMatch(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	startCalled := false
	helloCalled := false

	r.HandleCommand("/start", func(ctx Context) error {
		startCalled = true
		return nil
	})
	r.HandleText("Привет", func(ctx Context) error {
		helloCalled = true
		return nil
	})

	_ = r.HandleCtx(context.Background(), messageUpdate("/start"))

	if !startCalled {
		t.Error("хэндлер /start должен был вызваться")
	}
	if helloCalled {
		t.Error("хэндлер 'Привет' не должен был вызваться")
	}

	startCalled = false
	_ = r.HandleCtx(context.Background(), messageUpdate("Привет"))

	if !helloCalled {
		t.Error("хэндлер 'Привет' должен был вызваться")
	}
	if startCalled {
		t.Error("хэндлер /start не должен был вызваться")
	}
}

func TestRegexpMatch(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	pattern := regexp.MustCompile(`^/user_(\d+)$`)
	gotText := ""

	r.HandleRegexpText(pattern, func(ctx Context) error {
		gotText = ctx.Text()
		return nil
	})

	_ = r.HandleCtx(context.Background(), messageUpdate("/user_123"))

	matches := pattern.FindStringSubmatch(gotText)
	if len(matches) < 2 {
		t.Fatalf("хэндлер не вызвался или текст пустой, gotText=%q", gotText)
	}
	if matches[1] != "123" {
		t.Errorf("ожидался ID '123', получен '%s'", matches[1])
	}
}

func TestNotFound(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	knownCalled := false
	notFoundCalled := false

	r.HandleCommand("/start", func(ctx Context) error {
		knownCalled = true
		return nil
	})
	r.NotFound(func(ctx Context) error {
		notFoundCalled = true
		return nil
	})

	_ = r.HandleCtx(context.Background(), messageUpdate("/unknown"))

	if knownCalled {
		t.Error("хэндлер /start не должен был вызваться")
	}
	if !notFoundCalled {
		t.Error("NotFound-хэндлер должен был вызваться")
	}
}

func TestCallbackMatch(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	buyCalled := false
	sellCalled := false
	gotPayload := ""

	r.HandleCallback("action:buy", func(ctx Context) error {
		buyCalled = true
		gotPayload = ctx.Data()
		return nil
	})
	r.HandleCallback("action:sell", func(ctx Context) error {
		sellCalled = true
		return nil
	})

	_ = r.HandleCtx(context.Background(), model.Update{
		UpdateType: model.UpdateMessageCallback,
		Callback:   &model.Callback{Payload: "action:buy"},
	})

	if !buyCalled {
		t.Error("callback-хэндлер 'action:buy' должен был вызваться")
	}
	if sellCalled {
		t.Error("callback-хэндлер 'action:sell' не должен был вызваться")
	}
	if gotPayload != "action:buy" {
		t.Errorf("ожидался payload 'action:buy', получен '%s'", gotPayload)
	}
}
