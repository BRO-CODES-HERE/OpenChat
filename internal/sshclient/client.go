package sshclient

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
	"github.com/BRO-CODES-HERE/OpenChat/internal/crypto"
	"github.com/BRO-CODES-HERE/OpenChat/internal/keys"
	"github.com/BRO-CODES-HERE/OpenChat/internal/tui"
)

// Config configures an SSH chat client connection.
type Config struct {
	Identity   *keys.Identity
	Hub        *chat.Hub
	Addr       string
	User       string
	HostKey    ssh.PublicKey
	OnVerified func(crypto.Fingerprint)
}

// Client connects to a remote SSH chat server.
type Client struct {
	cfg      Config
	conn     ssh.Conn
	session  *ssh.Session
	writer   io.WriteCloser
	serverKey ssh.PublicKey
}

// Dial connects to addr using TCP and performs the SSH handshake.
func Dial(cfg Config) (*Client, error) {
	conn, err := net.DialTimeout("tcp", cfg.Addr, 15*time.Second)
	if err != nil {
		return nil, err
	}
	return handshake(cfg, conn)
}

// DialConn performs SSH handshake over an existing net.Conn (libp2p transport).
func DialConn(cfg Config, conn net.Conn) (*Client, error) {
	return handshake(cfg, conn)
}

func handshake(cfg Config, conn net.Conn) (*Client, error) {
	var serverKey ssh.PublicKey

	sshConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(cfg.Identity.UserKey),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			serverKey = key
			fp := crypto.HostFingerprint(key)
			if cfg.OnVerified != nil {
				cfg.OnVerified(fp)
			}
			if cfg.HostKey != nil && !bytes.Equal(key.Marshal(), cfg.HostKey.Marshal()) {
				return io.EOF // host key mismatch
			}
			return nil
		},
		Timeout: 15 * time.Second,
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, cfg.Addr, sshConfig)
	if err != nil {
		return nil, err
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, err
	}

	stdinR, stdinW := io.Pipe()
	session.Stdin = stdinR

	stdoutR, stdoutW := io.Pipe()
	session.Stdout = stdoutW
	session.Stderr = stdoutW

	if err := session.Shell(); err != nil {
		session.Close()
		client.Close()
		return nil, err
	}

	c := &Client{
		cfg:       cfg,
		conn:      sshConn,
		session:   session,
		writer:    stdinW,
		serverKey: serverKey,
	}

	go c.readLoop(stdoutR)
	c.wireSend()
	return c, nil
}

func (c *Client) readLoop(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		msg, err := tui.ParseMessage(line)
		if err != nil {
			c.cfg.Hub.Publish(chat.Message{Sender: "remote", Content: line})
			continue
		}
		c.cfg.Hub.Publish(msg)
	}
}

func (c *Client) wireSend() {
	c.cfg.Hub.OnSend(func(msg chat.Message) {
		line := tui.FormatMessage(msg) + "\n"
		_, _ = io.WriteString(c.writer, line)
	})
}

// Close terminates the SSH session.
func (c *Client) Close() error {
	if c.session != nil {
		_ = c.session.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// RemoteFingerprint returns the server's host key fingerprint after connect.
func (c *Client) RemoteFingerprint() crypto.Fingerprint {
	return crypto.HostFingerprint(c.serverKey)
}
