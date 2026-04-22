package maxrouter

import (
	"regexp"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
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

	typedHandlers map[schemes.UpdateType]RouteHandler

	notFound RouteHandler
}

func newRouteStore() *routeStore {
	return &routeStore{
		exactCommands:  make(map[string]RouteHandler),
		exactTexts:     make(map[string]RouteHandler),
		exactCallbacks: make(map[string]RouteHandler),
		typedHandlers:  make(map[schemes.UpdateType]RouteHandler),
	}
}

func (s *routeStore) resolve(ctx *maxContext) RouteHandler {
	switch u := ctx.update.(type) {
	case *schemes.MessageCreatedUpdate:
		if cmd := u.GetCommand(); cmd != "" && cmd != schemes.CommandUndefined {
			if h, ok := s.exactCommands[cmd]; ok {
				return h
			}
		}
		text := u.Message.Body.Text
		if h, ok := s.exactTexts[text]; ok {
			return h
		}
		for _, e := range s.regexTexts {
			if e.re.MatchString(text) {
				return e.h
			}
		}

	case *schemes.MessageCallbackUpdate:
		payload := u.Callback.Payload
		if h, ok := s.exactCallbacks[payload]; ok {
			return h
		}
		for _, e := range s.regexCallbacks {
			if e.re.MatchString(payload) {
				return e.h
			}
		}
	}

	if h, ok := s.typedHandlers[ctx.update.GetUpdateType()]; ok {
		return h
	}

	return s.notFound
}
