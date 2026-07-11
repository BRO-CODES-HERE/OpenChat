package chat_test

import (
	"testing"

	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
)

func TestHubSendAndPublish(t *testing.T) {
	h := chat.NewHub()
	ch := h.Subscribe()

	var sent bool
	h.OnSend(func(msg chat.Message) {
		sent = true
		if msg.Content != "hello" {
			t.Fatalf("unexpected content: %s", msg.Content)
		}
	})

	h.Send("alice", "hello")

	msg := <-ch
	if msg.Sender != "alice" || msg.Content != "hello" {
		t.Fatalf("unexpected message: %+v", msg)
	}
	if !sent {
		t.Fatal("onSend callback not invoked")
	}
}

func TestHubMultipleCallbacks(t *testing.T) {
	h := chat.NewHub()
	var count int
	h.OnSend(func(chat.Message) { count++ })
	h.OnSend(func(chat.Message) { count++ })
	h.Send("bob", "test")
	if count != 2 {
		t.Fatalf("expected 2 callbacks, got %d", count)
	}
}
