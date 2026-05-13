package maxrouter

import (
	"context"
	"errors"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

type Context interface {
	Update() schemes.UpdateInterface
	API() *maxbot.Api
	Ctx() context.Context

	Message() *schemes.Message
	Callback() *schemes.Callback
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

	User() *schemes.User
	FullName() string
	Username() string
	FirstName() string
	LastName() string

	Handled() bool
}

type maxContext struct {
	update  schemes.UpdateInterface
	api     *maxbot.Api
	goCtx   context.Context
	handled bool
	store   map[string]any
}

func newContext(goCtx context.Context, api *maxbot.Api, upd schemes.UpdateInterface) *maxContext {
	return &maxContext{
		update: upd,
		api:    api,
		goCtx:  goCtx,
	}
}

func (c *maxContext) Update() schemes.UpdateInterface { return c.update }
func (c *maxContext) API() *maxbot.Api                { return c.api }
func (c *maxContext) Ctx() context.Context            { return c.goCtx }
func (c *maxContext) Handled() bool                   { return c.handled }

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

func (c *maxContext) Message() *schemes.Message {
	switch u := c.update.(type) {
	case *schemes.MessageCreatedUpdate:
		return &u.Message
	case *schemes.MessageEditedUpdate:
		return &u.Message
	case *schemes.MessageCallbackUpdate:
		if u.Message != nil {
			return u.Message
		}
	}
	return nil
}

func (c *maxContext) Callback() *schemes.Callback {
	if u, ok := c.update.(*schemes.MessageCallbackUpdate); ok {
		return &u.Callback
	}
	return nil
}

func (c *maxContext) Text() string {
	if m := c.Message(); m != nil {
		return m.Body.Text
	}
	return ""
}

func (c *maxContext) Command() string {
	if u, ok := c.update.(*schemes.MessageCreatedUpdate); ok {
		return u.GetCommand()
	}
	return ""
}

func (c *maxContext) Data() string {
	if cb := c.Callback(); cb != nil {
		return cb.Payload
	}
	return ""
}

func (c *maxContext) ChatID() int64 { return c.update.GetChatID() }
func (c *maxContext) UserID() int64 { return c.update.GetUserID() }

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
	if o.Notify {
		msg = msg.SetNotify(true)
	}
	if o.Format != "" {
		msg = msg.SetFormat(o.Format)
	}
	return c.api.Messages.Send(c.goCtx, msg)
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
	if m == nil {
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
	return c.api.Messages.EditMessage(c.goCtx, m.Body.Mid, req)
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
		&schemes.CallbackAnswer{Notification: notification})
	return err
}

func (c *maxContext) User() *schemes.User {
	switch u := c.update.(type) {
	case *schemes.MessageCreatedUpdate:
		return &u.Message.Sender
	case *schemes.MessageEditedUpdate:
		return &u.Message.Sender
	case *schemes.MessageCallbackUpdate:
		return &u.Callback.User
	case *schemes.BotStartedUpdate:
		return &u.User
	case *schemes.UserAddedToChatUpdate:
		return &u.User
	case *schemes.ChatTitleChangedUpdate:
		return &u.User
	case *schemes.BotAddedToChatUpdate:
		return &u.User
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
