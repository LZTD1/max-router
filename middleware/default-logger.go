package middleware

import (
	"github.com/LZTD1/max-router/v2"
	"log"
	"time"
)

func DefaultLogger() maxrouter.Middleware {
	return func(next maxrouter.RouteHandler) maxrouter.RouteHandler {
		return maxrouter.HandlerFunc(func(c maxrouter.Context) error {
			start := time.Now()

			log.Printf("[maxrouter] Started %s | UserID: %d | ChatID: %d",
				c.Update().UpdateType, c.UserID(), c.ChatID())

			err := next.ServeContext(c)

			log.Printf("[maxrouter] Completed %s in %v | Error: %v",
				c.Update().UpdateType, time.Since(start), err)

			return err
		})
	}
}
