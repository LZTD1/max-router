package maxrouter

import (
	"log"
	"regexp"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

type ErrorHandler func(err error, ctx Context)

type Router struct {
	api          *maxbot.Api
	store        *routeStore
	middlewares  []Middleware
	errorHandler ErrorHandler
	isAsync      bool
}

type RouterOption func(*Router)

func WithAsync(async bool) RouterOption {
	return func(r *Router) { r.isAsync = async }
}

func NewRouter(api *maxbot.Api, opts ...RouterOption) *Router {
	r := &Router{
		api:          api,
		store:        newRouteStore(),
		errorHandler: defaultErrorHandler,
		isAsync:      true,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func defaultErrorHandler(err error, ctx Context) {
	log.Printf("[maxrouter] Handler error (Update: %s): %v", ctx.Update().GetUpdateType(), err)
}

func (r *Router) OnError(eh ErrorHandler) {
	r.errorHandler = eh
}

func (r *Router) wrap(h RouteHandler) RouteHandler {
	return chain(h, r.middlewares...)
}

func (r *Router) Use(mws ...Middleware) {
	r.middlewares = append(r.middlewares, mws...)
}

func (r *Router) Group(fn func(sub *Router)) {
	sub := &Router{
		api:          r.api,
		store:        r.store,
		middlewares:  append([]Middleware(nil), r.middlewares...), // Копия
		errorHandler: r.errorHandler,
	}
	fn(sub)
}

func (r *Router) With(mws ...Middleware) *Router {
	return &Router{
		api:          r.api,
		store:        r.store,
		middlewares:  append(append([]Middleware(nil), r.middlewares...), mws...),
		errorHandler: r.errorHandler,
	}
}

func (r *Router) NotFound(f HandlerFunc) {
	if f == nil {
		panic(ErrNilHandler)
	}
	r.store.notFound = r.wrap(f)
}

func (r *Router) HandleCommand(cmd string, f HandlerFunc) {
	if cmd == "" {
		panic(ErrEmptyString)
	}
	r.store.exactCommands[cmd] = r.wrap(f)
}

func (r *Router) HandleText(text string, f HandlerFunc) {
	if text == "" {
		panic(ErrEmptyString)
	}
	r.store.exactTexts[text] = r.wrap(f)
}

func (r *Router) HandleRegexpText(re *regexp.Regexp, f HandlerFunc) {
	r.store.regexTexts = append(r.store.regexTexts, regexEntry{re: re, h: r.wrap(f)})
}

func (r *Router) HandleCallback(payload string, f HandlerFunc) {
	if payload == "" {
		panic(ErrEmptyString)
	}
	r.store.exactCallbacks[payload] = r.wrap(f)
}

func (r *Router) HandleRegexpCallback(re *regexp.Regexp, f HandlerFunc) {
	r.store.regexCallbacks = append(r.store.regexCallbacks, regexEntry{re: re, h: r.wrap(f)})
}

func (r *Router) HandleType(t schemes.UpdateType, f HandlerFunc) {
	r.store.typedHandlers[t] = r.wrap(f)
}

func (r *Router) HandleStart(f HandlerFunc) {
	r.HandleType(schemes.TypeBotStarted, f)
}
