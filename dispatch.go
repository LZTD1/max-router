package maxrouter

import (
	"context"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

func (r *Router) Handle(upd model.Update, ctx context.Context) {
	if r.isAsync {
		go func() {
			_ = r.HandleCtx(ctx, upd)
		}()
	} else {
		_ = r.HandleCtx(ctx, upd)
	}
}

// UpdateHandler adapts the router to maxbot.Api.GetHandler.
func (r *Router) UpdateHandler(ctx context.Context, upd model.Update) {
	// The HTTP request context is canceled after this function returns, so
	// webhook handlers must finish before acknowledging the request.
	_ = r.HandleCtx(ctx, upd)
}

func (r *Router) HandleCtx(ctx context.Context, upd model.Update) error {
	c := newContext(ctx, r.api, upd)

	h := r.store.resolve(c)
	if h != nil {
		err := h.ServeContext(c)
		if err != nil && r.errorHandler != nil {
			r.errorHandler(err, c)
		}
		return err
	}

	return nil
}
