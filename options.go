package maxrouter

import (
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

type SendOptions struct {
	Keyboard  *model.Keyboard
	ReplyToID string
	Notify    *bool
	Format    model.TextFormat
}

type Option func(*SendOptions)

func WithKeyboard(kb *model.Keyboard) Option {
	return func(o *SendOptions) { o.Keyboard = kb }
}
func WithReplyTo(messageID string) Option {
	return func(o *SendOptions) { o.ReplyToID = messageID }
}

func WithNotify(notify bool) Option {
	return func(o *SendOptions) { o.Notify = &notify }
}

func WithFormat(format model.TextFormat) Option {
	return func(o *SendOptions) { o.Format = format }
}

func buildOptions(opts []Option) SendOptions {
	var out SendOptions
	for _, o := range opts {
		o(&out)
	}
	return out
}
