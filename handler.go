package maxrouter

type RouteHandler interface {
	ServeContext(ctx Context) error
}

type HandlerFunc func(ctx Context) error

func (f HandlerFunc) ServeContext(ctx Context) error { return f(ctx) }

type Middleware func(next RouteHandler) RouteHandler

func chain(h RouteHandler, mws ...Middleware) RouteHandler {
	if len(mws) == 0 {
		return h
	}
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
