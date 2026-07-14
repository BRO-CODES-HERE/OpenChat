# OpenChat (ChatSSH)

[![Go Version](https://img.shields.io/github/go-mod/go-version/BRO-CODES-HERE/OpenChat?color=00ADD8)](https://golang.org)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](#quick-start)
[![Platform support](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)](#precompiled-downloads)

OpenChat (originally ChatSSH) is a secure, terminal-based, decentralized chat application. It combines cryptographic SSH channels, Bubble Tea text user interface (TUI) aesthetics, and `go-libp2p` peer-to-peer transport with NAT hole punching.

---

## Precompiled Downloads

Get the precompiled executable directly for your operating system:

* 💻 **Windows:** [Download OpenChat-Win.exe (64-bit)](file:///c:/Users/hdk99/Desktop/Bro's_Git/OpenChat/OpenChat-Win.exe)
* 🍎 **macOS:** [Download OpenChat-Mac (Apple Silicon & Intel)](file:///c:/Users/hdk99/Desktop/Bro's_Git/OpenChat/OpenChat-Mac)
* 🐧 **Linux:** [Download OpenChat-linux (64-bit)](file:///c:/Users/hdk99/Desktop/Bro's_Git/OpenChat/OpenChat-linux)

---

## Useful Reference Guides

Learn more about using and configuring OpenChat:

* 🚀 **[Quick Start Guide](file:///c:/Users/hdk99/Desktop/Bro's_Git/OpenChat/GetStarted.md)** — Step-by-step instructions on starting your first server and connecting clients.
* ⚙️ **[CLI Flags Reference](file:///c:/Users/hdk99/Desktop/Bro's_Git/OpenChat/Flags.md)** — Complete configuration index for running custom settings via command-line arguments.
* 📐 **[System Design Document](file:///c:/Users/hdk99/Desktop/Bro's_Git/OpenChat/DesignSystem/ChatSSH.md)** — Architectural layout, wire frames, connection models, and cryptographic details.
* 🔧 **[Security & Debugging Audit](file:///c:/Users/hdk99/Desktop/Bro's_Git/OpenChat/Report01.md)** — Historical overview of issues resolved during security hardening (UTF-8, goroutine leaks, handshake retry logic).

---

## Core Features

* **Interactive TUI Setup Wizard:** Starts automatically if you launch the application without CLI arguments (e.g. when double-clicking the binary). Walk through step-by-step options (Username, Hosting vs. Connecting, Storage type) completely in the terminal.
* **Premium Bubble Tea UI:** Fully responsive terminal layout using Bubble Tea, rendering message feeds, colored timestamps, dynamic hashed username coloring, and a powerline-style status bar.
* **SSH Encrypted Transport:** Uses standard SSH protocols to execute encrypted end-to-end handshakes. Incorporates visual fingerprint checks (Randomart and Emojiart) for MitM verification.
* **Decentralized NAT Traversal:** Leverages P2P networking powered by `libp2p`. Autodetects networks, performs NAT hole-punching (DCUtR), and falls back to Circuit Relays to connect nodes behind firewalls automatically.
* **Flexible Storage Profiles:**
  * *Local Encrypted Storage (Default):* Saves chat history inside a local SQLite database encrypted with AES-256 (passphrase-derived key).
  * *Ghost Mode:* Volatile RAM-only message caching. Wipes and scrubs all chats from memory upon application termination, writing zero bytes to the disk.

---

## Build from Source

### Prerequisites
* Go 1.22+ installed.
* GCC (Optional, for running tests).

### Compile Binary
```bash
go build -o OpenChat ./cmd/chatssh
```

---

## Quick Start

### 1. Launch Setup Wizard
Double-click the downloaded binary file or run the compiled executable without parameters to load the interactive setup wizard:
```bash
./OpenChat
```

### 2. Connect from Terminal (CLI Option)
If you prefer to bypass the setup wizard and launch directly using the command line:

* **Host a room:**
  ```bash
  go run ./cmd/chatssh --mode server --addr :2222 --user alice
  ```
* **Connect as a client:**
  ```bash
  go run ./cmd/chatssh --connect "172.29.160.1:2222" --user bob
  ```

---

## Project Directory Layout

```
cmd/chatssh/          Application main entrypoint
internal/
  app/                Application setup, CLI parsing, and interactive setup wizard
  chat/               Event hub and subscriber mechanisms
  crypto/             Key exchange and fingerprint utilities
  keys/               Ed25519 secure key management
  p2p/                libp2p network stacks and NAT routing
  room/               Multiplexed room broadcasters
  sshclient/          SSH client wrapper
  sshserver/          SSH server implementation
  storage/            AES-GCM SQLite database and Ghost storage handlers
  tui/                Bubble Tea view and components
DesignSystem/         Original architecture design definitions
WebView/              Stunning 3D P2P landing page website assets
```

---

## Contributing

We welcome contributions to OpenChat! To contribute:
1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/NewFeature`).
3. Commit your changes (`git commit -m "Add some NewFeature"`).
4. Push to the branch (`git push origin feature/NewFeature`).
5. Open a Pull Request.
