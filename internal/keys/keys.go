package keys

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

const (
	hostKeyFile  = "host_key"
	userKeyFile  = "user_key"
	userCertFile = "user_key.pub"
)

// Identity holds SSH host and client credentials.
type Identity struct {
	HostKey ssh.Signer
	UserKey ssh.Signer
	Dir     string
}

// LoadOrCreate loads existing keys or generates new Ed25519 SSH keys.
func LoadOrCreate() (*Identity, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	hostKey, err := loadOrCreateSigner(filepath.Join(dir, hostKeyFile))
	if err != nil {
		return nil, fmt.Errorf("host key: %w", err)
	}
	userKey, err := loadOrCreateSigner(filepath.Join(dir, userKeyFile))
	if err != nil {
		return nil, fmt.Errorf("user key: %w", err)
	}

	pubPath := filepath.Join(dir, userCertFile)
	if _, err := os.Stat(pubPath); os.IsNotExist(err) {
		if err := os.WriteFile(pubPath, ssh.MarshalAuthorizedKey(userKey.PublicKey()), 0o600); err != nil {
			return nil, err
		}
	}

	return &Identity{HostKey: hostKey, UserKey: userKey, Dir: dir}, nil
}

// HostPublicKey returns the SSH public key for fingerprint display.
func (id *Identity) HostPublicKey() ssh.PublicKey {
	return id.HostKey.PublicKey()
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

func loadOrCreateSigner(path string) (ssh.Signer, error) {
	if data, err := os.ReadFile(path); err == nil {
		return ssh.ParsePrivateKey(data)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pemBlock, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0o600); err != nil {
		return nil, err
	}

	return ssh.NewSignerFromKey(priv)
}
