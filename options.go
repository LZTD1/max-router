package maxrouter

import (
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

type SendOptions struct {
	Keyboard    *maxbot.Keyboard
	Attachments []any
	ReplyToID   string
	Notify      bool
	Format      schemes.Format
}

type Option func(*SendOptions)

func WithKeyboard(kb *maxbot.Keyboard) Option {
	return func(o *SendOptions) { o.Keyboard = kb }
}
func WithReplyTo(messageID string) Option {
	return func(o *SendOptions) { o.ReplyToID = messageID }
}

func WithNotify(notify bool) Option {
	return func(o *SendOptions) { o.Notify = notify }
}

func WithFormat(format schemes.Format) Option {
	return func(o *SendOptions) { o.Format = format }
}

func WithAttachments(a ...any) Option {
	return func(o *SendOptions) { o.Attachments = append(o.Attachments, a...) }
}

func buildOptions(opts []Option) SendOptions {
	var out SendOptions
	for _, o := range opts {
		o(&out)
	}
	return out
}
