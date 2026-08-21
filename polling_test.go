package maxrouter

import (
	"context"
	"errors"
	"testing"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

func TestRunPollingDispatchesUpdates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	polling := &mockSubscriptions{
		responses: []pollingResponse{{updates: []model.Update{messageUpdate("/start")}, marker: 10}},
	}
	router := NewRouter(&maxbot.Api{Subscriptions: polling}, WithAsync(false))
	called := false
	router.HandleCommand("/start", func(Context) error {
		called = true
		cancel()
		return nil
	})

	err := router.RunPolling(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunPolling: expected context.Canceled, got %v", err)
	}
	if !called {
		t.Error("registered handler was not called")
	}
}

func TestRunPollingPassesNextMarker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	polling := &mockSubscriptions{
		responses: []pollingResponse{{marker: 42}, {marker: 43}},
		onCall: func(call int, marker int64) {
			switch call {
			case 1:
				if marker != 0 {
					t.Errorf("first marker: expected 0, got %d", marker)
				}
			case 2:
				if marker != 42 {
					t.Errorf("second marker: expected 42, got %d", marker)
				}
				cancel()
			}
		},
	}

	err := NewRouter(&maxbot.Api{Subscriptions: polling}, WithAsync(false)).RunPolling(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunPolling: expected context.Canceled, got %v", err)
	}
}

func TestRunPollingCallsErrorHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pollingError := errors.New("polling failed")
	polling := &mockSubscriptions{responses: []pollingResponse{{err: pollingError}}}
	var received error

	err := NewRouter(&maxbot.Api{Subscriptions: polling}).RunPolling(
		ctx,
		WithPollingErrorHandler(func(err error) {
			received = err
			cancel()
		}),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunPolling: expected context.Canceled, got %v", err)
	}
	if !errors.Is(received, pollingError) {
		t.Fatalf("error handler: expected %v, got %v", pollingError, received)
	}
}

func TestRunPollingRejectsNilDependencies(t *testing.T) {
	if err := NewRouter(nil).RunPolling(context.Background()); !errors.Is(err, ErrNilAPI) {
		t.Fatalf("nil API: expected ErrNilAPI, got %v", err)
	}
	if err := NewRouter(&maxbot.Api{}).RunPolling(context.Background()); !errors.Is(err, ErrNilSubscriptions) {
		t.Fatalf("nil subscriptions: expected ErrNilSubscriptions, got %v", err)
	}
}

type pollingResponse struct {
	updates []model.Update
	marker  int64
	err     error
}

type mockSubscriptions struct {
	responses []pollingResponse
	calls     int
	onCall    func(call int, marker int64)
}

func (m *mockSubscriptions) GetUpdates(_ context.Context, marker int64) ([]model.Update, int64, error) {
	m.calls++
	if m.onCall != nil {
		m.onCall(m.calls, marker)
	}
	index := m.calls - 1
	if index >= len(m.responses) {
		return nil, marker, context.Canceled
	}
	response := m.responses[index]
	return response.updates, response.marker, response.err
}

func (m *mockSubscriptions) GetSubscriptions(context.Context) (model.GetSubscriptionsResult, error) {
	return model.GetSubscriptionsResult{}, errors.New("not implemented")
}

func (m *mockSubscriptions) Subscribe(context.Context, string, string, []string, string) (model.SimpleQueryResult, error) {
	return model.SimpleQueryResult{}, errors.New("not implemented")
}

func (m *mockSubscriptions) Unsubscribe(context.Context, string) (model.SimpleQueryResult, error) {
	return model.SimpleQueryResult{}, errors.New("not implemented")
}
