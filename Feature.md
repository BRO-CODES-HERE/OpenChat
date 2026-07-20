# OpenChat — Feature Roadmap

A curated list of planned and proposed features to make OpenChat more powerful, user-friendly, and fun. Contributions are welcome — pick any feature and open a PR!

---

## ?? Quick Wins — Easy to Build

### ?? TUI & UX Polish

| Feature | Description |
|---------|-------------|
| **Scrollable message history** | Use `? / ?` arrow keys to scroll back through previous messages. Currently only a fixed window is shown. |
| **Username color persistence** | Save a user's assigned color so the same username always appears in the same color across sessions. |
| **Mention highlights** | Highlight and ring the terminal bell when your username is mentioned (`@alice`). |
| **Timestamp toggle** | Press `T` to show or hide message timestamps in the feed. |
| **Unread message counter** | Display a badge or indicator when new messages arrive while the user is typing. |

### ?? Input Improvements

| Feature | Description |
|---------|-------------|
| **In-chat commands** | Support slash commands: `/clear`, `/users`, `/quit`, `/help`, `/dm <user> <msg>`. |
| **Edit last message** | Press `?` to recall and edit the last sent message before re-sending. |
| **Multi-line messages** | Use `Shift+Enter` to insert a newline before sending a longer message. |

---

## ??? Security & Privacy

| Feature | Description |
|---------|-------------|
| **Room passwords** | Optional passphrase required to join a room, enforced during the SSH handshake. |
| **Message expiry / TTL** | Auto-delete messages after a configurable time (e.g. 10 minutes, 1 hour). Works in Ghost Mode too. |
| **Read receipts** | Show `?` (delivered) and `??` (read by all) indicators on messages. |
| **Identity verification badge** | Display a `verified` or `unverified` badge next to a username based on known public keys. |
| **Key pinning** | Remember a host's public key fingerprint after first connect and warn on mismatch (TOFU model). |

---

## ?? Connectivity & Networking

| Feature | Description |
|---------|-------------|
| **LAN room discovery (mDNS)** | Broadcast room presence on the local network via mDNS so nearby users can find rooms without manually typing IP addresses. |
| **QR code room invite** | Generate a QR code containing the full connection string. Scan it with a phone camera to instantly get the address. |
| **Auto-reconnect on drop** | Automatically retry the connection if the peer disconnects temporarily, with exponential backoff. |
| **Dedicated relay / headless mode** | A `--relay` mode that runs a persistent, always-on message relay server with no TUI. |
| **STUN/TURN fallback** | Improve NAT traversal by supporting standard STUN/TURN servers as a last-resort relay when libp2p relays fail. |

---

## ?? Chat Features

| Feature | Description |
|---------|-------------|
| **File transfer** | Send small files (images, documents) through the encrypted SSH channel using a simple protocol. |
| **Emoji reactions** | React to any message with an emoji (`:+1:`, `:fire:`, `:heart:`) by hovering and pressing a shortcut. |
| **Private DMs (whisper)** | `/dm bob Hey!` sends a private message to a specific user inside a multi-user room. |
| **Message search** | `/find <keyword>` searches message history and highlights matching results. |
| **Bot / webhook support** | A simple stdin/stdout interface for plugging in bots (weather, reminders, GitHub notifications). |
| **Typing indicators** | Show `alice is typing...` in real-time when another user is composing a message. |
| **Message pinning** | Pin important messages to the top of the chat for easy reference. |

---

## ??? Launcher & OS Integration

| Feature | Description |
|---------|-------------|
| **Config file support** | Read settings from `~/.chatssh/config.toml` — preferred username, default room, theme, storage mode, etc. |
| **System tray icon** | Run OpenChat as a background daemon with a system tray icon on Windows and macOS. |
| **Desktop notifications** | Fire an OS-level toast notification when a new message arrives while the app is in the background. |
| **Web UI companion** | A lightweight local HTTP server (e.g. `:8080`) that renders the chat in a browser as a fallback. |
| **Auto-start on login** | Register OpenChat as a startup service so it connects to a saved room on boot. |

---

## ?? Fun & Viral Features

| Feature | Description |
|---------|-------------|
| **ASCII art welcome banner** | Display a random cool ASCII art banner on connect instead of plain text. |
| **Theme switcher** | Built-in color themes: Dracula, Nord, Gruvbox, Catppuccin, Tokyo Night. Switch with `/theme <name>`. |
| **User list sidebar** | A live side panel showing all connected users, their status, and peer counts. |
| **Custom status messages** | Set a short status like `/status ?? coding` that other users can see next to your name. |
| **Confetti / fun reactions** | Trigger a terminal confetti animation when someone reacts with ??. |

---

## ?? Priority Matrix

| Priority | Feature | Effort | Impact |
|----------|---------|--------|--------|
| ?? **High** | Scrollable message history | Low | High |
| ?? **High** | `/help`, `/users`, `/clear` commands | Low | High |
| ?? **High** | Auto-reconnect on drop | Medium | High |
| ?? **High** | Config file (`config.toml`) | Medium | High |
| ? **Medium** | Typing indicators | Medium | Medium |
| ? **Medium** | LAN room discovery (mDNS) | Medium | High |
| ? **Medium** | QR code room invite | Low | Medium |
| ? **Medium** | Theme switcher | Low | Medium |
| ? **Medium** | Desktop notifications | Medium | Medium |
| ?? **Stretch** | File transfer | High | High |
| ?? **Stretch** | Web UI companion | High | Medium |
| ?? **Stretch** | Bot / webhook support | High | Medium |

---

## ?? Contributing a Feature

1. Pick a feature from this list (or propose your own).
2. Open an Issue to discuss the design before coding.
3. Fork the repo and create a branch: `git checkout -b feature/typing-indicators`
4. Submit a Pull Request with tests where applicable.

See [README.md](README.md) for full contribution guidelines.
