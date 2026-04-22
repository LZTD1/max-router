package maxrouter

import (
	"context"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func (r *Router) Handle(upd schemes.UpdateInterface) {
	if r.isAsync {
		go func() {
			_ = r.HandleCtx(context.Background(), upd)
		}()
	} else {
		_ = r.HandleCtx(context.Background(), upd)
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
