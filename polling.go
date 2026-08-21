package maxrouter

import (
	"context"
	"time"
)

const pollingRetryDelay = time.Second

type pollingOptions struct {
	errorHandler func(error)
}

type PollingOption func(*pollingOptions)

// WithPollingErrorHandler receives errors returned while fetching updates.
// It is separate from Router.OnError, which receives route handler errors.
func WithPollingErrorHandler(handler func(error)) PollingOption {
	return func(options *pollingOptions) {
		options.errorHandler = handler
	}
}

func buildPollingOptions(options []PollingOption) pollingOptions {
	var result pollingOptions
	for _, option := range options {
		if option != nil {
			option(&result)
		}
	}
	return result
}

// RunPolling receives updates until ctx is canceled. It keeps the MAX API
// marker internally and dispatches each update through Router.Handle.
func (r *Router) RunPolling(ctx context.Context, options ...PollingOption) error {
	if r.api == nil {
		return ErrNilAPI
	}
	if r.api.Subscriptions == nil {
		return ErrNilSubscriptions
	}

	config := buildPollingOptions(options)
	var marker int64

	for {
		updates, nextMarker, err := r.api.Subscriptions.GetUpdates(ctx, marker)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if config.errorHandler != nil {
				config.errorHandler(err)
			}
			if err := waitPollingRetry(ctx); err != nil {
				return err
			}
			continue
		}

		marker = nextMarker
		for _, update := range updates {
			r.Handle(update, ctx)
		}
	}
}

func waitPollingRetry(ctx context.Context) error {
	timer := time.NewTimer(pollingRetryDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
