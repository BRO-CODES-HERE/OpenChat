package sshserver

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
	"github.com/BRO-CODES-HERE/OpenChat/internal/keys"
	"github.com/BRO-CODES-HERE/OpenChat/internal/room"
	"github.com/BRO-CODES-HERE/OpenChat/internal/tui"
)

// Config configures the SSH chat server.
type Config struct {
	Identity *keys.Identity
	Hub      *chat.Hub
	Room     *room.Host
	Addr     string
}

// Server serves SSH chat sessions.
type Server struct {
	cfg      Config
	listener net.Listener
	mu       sync.Mutex
	sessions map[string]ssh.Channel
}

// New creates an SSH chat server.
func New(cfg Config) *Server {
	return &Server{
		cfg:      cfg,
		sessions: make(map[string]ssh.Channel),
	}
}

func (s *Server) serverConfig() *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{}, nil
		},
		NoClientAuth: true,
	}
	config.AddHostKey(s.cfg.Identity.HostKey)
	return config
}

// ListenAndServe starts accepting SSH connections.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return err
	}
	s.listener = ln
	return s.Serve(ln)
}

// Serve accepts connections on an existing listener (for libp2p transport).
func (s *Server) Serve(ln net.Listener) error {
	s.listener = ln
	config := s.serverConfig()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(config, conn)
	}
}

// Addr returns the bound listen address.
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.cfg.Addr
}

// Close stops the server.
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConn(config *ssh.ServerConfig, conn net.Conn) {
	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}
	defer sshConn.Close()
	go ssh.DiscardRequests(reqs)

	peerID := sshConn.User()
	if peerID == "" {
		peerID = sshConn.RemoteAddr().String()
	}

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "only session channels supported")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go s.handleSession(peerID, channel, requests)
	}
}

func (s *Server) handleSession(peerID string, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	s.registerSession(peerID, channel)
	defer s.unregisterSession(peerID)

	if s.cfg.Room != nil {
		s.cfg.Room.AddClient(peerID, channel)
		s.cfg.Room.AnnounceJoin(peerID)
		defer s.cfg.Room.AnnounceLeave(peerID)
	}

	go func() {
		for req := range requests {
			ok := false
			switch req.Type {
			case "shell", "pty-req":
				ok = true
			}
			_ = req.Reply(ok, nil)
		}
	}()

	_, _ = io.WriteString(channel, Greeting(s.Addr()))

	scanner := bufio.NewScanner(channel)
	for scanner.Scan() {
		msg, err := tui.ParseMessage(scanner.Text())
		if err != nil {
			continue
		}
		if msg.Sender == "" {
			msg.Sender = peerID
		}
		s.cfg.Hub.Publish(msg)
		if s.cfg.Room != nil {
			s.cfg.Room.Broadcast(msg, peerID)
		} else {
			s.BroadcastToSessions(msg)
		}
	}
}

func (s *Server) registerSession(id string, ch ssh.Channel) {
	s.mu.Lock()
	s.sessions[id] = ch
	s.mu.Unlock()
}

func (s *Server) unregisterSession(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

// BroadcastToSessions sends a message to all active SSH sessions.
func (s *Server) BroadcastToSessions(msg chat.Message) {
	line := tui.FormatMessage(msg) + "\n"
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, ch := range s.sessions {
		if id == msg.Sender {
			continue
		}
		_, _ = io.WriteString(ch, line)
	}
}

// WireHub configures the hub to forward local sends to SSH sessions/room.
func WireHub(hub *chat.Hub, srv *Server, roomHost *room.Host) {
	hub.OnSend(func(msg chat.Message) {
		if roomHost != nil {
			roomHost.Broadcast(msg, msg.Sender)
		} else if srv != nil {
			srv.BroadcastToSessions(msg)
		}
	})
}

// Greeting returns a welcome line for new SSH clients.
func Greeting(addr string) string {
	return fmt.Sprintf("Connected to ChatSSH at %s. Messages are E2E encrypted over SSH.\n", addr)
}
