package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
	"golang.org/x/crypto/pbkdf2"
)

// Mode selects how messages are persisted.
type Mode int

const (
	ModeGhost Mode = iota
	ModeLocal
)

// Store handles message persistence.
type Store struct {
	mode      Mode
	mu        sync.Mutex
	db        *sql.DB
	gcm       cipher.AEAD
	ghostMsgs []chat.Message
	passphrase string
}

// Open creates a store for the given mode.
func Open(mode Mode, passphrase string) (*Store, error) {
	s := &Store{mode: mode, passphrase: passphrase}
	if mode == ModeGhost {
		return s, nil
	}
	return s, s.initDB(passphrase)
}

func (s *Store) initDB(passphrase string) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	dbPath := filepath.Join(dir, "messages.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	s.db = db

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender TEXT NOT NULL,
		content_enc TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`); err != nil {
		return err
	}

	key := deriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	s.gcm = gcm
	return nil
}

func deriveKey(passphrase string) []byte {
	salt := []byte("chatssh-sqlcipher-v1")
	return pbkdf2.Key([]byte(passphrase), salt, 100_000, 32, sha256.New)
}

func configDir() (string, error) {
	if envDir := os.Getenv("CHATSSH_HOME"); envDir != "" {
		return envDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".chatssh"), nil
}

// Save persists a message according to the active mode.
func (s *Store) Save(msg chat.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch s.mode {
	case ModeGhost:
		s.ghostMsgs = append(s.ghostMsgs, msg)
		return nil
	case ModeLocal:
		return s.saveEncrypted(msg)
	default:
		return nil
	}
}

func (s *Store) saveEncrypted(msg chat.Message) error {
	if s.db == nil || s.gcm == nil {
		return fmt.Errorf("storage not initialized")
	}
	plain := []byte(msg.Content)
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ciphertext := s.gcm.Seal(nonce, nonce, plain, nil)
	enc := base64.StdEncoding.EncodeToString(ciphertext)
	ts := msg.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	_, err := s.db.Exec(
		`INSERT INTO messages (sender, content_enc, created_at) VALUES (?, ?, ?)`,
		msg.Sender, enc, ts.Format(time.RFC3339Nano),
	)
	return err
}

// Load returns stored messages for local mode.
func (s *Store) Load() ([]chat.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch s.mode {
	case ModeGhost:
		out := make([]chat.Message, len(s.ghostMsgs))
		copy(out, s.ghostMsgs)
		return out, nil
	case ModeLocal:
		return s.loadEncrypted()
	default:
		return nil, nil
	}
}

func (s *Store) loadEncrypted() ([]chat.Message, error) {
	rows, err := s.db.Query(`SELECT sender, content_enc, created_at FROM messages ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []chat.Message
	for rows.Next() {
		var sender, enc, created string
		if err := rows.Scan(&sender, &enc, &created); err != nil {
			return nil, err
		}
		raw, err := base64.StdEncoding.DecodeString(enc)
		if err != nil {
			return nil, err
		}
		nonceSize := s.gcm.NonceSize()
		if len(raw) < nonceSize {
			return nil, fmt.Errorf("invalid ciphertext")
		}
		plain, err := s.gcm.Open(nil, raw[:nonceSize], raw[nonceSize:], nil)
		if err != nil {
			return nil, err
		}
		ts, _ := time.Parse(time.RFC3339Nano, created)
		msgs = append(msgs, chat.Message{
			Sender:    sender,
			Content:   string(plain),
			Timestamp: ts,
		})
	}
	return msgs, rows.Err()
}

// Close releases resources and scrubs ghost-mode memory.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.mode == ModeGhost {
		for i := range s.ghostMsgs {
			s.ghostMsgs[i] = chat.Message{}
		}
		s.ghostMsgs = nil
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// ModeName returns a human-readable storage mode label.
func ModeName(m Mode) string {
	switch m {
	case ModeGhost:
		return "Ghost"
	case ModeLocal:
		return "Encrypted"
	default:
		return "Unknown"
	}
}
