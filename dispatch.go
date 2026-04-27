package maxrouter

import (
	"context"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func (r *Router) Handle(upd schemes.UpdateInterface, ctx context.Context) {
	if r.isAsync {
		go func() {
			_ = r.HandleCtx(ctx, upd)
		}()
	} else {
		_ = r.HandleCtx(ctx, upd)
	}
}

func (r *Router) HandleCtx(ctx context.Context, upd schemes.UpdateInterface) error {
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
