# OpenChat (ChatSSH)

[![Go Version](https://img.shields.io/github/go-mod/go-version/BRO-CODES-HERE/OpenChat?color=00ADD8)](https://golang.org)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](#how-to-use)
[![Platform support](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)](#precompiled-downloads)

OpenChat is a secure, terminal-based, decentralized chat application. It combines cryptographic SSH channels, Bubble Tea text user interface (TUI) aesthetics, and `go-libp2p` peer-to-peer transport with NAT hole punching.

With our built-in interactive setup wizard, you no longer need any terminal commands to host or join a chat!

---

## Precompiled Downloads

Get the precompiled executable directly for your operating system:

* 💻 **Windows:** [Download OpenChat-Win.exe (64-bit)](https://github.com/BRO-CODES-HERE/OpenChat/releases/download/v1.1.0/OpenChat-Win.exe)
* 🍎 **macOS:** [Download OpenChat-Mac.zip (Apple Silicon & Intel)](https://github.com/BRO-CODES-HERE/OpenChat/releases/download/v1.1.0/OpenChat-Mac.zip)
* 🐧 **Linux:** [Download OpenChat-linux (64-bit)](https://github.com/BRO-CODES-HERE/OpenChat/releases/download/v1.1.0/OpenChat-linux)

---

## How to Use

No complex terminal commands or configurations are required. Getting started is simple:

1. **Download:** Grab the binary for your operating system from the links above.
2. **Launch:** Double-click the downloaded file to start the **Interactive Setup Wizard**.
3. **Configure:** The wizard will guide you through choosing your username, deciding whether to host a room or connect to one, and selecting your storage preferences.
4. **Chat:** Once completed, the secure, responsive chat interface will open automatically!

---

## Core Features

* **Interactive Setup Wizard:** No command-line flags required. A guided step-by-step wizard configures everything directly inside your terminal window.
* **Bubble Tea UI:** A premium, fully responsive terminal user interface featuring colored message feeds, visual timestamps, dynamic hashed username colors, and a powerline status bar.
* **End-to-End Encryption:** Secured via standard SSH protocols and cryptographic handshakes, complete with randomart/emoji fingerprint checks to prevent middleman attacks.
* **P2P NAT Hole Punching:** Powered by `libp2p`. Automatically attempts direct connections across routers/firewalls, falling back to secure encrypted relays if direct access fails.
* **Storage Modes:** Choose between persistent local database logging (AES-256 encrypted SQLite) or high-privacy **Ghost Mode** (messages kept strictly in volatile RAM and completely wiped upon exit).

---

## Learn More

For complete developer details and advanced configurations:
* 📐 **[System Design Document](DesignSystem/ChatSSH.md)** — Architectural layout, wireframes, and connection models.
* ⚙️ **[CLI Flags Reference](Flags.md)** — Index for running custom manual settings via CLI arguments.
* 🔧 **[Security & Debugging Audit](Report01.md)** — Historical overview of issues resolved during security hardening.
