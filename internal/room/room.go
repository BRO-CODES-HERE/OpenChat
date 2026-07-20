package room

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
	"github.com/BRO-CODES-HERE/OpenChat/internal/tui"
)

// Host manages a star-topology room and broadcasts messages to all peers.
type Host struct {
	mu       sync.RWMutex
	name     string
	hub      *chat.Hub
	clients  map[string]*Client
	HostUser string
}

// Client represents one connected peer in a room.
type Client struct {
	ID   string
	Conn io.ReadWriteCloser
}

// NewHost creates a room host.
func NewHost(name string, hub *chat.Hub, hostUser string) *Host {
	return &Host{
		name:     name,
		hub:      hub,
		clients:  make(map[string]*Client),
		HostUser: hostUser,
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

	h.updateAndBroadcastCount()

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

	h.updateAndBroadcastCount()
}

func (h *Host) updateAndBroadcastCount() {
	count := h.ClientCount() + 1
	countMsg := chat.Message{
		Sender:  "system:count",
		Content: fmt.Sprintf("%d", count),
	}
	h.hub.Publish(countMsg)
	h.Broadcast(countMsg, "")
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

		// Intercept /users
		if msg.Content == "/users" {
			h.mu.RLock()
			var active []string
			if h.HostUser != "" {
				active = append(active, h.HostUser+" (host)")
			} else {
				active = append(active, "host (host)")
			}
			for cid := range h.clients {
				active = append(active, cid)
			}
			h.mu.RUnlock()

			reply := chat.Message{
				Sender:    "system",
				Timestamp: time.Now(),
				Content:   fmt.Sprintf("Active users: %s", strings.Join(active, ", ")),
			}
			line := tui.FormatMessage(reply) + "\n"
			_, _ = io.WriteString(conn, line)
			continue
		}

		// Intercept /dm
		if strings.HasPrefix(msg.Content, "/dm ") {
			parts := strings.SplitN(msg.Content, " ", 3)
			if len(parts) < 3 {
				reply := chat.Message{
					Sender:    "system",
					Timestamp: time.Now(),
					Content:   "Usage: /dm <username> <message>",
				}
				line := tui.FormatMessage(reply) + "\n"
				_, _ = io.WriteString(conn, line)
				continue
			}
			target := parts[1]
			dmMsg := parts[2]

			// Send to target if it exists
			h.mu.RLock()
			var targetConn io.ReadWriteCloser
			if targetClient, ok := h.clients[target]; ok {
				targetConn = targetClient.Conn
			}
			h.mu.RUnlock()

			// Check if target is host
			if target == h.HostUser {
				// Send to host's local TUI hub
				h.hub.Publish(chat.Message{
					Sender:    "system",
					Timestamp: time.Now(),
					Content:   fmt.Sprintf("[DM from %s]: %s", id, dmMsg),
				})

				// Send confirmation back to sender
				reply := chat.Message{
					Sender:    "system",
					Timestamp: time.Now(),
					Content:   fmt.Sprintf("[DM to %s]: %s", target, dmMsg),
				}
				line := tui.FormatMessage(reply) + "\n"
				_, _ = io.WriteString(conn, line)
				continue
			}

			if targetConn != nil {
				// Write to target
				replyToTarget := chat.Message{
					Sender:    "system",
					Timestamp: time.Now(),
					Content:   fmt.Sprintf("[DM from %s]: %s", id, dmMsg),
				}
				lineTarget := tui.FormatMessage(replyToTarget) + "\n"
				_, _ = io.WriteString(targetConn, lineTarget)

				// Write confirmation back to sender
				replyToSender := chat.Message{
					Sender:    "system",
					Timestamp: time.Now(),
					Content:   fmt.Sprintf("[DM to %s]: %s", target, dmMsg),
				}
				lineSender := tui.FormatMessage(replyToSender) + "\n"
				_, _ = io.WriteString(conn, lineSender)
			} else {
				// User not found
				reply := chat.Message{
					Sender:    "system",
					Timestamp: time.Now(),
					Content:   fmt.Sprintf("User '%s' not found.", target),
				}
				line := tui.FormatMessage(reply) + "\n"
				_, _ = io.WriteString(conn, line)
			}
			continue
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
	msg := chat.Message{
		Sender:  "system",
		Content: fmt.Sprintf("hey every body meet '%s' joined chat now", id),
	}
	h.hub.Publish(msg)
	h.Broadcast(msg, "")
}

// AnnounceLeave notifies the room when a peer disconnects.
func (h *Host) AnnounceLeave(id string) {
	msg := chat.Message{
		Sender:  "system",
		Content: fmt.Sprintf("%s left the room", id),
	}
	h.hub.Publish(msg)
	h.Broadcast(msg, "")
}

// ClientIDs returns a list of active client usernames.
func (h *Host) ClientIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var ids []string
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}

// GetClientConn retrieves the connection writer for a specific client.
func (h *Host) GetClientConn(id string) io.ReadWriteCloser {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if c, ok := h.clients[id]; ok {
		return c.Conn
	}
	return nil
}
