package maxrouter

import (
	"context"
	"errors"
	"testing"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

func TestContextData(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))

	var gotCommand, gotText string

	r.HandleCommand("/help", func(ctx Context) error {
		gotCommand = ctx.Command()
		gotText = ctx.Text()
		return nil
	})

	_ = r.HandleCtx(context.Background(), messageUpdate("/help"))

	if gotCommand != "/help" {
		t.Errorf("Command(): ожидалось '/help', получено '%s'", gotCommand)
	}
	if gotText != "/help" {
		t.Errorf("Text(): ожидалось '/help', получено '%s'", gotText)
	}
}

func TestContextLocals(t *testing.T) {
	t.Run("set and get", func(t *testing.T) {
		r := NewRouter(nil, WithAsync(false))

		r.HandleCommand("/start", func(ctx Context) error {
			ctx.Set("key", "val")

			v, ok := ctx.Get("key")
			if !ok {
				t.Error("Get: ключ должен существовать после Set")
			}
			if v != "val" {
				t.Errorf("Get: ожидалось 'val', получено '%v'", v)
			}
			return nil
		})

		_ = r.HandleCtx(context.Background(), messageUpdate("/start"))
	})

	t.Run("get missing key", func(t *testing.T) {
		r := NewRouter(nil, WithAsync(false))

		r.HandleCommand("/start", func(ctx Context) error {
			_, ok := ctx.Get("ghost")
			if ok {
				t.Error("Get: несуществующий ключ не должен возвращать ok=true")
			}
			return nil
		})

		_ = r.HandleCtx(context.Background(), messageUpdate("/start"))
	})
}

func TestContextCallbackUsesCallbackUser(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))
	var userID int64

	r.HandleCallback("action", func(ctx Context) error {
		userID = ctx.UserID()
		return nil
	})

	_ = r.HandleCtx(context.Background(), model.Update{
		UpdateType: model.UpdateMessageCallback,
		UserID:     10,
		Callback: &model.Callback{
			Payload: "action",
			User:    model.User{UserID: 20},
		},
	})

	if userID != 20 {
		t.Errorf("UserID(): ожидался callback user ID 20, получен %d", userID)
	}
}

func TestContextCommandWithParameters(t *testing.T) {
	r := NewRouter(nil, WithAsync(false))
	called := false

	r.HandleCommand("/help", func(Context) error {
		called = true
		return nil
	})

	_ = r.HandleCtx(context.Background(), messageUpdate("/help:topic more text"))

	if !called {
		t.Error("маршрут команды должен использовать имя команды без параметров")
	}
}

func TestContextCallbackWithoutMessage(t *testing.T) {
	c := newContext(context.Background(), nil, model.Update{
		UpdateType: model.UpdateMessageCallback,
		Message:    &model.MessageUpdate{},
	})

	if c.Message() != nil {
		t.Error("callback without a message must not expose an empty message")
	}
	if err := c.Edit("text"); err == nil {
		t.Error("Edit must reject a callback without a message")
	}
}

func TestContextOutboundActions(t *testing.T) {
	messages := &mockMessages{}
	api := &maxbot.Api{Messages: messages}
	update := model.Update{
		ChatID: 1,
		Message: &model.MessageUpdate{
			Body: model.MessageBody{Mid: "message-id"},
		},
	}
	c := newContext(context.Background(), api, update)

	keyboard := model.NewKeyboard()
	keyboard.AddRow().AddCallBack("action", "payload")
	if err := c.Send("send", WithKeyboard(keyboard), WithNotify(false)); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !messages.sent {
		t.Error("Send must call Messages.Send")
	}
	if messages.sendBody.Notify == nil || *messages.sendBody.Notify {
		t.Error("WithNotify(false) must disable notifications")
	}
	if len(messages.sendBody.Attachments) != 1 {
		t.Errorf("Send keyboard attachments: expected 1, got %d", len(messages.sendBody.Attachments))
	}

	if err := c.Reply("reply"); err != nil {
		t.Fatalf("Reply: %v", err)
	}
	if !messages.sent {
		t.Error("Reply must call Messages.Send")
	}
	if messages.sendBody.Link == nil || messages.sendBody.Link.Mid != "message-id" {
		t.Error("Reply must use the source message ID")
	}

	if err := c.Edit("edit", WithFormat(model.FormatHTML)); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if messages.editMessageID != "message-id" {
		t.Errorf("Edit message ID: expected message-id, got %q", messages.editMessageID)
	}
	if messages.editBody.Format != model.FormatHTML {
		t.Errorf("Edit format: expected HTML, got %q", messages.editBody.Format)
	}

	c = newContext(context.Background(), api, model.Update{
		Callback: &model.Callback{CallbackID: "callback-id"},
	})
	if err := c.Answer("done"); err != nil {
		t.Fatalf("Answer: %v", err)
	}
	if messages.callbackID != "callback-id" {
		t.Errorf("Answer callback ID: expected callback-id, got %q", messages.callbackID)
	}
	if messages.answer.Notification == nil || *messages.answer.Notification != "done" {
		t.Error("Answer must send the notification text")
	}
}

func TestUpdateHandlerWaitsForRoute(t *testing.T) {
	r := NewRouter(nil)
	completed := false
	r.HandleCommand("/start", func(Context) error {
		completed = true
		return nil
	})

	r.UpdateHandler(context.Background(), messageUpdate("/start"))

	if !completed {
		t.Error("UpdateHandler must finish routing before returning")
	}
}

type mockMessages struct {
	sent          bool
	sendBody      model.NewMessageBody
	editMessageID string
	editBody      model.NewMessageBody
	callbackID    string
	answer        model.CallbackAnswer
}

func (m *mockMessages) GetMessages(context.Context, int64, int64, int64, int64, []string) (model.MessageList, error) {
	return model.MessageList{}, errors.New("not implemented")
}

func (m *mockMessages) GetMessageByID(context.Context, string) (model.Message, error) {
	return model.Message{}, errors.New("not implemented")
}

func (m *mockMessages) Send(_ context.Context, message *maxbot.Message) (model.SendMessageResult, error) {
	m.sent = true
	m.sendBody = message.MessageBody()
	return model.SendMessageResult{}, nil
}

func (m *mockMessages) EditMessage(_ context.Context, messageID string, body model.NewMessageBody) (model.SimpleQueryResult, error) {
	m.editMessageID = messageID
	m.editBody = body
	return model.SimpleQueryResult{}, nil
}

func (m *mockMessages) DeleteMessage(context.Context, string) (model.SimpleQueryResult, error) {
	return model.SimpleQueryResult{}, errors.New("not implemented")
}

func (m *mockMessages) AnswerOnCallback(_ context.Context, callbackID string, answer model.CallbackAnswer) (model.SimpleQueryResult, error) {
	m.callbackID = callbackID
	m.answer = answer
	return model.SimpleQueryResult{}, nil
}

func (m *mockMessages) GetVideoAttachmentDetails(context.Context, string) (model.VideoAttachmentDetails, error) {
	return model.VideoAttachmentDetails{}, errors.New("not implemented")
}

func messageUpdate(text string) model.Update {
	return model.Update{
		UpdateType: model.UpdateMessageCreated,
		Message: &model.MessageUpdate{
			Body: model.MessageBody{Text: text},
		},
	}
}
