# ChatSSH (OpenChat)

Terminal-based secure P2P chat over SSH with libp2p NAT traversal.

## Features

- **Bubble Tea TUI** — split-screen chat feed and input
- **SSH transport** — Ed25519 keys, encrypted channels (E2EE)
- **libp2p networking** — AutoNAT, hole punching (DCUtR), Circuit Relay v2 fallback
- **Room chat** — star-topology group rooms with broadcast
- **Storage modes** — local AES-256 encrypted SQLite logs, or Ghost Mode (RAM-only)

## Requirements

- Go 1.22+
- OpenSSH client (optional, for `ssh localhost -p 2222` verification)

## Build

```bash
go build -o chatssh ./cmd/chatssh
```

## Usage

### Phase 1 — Local SSH server

Start a local server with TUI:

```bash
go run ./cmd/chatssh --mode server --addr :2222 --user alice
```

Connect from another terminal (native SSH client):

```bash
ssh -p 2222 -o StrictHostKeyChecking=no me@localhost
```

Or use the built-in client + TUI:

```bash
go run ./cmd/chatssh --connect localhost:2222 --user bob
```

### P2P mode (libp2p + NAT traversal)

Host with libp2p:

```bash
go run ./cmd/chatssh --mode server --p2p --port 4001 --user alice
```

Connect using peer multiaddr or IP:

```bash
go run ./cmd/chatssh --connect <peer-ip>:4001 --p2p --user bob
```

### Room chat

Create a public room:

```bash
go run ./cmd/chatssh --mode server --room --room-name lobby --user host
```

Others join:

```bash
go run ./cmd/chatssh --connect localhost:2222 --user guest1
```

### Storage

**Local encrypted storage** (default):

```bash
go run ./cmd/chatssh --mode server --passphrase "your-secret"
```

**Ghost mode** (no disk writes, memory scrubbed on exit):

```bash
go run ./cmd/chatssh --mode server --ghost
```

## Project layout

```
cmd/chatssh/          CLI entrypoint
internal/
  app/                Application orchestration
  chat/               Message hub
  crypto/             Host-key fingerprint verification
  keys/               Ed25519 SSH key management
  p2p/                libp2p host, stream net.Conn adapter
  room/               Star-topology room broadcast
  sshclient/          SSH client
  sshserver/          SSH server
  storage/            Encrypted SQLite + Ghost Mode
  tui/                Bubble Tea chat UI
DesignSystem/         Original design docs
```

## Design docs

See [DesignSystem/ChatSSH.md](DesignSystem/ChatSSH.md) for the full system design and implementation phases.
