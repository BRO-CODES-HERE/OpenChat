package tui_test

import (
	"testing"
	"time"

	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
	"github.com/BRO-CODES-HERE/OpenChat/internal/tui"
)

func TestFormatParseMessage(t *testing.T) {
	orig := chat.Message{
		Sender:    "alice",
		Content:   "hello|world",
		Timestamp: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}
	line := tui.FormatMessage(orig)
	parsed, err := tui.ParseMessage(line)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Sender != orig.Sender || parsed.Content != orig.Content {
		t.Fatalf("roundtrip failed: %+v", parsed)
	}
}
