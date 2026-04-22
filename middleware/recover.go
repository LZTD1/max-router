package middleware

import (
	"fmt"
	"github.com/LZTD1/max-router"
	"log"
)

func Recover() maxrouter.Middleware {
	return func(next maxrouter.RouteHandler) maxrouter.RouteHandler {
		return maxrouter.HandlerFunc(func(c maxrouter.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PANIC RECOVERED] %v", r)
					err = fmt.Errorf("panic: %v", r)
				}
			}()

			return next.ServeContext(c)
		})
	}
}
