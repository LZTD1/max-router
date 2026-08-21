package maxrouter

import (
	"regexp"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

type regexEntry struct {
	re *regexp.Regexp
	h  RouteHandler
}

type routeStore struct {
	exactCommands  map[string]RouteHandler
	exactTexts     map[string]RouteHandler
	exactCallbacks map[string]RouteHandler

	regexTexts     []regexEntry
	regexCallbacks []regexEntry

	typedHandlers map[model.UpdateType]RouteHandler

	notFound RouteHandler
}

func newRouteStore() *routeStore {
	return &routeStore{
		exactCommands:  make(map[string]RouteHandler),
		exactTexts:     make(map[string]RouteHandler),
		exactCallbacks: make(map[string]RouteHandler),
		typedHandlers:  make(map[model.UpdateType]RouteHandler),
	}
}

func (s *routeStore) resolve(ctx *maxContext) RouteHandler {
	switch ctx.update.UpdateType {
	case model.UpdateMessageCreated:
		if cmd := ctx.Command(); cmd != "" {
			if h, ok := s.exactCommands[cmd]; ok {
				return h
			}
		}
		text := ctx.Text()
		if h, ok := s.exactTexts[text]; ok {
			return h
		}
		for _, e := range s.regexTexts {
			if e.re.MatchString(text) {
				return e.h
			}
		}

	case model.UpdateMessageCallback:
		payload := ctx.Data()
		if h, ok := s.exactCallbacks[payload]; ok {
			return h
		}
		for _, e := range s.regexCallbacks {
			if e.re.MatchString(payload) {
				return e.h
			}
		}
	}

	if h, ok := s.typedHandlers[ctx.update.UpdateType]; ok {
		return h
	}

	return s.notFound
}
