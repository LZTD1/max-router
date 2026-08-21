package maxrouter

import (
	"context"
	"errors"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

type Context interface {
	Update() model.Update
	API() *maxbot.Api
	Ctx() context.Context

	Message() *model.MessageUpdate
	Callback() *model.Callback
	Text() string
	Command() string
	Data() string
	ChatID() int64
	UserID() int64

	Set(key string, val any)
	Get(key string) (any, bool)

	Send(text string, opts ...Option) error
	Reply(text string, opts ...Option) error
	Edit(text string, opts ...Option) error
	Answer(notification string) error

	User() *model.User
	FullName() string
	Username() string
	FirstName() string
	LastName() string

	Handled() bool
}

type maxContext struct {
	update  model.Update
	api     *maxbot.Api
	goCtx   context.Context
	handled bool
	store   map[string]any
}

func newContext(goCtx context.Context, api *maxbot.Api, upd model.Update) *maxContext {
	return &maxContext{
		update: upd,
		api:    api,
		goCtx:  goCtx,
	}
}

func (c *maxContext) Update() model.Update { return c.update }
func (c *maxContext) API() *maxbot.Api     { return c.api }
func (c *maxContext) Ctx() context.Context { return c.goCtx }
func (c *maxContext) Handled() bool        { return c.handled }

func (c *maxContext) Set(key string, val any) {
	if c.store == nil {
		c.store = make(map[string]any)
	}
	c.store[key] = val
}

func (c *maxContext) Get(key string) (any, bool) {
	if c.store == nil {
		return nil, false
	}
	val, ok := c.store[key]
	return val, ok
}

func (c *maxContext) Message() *model.MessageUpdate {
	message := c.update.Message
	if message == nil {
		return nil
	}
	// The v2 client creates an empty MessageUpdate for callbacks without an
	// original message. Preserve the v1 router's "no message" behavior.
	if c.update.UpdateType == model.UpdateMessageCallback && message.Body.Mid == "" {
		return nil
	}
	return message
}

func (c *maxContext) Callback() *model.Callback { return c.update.Callback }

func (c *maxContext) Text() string {
	if m := c.Message(); m != nil {
		return m.Body.Text
	}
	return ""
}

func (c *maxContext) Command() string {
	if c.update.UpdateType == model.UpdateMessageCreated {
		return c.update.GetCommand().Command
	}
	return ""
}

func (c *maxContext) Data() string {
	if cb := c.Callback(); cb != nil {
		return cb.Payload
	}
	return ""
}

func (c *maxContext) ChatID() int64 { return c.update.ChatID }

func (c *maxContext) UserID() int64 {
	if callback := c.Callback(); callback != nil {
		return callback.User.UserID
	}
	return c.update.UserID
}

func (c *maxContext) Send(text string, opts ...Option) error {
	c.handled = true
	if c.api == nil {
		return ErrNilAPI
	}
	chatID := c.ChatID()
	if chatID == 0 {
		return errors.New("maxrouter: no chat id in update")
	}
	o := buildOptions(opts)

	msg := maxbot.NewMessage().SetChat(chatID).SetText(text)
	if o.Keyboard != nil {
		msg = msg.AddKeyboard(o.Keyboard)
	}
	if o.ReplyToID != "" {
		msg = msg.SetReply(text, o.ReplyToID)
	}
	if o.Notify != nil && !*o.Notify {
		msg = msg.WithoutNotify()
	}
	if o.Format != "" {
		msg = msg.SetFormat(o.Format)
	}
	_, err := c.api.Messages.Send(c.goCtx, msg)
	return err
}

func (c *maxContext) Reply(text string, opts ...Option) error {
	if m := c.Message(); m != nil {
		opts = append(opts, WithReplyTo(m.Body.Mid))
	}
	return c.Send(text, opts...)
}

func (c *maxContext) Edit(text string, opts ...Option) error {
	c.handled = true
	if c.api == nil {
		return ErrNilAPI
	}
	m := c.Message()
	if m == nil || m.Body.Mid == "" {
		return errors.New("maxrouter: no message to edit")
	}
	o := buildOptions(opts)
	req := maxbot.NewMessage().SetText(text)
	if o.Keyboard != nil {
		req = req.AddKeyboard(o.Keyboard)
	}
	if o.Format != "" {
		req = req.SetFormat(o.Format)
	}
	_, err := c.api.Messages.EditMessage(c.goCtx, m.Body.Mid, req.MessageBody())
	return err
}

func (c *maxContext) Answer(notification string) error {
	c.handled = true
	if c.api == nil {
		return ErrNilAPI
	}
	cb := c.Callback()
	if cb == nil {
		return errors.New("maxrouter: no callback in update")
	}
	_, err := c.api.Messages.AnswerOnCallback(c.goCtx, cb.CallbackID,
		model.CallbackAnswer{Notification: &notification})
	return err
}

func (c *maxContext) User() *model.User {
	if callback := c.Callback(); callback != nil {
		return &callback.User
	}
	if c.update.User != nil {
		return c.update.User
	}
	if message := c.Message(); message != nil {
		return &model.User{
			UserID:    message.Sender.UserID,
			FirstName: message.Sender.FirstName,
			LastName:  message.Sender.LastName,
			Username:  message.Sender.Username,
			IsBot:     message.Sender.IsBot,
			Name:      message.Sender.Name,
		}
	}
	return nil
}

func (c *maxContext) FullName() string {
	if user := c.User(); user != nil {
		return user.Name
	}
	return ""
}

func (c *maxContext) Username() string {
	if user := c.User(); user != nil {
		return user.Username
	}
	return ""
}

func (c *maxContext) FirstName() string {
	if user := c.User(); user != nil {
		return user.FirstName
	}
	return ""
}

func (c *maxContext) LastName() string {
	if user := c.User(); user != nil {
		return user.LastName
	}
	return ""
}
