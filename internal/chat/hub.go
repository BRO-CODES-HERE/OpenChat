package chat

import (
	"sync"
	"time"
)

// Message represents a single chat line.
type Message struct {
	Sender    string
	Content   string
	Timestamp time.Time
}

// Hub routes messages between the TUI and SSH sessions.
type Hub struct {
	mu       sync.RWMutex
	messages []Message
	subs     []chan Message
	onSend   []func(Message)
}

// NewHub creates a message hub.
func NewHub() *Hub {
	return &Hub{}
}

// OnSend registers a callback invoked when the local user sends a message.
func (h *Hub) OnSend(fn func(Message)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onSend = append(h.onSend, fn)
}

// Subscribe returns a channel that receives new messages.
func (h *Hub) Subscribe() <-chan Message {
	ch := make(chan Message, 64)
	h.mu.Lock()
	h.subs = append(h.subs, ch)
	h.mu.Unlock()
	return ch
}

// Messages returns a copy of the message history.
func (h *Hub) Messages() []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Message, len(h.messages))
	copy(out, h.messages)
	return out
}

// Publish adds an incoming message and notifies subscribers.
func (h *Hub) Publish(msg Message) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	h.mu.Lock()
	h.messages = append(h.messages, msg)
	subs := append([]chan Message(nil), h.subs...)
	h.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- msg:
		default:
		}
	}
}

// Send publishes a message from the local user and triggers the send callback.
func (h *Hub) Send(sender, content string) {
	msg := Message{
		Sender:    sender,
		Content:   content,
		Timestamp: time.Now(),
	}
	h.Publish(msg)

	h.mu.RLock()
	fns := append([]func(Message){}, h.onSend...)
	h.mu.RUnlock()
	for _, fn := range fns {
		fn(msg)
	}
}
