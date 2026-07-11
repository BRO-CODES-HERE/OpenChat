package crypto_test

import (
	"crypto/ed25519"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/BRO-CODES-HERE/OpenChat/internal/crypto"
)

func TestHostFingerprint(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	fp := crypto.HostFingerprint(signer.PublicKey())
	if fp.SHA256 == "" || fp.MD5 == "" || fp.Emoji == "" {
		t.Fatalf("missing fingerprint fields: %+v", fp)
	}
	if !strings.Contains(fp.RandomArt, "SHA256") {
		t.Fatalf("expected randomart fingerprint: %s", fp.RandomArt)
	}
	screen := crypto.VerificationScreen(fp, "localhost:2222")
	if !strings.Contains(screen, "Verify peer host key") {
		t.Fatalf("unexpected screen: %s", screen)
	}
}
