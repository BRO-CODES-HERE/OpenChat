package storage_test

import (
	"testing"

	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
	"github.com/BRO-CODES-HERE/OpenChat/internal/storage"
)

func TestGhostMode(t *testing.T) {
	store, err := storage.Open(storage.ModeGhost, "")
	if err != nil {
		t.Fatal(err)
	}
	msg := chat.Message{Sender: "alice", Content: "secret"}
	if err := store.Save(msg); err != nil {
		t.Fatal(err)
	}
	msgs, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0].Content != "secret" {
		t.Fatalf("unexpected messages: %+v", msgs)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestLocalEncryptedStorage(t *testing.T) {
	store, err := storage.Open(storage.ModeLocal, "test-passphrase")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	msg := chat.Message{Sender: "bob", Content: "encrypted"}
	if err := store.Save(msg); err != nil {
		t.Fatal(err)
	}
	msgs, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0].Content != "encrypted" {
		t.Fatalf("unexpected messages: %+v", msgs)
	}
}
