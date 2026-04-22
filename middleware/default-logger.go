package middleware

import (
	"github.com/LZTD1/max-router"
	"log"
	"time"
)

func DefaultLogger() maxrouter.Middleware {
	return func(next maxrouter.RouteHandler) maxrouter.RouteHandler {
		return maxrouter.HandlerFunc(func(c maxrouter.Context) error {
			start := time.Now()

			log.Printf("[maxrouter] Started %s | UserID: %d | ChatID: %d",
				c.Update().GetUpdateType(), c.UserID(), c.ChatID())

			err := next.ServeContext(c)

			log.Printf("[maxrouter] Completed %s in %v | Error: %v",
				c.Update().GetUpdateType(), time.Since(start), err)

			return err
		})
	}
}
