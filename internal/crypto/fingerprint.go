package crypto

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Fingerprint holds displayable host-key verification data.
type Fingerprint struct {
	SHA256   string
	MD5      string
	RandomArt string
	Emoji    string
}

// HostFingerprint builds verification strings for an SSH public key.
func HostFingerprint(pub ssh.PublicKey) Fingerprint {
	sha := sha256.Sum256(pub.Marshal())
	md5sum := md5.Sum(pub.Marshal())

	return Fingerprint{
		SHA256:    formatFingerprint(sha[:], ":"),
		MD5:       formatFingerprint(md5sum[:], ":"),
		RandomArt: ssh.FingerprintSHA256(pub),
		Emoji:     emojiFingerprint(sha[:]),
	}
}

func formatFingerprint(data []byte, sep string) string {
	parts := make([]string, len(data))
	for i, b := range data {
		parts[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(parts, sep)
}

func emojiFingerprint(data []byte) string {
	emojis := []string{"🔴", "🟠", "🟡", "🟢", "🔵", "🟣", "⚫", "⚪"}
	var b strings.Builder
	for i := 0; i < 8 && i < len(data); i++ {
		b.WriteString(emojis[int(data[i])%len(emojis)])
	}
	return b.String()
}

// VerificationScreen renders the host-key trust prompt.
func VerificationScreen(fp Fingerprint, host string) string {
	var b strings.Builder
	b.WriteString("Verify peer host key before continuing\n")
	b.WriteString(strings.Repeat("─", 50) + "\n")
	b.WriteString(fmt.Sprintf("Host:     %s\n", host))
	b.WriteString(fmt.Sprintf("SHA256:   %s\n", fp.SHA256))
	b.WriteString(fmt.Sprintf("MD5:      %s\n", fp.MD5))
	b.WriteString(fmt.Sprintf("Emoji:    %s\n", fp.Emoji))
	b.WriteString(fmt.Sprintf("Key:      %s\n", fp.RandomArt))
	b.WriteString(strings.Repeat("─", 50) + "\n")
	b.WriteString("Press Y to trust, N to abort")
	return b.String()
}

// ShortFingerprint returns a compact SHA256 prefix for the status bar.
func ShortFingerprint(pub ssh.PublicKey) string {
	sha := sha256.Sum256(pub.Marshal())
	return hex.EncodeToString(sha[:6])
}
