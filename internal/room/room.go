package room

import (
	"bufio"
	"fmt"
	"io"
	"sync"

	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
	"github.com/BRO-CODES-HERE/OpenChat/internal/tui"
)

// Host manages a star-topology room and broadcasts messages to all peers.
type Host struct {
	mu      sync.RWMutex
	name    string
	hub     *chat.Hub
	clients map[string]*Client
}

// Client represents one connected peer in a room.
type Client struct {
	ID   string
	Conn io.ReadWriteCloser
}

// NewHost creates a room host.
func NewHost(name string, hub *chat.Hub) *Host {
	return &Host{
		name:    name,
		hub:     hub,
		clients: make(map[string]*Client),
	}
}

// Name returns the room name.
func (h *Host) Name() string {
	return h.name
}

// AddClient registers a peer connection for broadcast.
func (h *Host) AddClient(id string, conn io.ReadWriteCloser) {
	h.mu.Lock()
	h.clients[id] = &Client{ID: id, Conn: conn}
	h.mu.Unlock()

	go h.readLoop(id, conn)
}

// RemoveClient disconnects a peer.
func (h *Host) RemoveClient(id string) {
	h.mu.Lock()
	if c, ok := h.clients[id]; ok {
		_ = c.Conn.Close()
		delete(h.clients, id)
	}
	h.mu.Unlock()
}

func (h *Host) readLoop(id string, conn io.ReadWriteCloser) {
	defer h.RemoveClient(id)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg, err := tui.ParseMessage(scanner.Text())
		if err != nil {
			continue
		}
		if msg.Sender == "" {
			msg.Sender = id
		}
		h.hub.Publish(msg)
		h.Broadcast(msg, id)
	}
}

// Broadcast sends a message to all clients except optionally the sender.
func (h *Host) Broadcast(msg chat.Message, except string) {
	line := tui.FormatMessage(msg) + "\n"
	h.mu.RLock()
	defer h.mu.RUnlock()
	for id, c := range h.clients {
		if id == except {
			continue
		}
		_, _ = io.WriteString(c.Conn, line)
	}
}

// ClientCount returns the number of connected peers.
func (h *Host) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// AnnounceJoin notifies the room of a new participant.
func (h *Host) AnnounceJoin(id string) {
	h.hub.Publish(chat.Message{
		Sender:  "system",
		Content: fmt.Sprintf("%s joined the room", id),
	})
}

// AnnounceLeave notifies the room when a peer disconnects.
func (h *Host) AnnounceLeave(id string) {
	h.hub.Publish(chat.Message{
		Sender:  "system",
		Content: fmt.Sprintf("%s left the room", id),
	})
}
