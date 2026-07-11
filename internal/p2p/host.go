package p2p

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/multiformats/go-multiaddr"
)

const ProtocolID = "/chatssh/ssh/1.0.0"

// Node wraps a libp2p host with ChatSSH transport helpers.
type Node struct {
	Host host.Host
}

// Config configures libp2p node creation.
type Config struct {
	ListenPort int
	Bootnodes  []string
	EnableRelay bool
}

// NewNode creates a libp2p host with NAT traversal support.
func NewNode(ctx context.Context, cfg Config) (*Node, error) {
	priv, _, err := crypto.GenerateEd25519Key(nil)
	if err != nil {
		return nil, err
	}

	rm, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.DefaultLimits.AutoScale()))
	if err != nil {
		return nil, err
	}

	port := cfg.ListenPort
	if port == 0 {
		port = 0
	}

	opts := []libp2p.Option{
		libp2p.Identity(priv),
		libp2p.ResourceManager(rm),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
		libp2p.NATPortMap(),
	}

	if cfg.EnableRelay {
		opts = append(opts,
			libp2p.EnableRelay(),
			libp2p.EnableAutoRelayWithStaticRelays(parseBootPeers(cfg.Bootnodes)),
		)
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	for _, addr := range cfg.Bootnodes {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			continue
		}
		info, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			continue
		}
		_ = h.Connect(ctx, *info)
	}

	return &Node{Host: h}, nil
}

func parseBootPeers(addrs []string) []peer.AddrInfo {
	var peers []peer.AddrInfo
	for _, addr := range addrs {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			continue
		}
		info, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			continue
		}
		peers = append(peers, *info)
	}
	return peers
}

// Addrs returns the node's advertised multiaddresses.
func (n *Node) Addrs() []string {
	var out []string
	for _, addr := range n.Host.Addrs() {
		out = append(out, fmt.Sprintf("%s/p2p/%s", addr, n.Host.ID()))
	}
	return out
}

// ID returns the peer ID string.
func (n *Node) ID() string {
	return n.Host.ID().String()
}

// ListenSSH registers a stream handler and returns a net.Listener adapter.
func (n *Node) ListenSSH() net.Listener {
	ln := newStreamListener()
	n.Host.SetStreamHandler(ProtocolID, func(s network.Stream) {
		ln.acceptStream(s)
	})
	return ln
}

// DialSSH opens a libp2p stream and wraps it as net.Conn.
func (n *Node) DialSSH(ctx context.Context, target string) (net.Conn, error) {
	ma, err := multiaddr.NewMultiaddr(target)
	if err != nil {
		return nil, err
	}
	info, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return nil, err
	}
	if err := n.Host.Connect(ctx, *info); err != nil {
		return nil, err
	}
	stream, err := n.Host.NewStream(ctx, info.ID, ProtocolID)
	if err != nil {
		return nil, err
	}
	return NewStreamConn(stream), nil
}

// DialPeerByIP attempts connection using IP-based multiaddr patterns.
func (n *Node) DialPeerByIP(ctx context.Context, ip string, port int) (net.Conn, error) {
	if port == 0 {
		port = 4001
	}
	target := fmt.Sprintf("/ip4/%s/tcp/%d", ip, port)
	ma, err := multiaddr.NewMultiaddr(target)
	if err != nil {
		return nil, err
	}
	addrInfo := peer.AddrInfo{Addrs: []multiaddr.Multiaddr{ma}}
	if err := n.Host.Connect(ctx, addrInfo); err != nil {
		// Fallback: try relayed connection via known peers
		for _, p := range n.Host.Network().Peers() {
			stream, err := n.Host.NewStream(ctx, p, ProtocolID)
			if err == nil {
				return NewStreamConn(stream), nil
			}
		}
		return nil, err
	}
	peers := n.Host.Network().Peers()
	if len(peers) == 0 {
		return nil, fmt.Errorf("no peers found at %s", ip)
	}
	stream, err := n.Host.NewStream(ctx, peers[0], ProtocolID)
	if err != nil {
		return nil, err
	}
	return NewStreamConn(stream), nil
}

// Close shuts down the libp2p host.
func (n *Node) Close() error {
	return n.Host.Close()
}

// streamListener adapts libp2p streams to net.Listener.
type streamListener struct {
	ch     chan network.Stream
	closed chan struct{}
	addr   net.Addr
}

func newStreamListener() *streamListener {
	return &streamListener{
		ch:     make(chan network.Stream, 16),
		closed: make(chan struct{}),
		addr:   &streamAddr{},
	}
}

func (l *streamListener) acceptStream(s network.Stream) {
	select {
	case l.ch <- s:
	case <-l.closed:
		_ = s.Close()
	}
}

func (l *streamListener) Accept() (net.Conn, error) {
	select {
	case s := <-l.ch:
		return NewStreamConn(s), nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *streamListener) Close() error {
	select {
	case <-l.closed:
		return net.ErrClosed
	default:
		close(l.closed)
		return nil
	}
}

func (l *streamListener) Addr() net.Addr {
	return l.addr
}

type streamAddr struct{}

func (streamAddr) Network() string { return "libp2p" }
func (streamAddr) String() string { return "libp2p-stream" }

// StreamConn wraps a libp2p stream as net.Conn for SSH.
type StreamConn struct {
	network.Stream
	mu       sync.Mutex
	deadline time.Time
}

// NewStreamConn wraps a libp2p stream.
func NewStreamConn(s network.Stream) *StreamConn {
	return &StreamConn{Stream: s}
}

func (c *StreamConn) Read(b []byte) (int, error) {
	return c.Stream.Read(b)
}

func (c *StreamConn) Write(b []byte) (int, error) {
	return c.Stream.Write(b)
}

func (c *StreamConn) LocalAddr() net.Addr {
	return maAddr{c.Stream.Conn().LocalMultiaddr().String()}
}

func (c *StreamConn) RemoteAddr() net.Addr {
	return maAddr{c.Stream.Conn().RemoteMultiaddr().String()}
}

type maAddr struct{ s string }

func (a maAddr) Network() string { return "libp2p" }
func (a maAddr) String() string  { return a.s }

func (c *StreamConn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	c.deadline = t
	c.mu.Unlock()
	_ = c.Stream.SetReadDeadline(t)
	return c.Stream.SetWriteDeadline(t)
}

func (c *StreamConn) SetReadDeadline(t time.Time) error {
	return c.Stream.SetReadDeadline(t)
}

func (c *StreamConn) SetWriteDeadline(t time.Time) error {
	return c.Stream.SetWriteDeadline(t)
}

func (c *StreamConn) Close() error {
	return c.Stream.Close()
}

var _ io.Closer = (*StreamConn)(nil)
