package app

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/BRO-CODES-HERE/OpenChat/internal/chat"
	"github.com/BRO-CODES-HERE/OpenChat/internal/crypto"
	"github.com/BRO-CODES-HERE/OpenChat/internal/keys"
	"github.com/BRO-CODES-HERE/OpenChat/internal/p2p"
	"github.com/BRO-CODES-HERE/OpenChat/internal/room"
	"github.com/BRO-CODES-HERE/OpenChat/internal/sshclient"
	"github.com/BRO-CODES-HERE/OpenChat/internal/sshserver"
	"github.com/BRO-CODES-HERE/OpenChat/internal/storage"
	"github.com/BRO-CODES-HERE/OpenChat/internal/tui"
)

// Options configures a ChatSSH session.
type Options struct {
	Mode       string // server, connect
	Addr       string
	UseP2P     bool
	Room       bool
	RoomName   string
	Storage    storage.Mode
	Passphrase string
	LocalUser  string
	ListenPort int
	Bootnodes  []string
}

// App orchestrates TUI, SSH, libp2p, storage, and room chat.
type App struct {
	opts     Options
	identity *keys.Identity
	hub      *chat.Hub
	store    *storage.Store
	p2pNode  *p2p.Node
	server   *sshserver.Server
	client   *sshclient.Client
	roomHost *room.Host
	cleanup  []func() error
	mu       sync.Mutex
}

// Run starts the ChatSSH application.
func Run(opts Options) error {
	identity, err := keys.LoadOrCreate()
	if err != nil {
		return err
	}

	store, err := storage.Open(opts.Storage, opts.Passphrase)
	if err != nil {
		return err
	}

	hub := chat.NewHub()
	a := &App{
		opts:     opts,
		identity: identity,
		hub:      hub,
		store:    store,
	}

	if saved, err := store.Load(); err == nil {
		for _, msg := range saved {
			hub.Publish(msg)
		}
	}

	hub.OnSend(func(msg chat.Message) {
		_ = store.Save(msg)
	})

	ctx := context.Background()

	var verifyScreen string

	switch opts.Mode {
	case "server":
		if err := a.startServer(ctx); err != nil {
			return err
		}
	case "connect":
		screen, err := a.startClient(ctx)
		if err != nil {
			return err
		}
		verifyScreen = screen
	default:
		return fmt.Errorf("unknown mode %q (use server or connect)", opts.Mode)
	}

	status := a.buildStatus()
	title := "ChatSSH"
	if opts.Room {
		title = fmt.Sprintf("ChatSSH Room: %s", opts.RoomName)
	}

	return tui.Run(tui.Config{
		Title:        title,
		LocalUser:    opts.LocalUser,
		Status:       status,
		Hub:          hub,
		VerifyScreen: verifyScreen,
		OnQuit: func() {
			a.shutdown()
		},
	})
}

func (a *App) startServer(ctx context.Context) error {
	var roomHost *room.Host
	if a.opts.Room {
		name := a.opts.RoomName
		if name == "" {
			name = "public"
		}
		roomHost = room.NewHost(name, a.hub)
		a.roomHost = roomHost
	}

	srv := sshserver.New(sshserver.Config{
		Identity: a.identity,
		Hub:      a.hub,
		Room:     roomHost,
		Addr:     a.opts.Addr,
	})
	a.server = srv
	sshserver.WireHub(a.hub, srv, roomHost)

	if a.opts.UseP2P {
		node, err := p2p.NewNode(ctx, p2p.Config{
			ListenPort:  a.opts.ListenPort,
			Bootnodes:   a.opts.Bootnodes,
			EnableRelay: true,
		})
		if err != nil {
			return err
		}
		a.p2pNode = node
		a.addCleanup(node.Close)

		ln := node.ListenSSH()
		go func() {
			_ = srv.Serve(ln)
		}()
	} else {
		go func() {
			_ = srv.ListenAndServe()
		}()
	}
	return nil
}

func (a *App) startClient(ctx context.Context) (string, error) {
	var verified crypto.Fingerprint
	verifyOnce := sync.Once{}
	onVerified := func(fp crypto.Fingerprint) {
		verifyOnce.Do(func() {
			verified = fp
		})
	}

	cfg := sshclient.Config{
		Identity:   a.identity,
		Hub:        a.hub,
		Addr:       a.opts.Addr,
		User:       a.opts.LocalUser,
		OnVerified: onVerified,
	}

	var client *sshclient.Client
	var err error

	if a.opts.UseP2P {
		node, err := p2p.NewNode(ctx, p2p.Config{
			ListenPort:  0,
			Bootnodes:   a.opts.Bootnodes,
			EnableRelay: true,
		})
		if err != nil {
			return "", err
		}
		a.p2pNode = node
		a.addCleanup(node.Close)

		var conn net.Conn
		if strings.Contains(a.opts.Addr, "/p2p/") {
			conn, err = node.DialSSH(ctx, a.opts.Addr)
		} else {
			host, port := splitHostPort(a.opts.Addr)
			conn, err = node.DialPeerByIP(ctx, host, port)
		}
		if err != nil {
			return "", err
		}
		client, err = sshclient.DialConn(cfg, conn)
	} else {
		client, err = sshclient.Dial(cfg)
	}
	if err != nil {
		return "", err
	}
	a.client = client
	a.addCleanup(client.Close)

	if verified.SHA256 != "" {
		return crypto.VerificationScreen(verified, a.opts.Addr), nil
	}
	return "", nil
}

func (a *App) buildStatus() string {
	parts := []string{storage.ModeName(a.opts.Storage)}
	fp := crypto.ShortFingerprint(a.identity.HostPublicKey())
	parts = append(parts, "key:"+fp)

	if a.opts.UseP2P && a.p2pNode != nil {
		parts = append(parts, "p2p:"+a.p2pNode.ID()[:8])
		addrs := a.p2pNode.Addrs()
		if len(addrs) > 0 {
			parts = append(parts, "addr:"+addrs[0])
		}
	} else if a.server != nil {
		parts = append(parts, "listen:"+a.server.Addr())
	}

	if a.roomHost != nil {
		parts = append(parts, fmt.Sprintf("peers:%d", a.roomHost.ClientCount()))
	}
	return strings.Join(parts, " | ")
}

func splitHostPort(addr string) (string, int) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return strings.TrimPrefix(addr, ":"), 2222
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return host, port
}

func (a *App) addCleanup(fn func() error) {
	a.mu.Lock()
	a.cleanup = append(a.cleanup, fn)
	a.mu.Unlock()
}

func (a *App) shutdown() {
	a.mu.Lock()
	fns := append([]func() error(nil), a.cleanup...)
	a.mu.Unlock()
	for _, fn := range fns {
		_ = fn()
	}
	if a.server != nil {
		_ = a.server.Close()
	}
	_ = a.store.Close()
}

// DefaultBootnodes returns public libp2p bootstrap nodes.
func DefaultBootnodes() []string {
	return []string{
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXHJ14uQ4ACWy",
	}
}

// ParseStorageMode converts a CLI flag to storage.Mode.
func ParseStorageMode(ghost bool) storage.Mode {
	if ghost {
		return storage.ModeGhost
	}
	return storage.ModeLocal
}

// Unused import guard for tea in case of future programmatic TUI.
var _ = tea.Quit
