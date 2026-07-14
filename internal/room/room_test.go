package room_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
	"github.com/BRO-CODES-HERE/OpenChat/internal/room"
)

type mockConn struct {
	io.Reader
	io.Writer
}

func (m *mockConn) Close() error { return nil }

func TestRoomHost_AddClient(t *testing.T) {
	hub := chat.NewHub()
	host := room.NewHost("test-room", hub)

	// Subscribe to local hub to verify updates
	sub := hub.Subscribe()

	inR, inW := io.Pipe()
	conn := &mockConn{Reader: inR, Writer: new(bytes.Buffer)}

	// Add client
	host.AddClient("bob", conn)
	defer inW.Close()

	// The client joining should trigger updateAndBroadcastCount:
	// A message with sender "system:count" and content "2" (host + bob).
	// We read it from the local hub subscription.
	select {
	case msg := <-sub:
		if msg.Sender != "system:count" || msg.Content != "2" {
			t.Fatalf("expected system:count message with content '2', got %+v", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for peer count update")
	}

	// Verify client count is 1
	if host.ClientCount() != 1 {
		t.Fatalf("expected client count to be 1, got %d", host.ClientCount())
	}
}
