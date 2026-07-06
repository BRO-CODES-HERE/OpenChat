# ChatSSH: Terminal-Based Secure P2P Chat over SSH

This document outlines the system design, tech stack, and implementation flow for **ChatSSH**, a terminal-based, end-to-end encrypted (E2E) chat application that connects two or more users worldwide using their IP addresses, running directly in the terminal over the SSH protocol.

---

## 1. System Overview

ChatSSH is a peer-to-peer (P2P) terminal application that combines the security of the SSH protocol with decentralized NAT traversal. It allows two users behind firewalls to establish a secure chat session using their public IP addresses, without configuring port forwarding.

```
+-----------------------------------------------------------------+
|                         Terminal UI                             |
|              (Bubble Tea / Go TUI Engine)                       |
+-------------------------------+---------------------------------+
                                |
+-------------------------------+---------------------------------+
|                     SSH Client & Server                         |
|      - Authenticates peers using Ed25519 SSH Keys                |
|      - Creates secure interactive chat channels                 |
+-------------------------------+---------------------------------+
                                |
+-------------------------------+---------------------------------+
|                     libp2p Network Layer                        |
|      - Performs NAT traversal (Hole Punching via AutoNAT/DCUtR)  |
|      - Falls back to E2E-encrypted relay nodes if direct fails  |
+-----------------------------------------------------------------+
```

### Key Features
* **Zero Config NAT Traversal:** Connects peers worldwide using public IPs via UDP/TCP hole punching and libp2p relays.
* **Pure SSH Protocol:** Uses SSH (RFC 4253) for key exchange, host verification, and channel multiplexing.
* **End-to-End Encryption (E2EE):** SSH provides transport-layer encryption. Even if traffic passes through public relays, the relay cannot decrypt the messages.
* **Flexible Storage:** Option to encrypt and store chat logs locally using SQLite + SQLCipher, or run in "Ghost Mode" where logs exist only in RAM and are securely shredded on exit.
* **Group Chat (Room Chat):** One user acts as a room host (running a multi-channel SSH server), and other peers connect as clients to form a group.

---

## 2. Network & NAT Traversal Protocol

Connecting two computers worldwide using only their IP addresses is difficult due to Network Address Translation (NAT) and firewalls. ChatSSH solves this by running SSH over **libp2p** transport streams.

### The Connection Flow

```mermaid
sequenceDiagram
    autonumber
    participant Peer A (Client/Host)
    participant STUN/DHT Server
    participant Peer B (Client/Server)

    Note over Peer A, Peer B: Phase 1: Address Discovery
    Peer A->>STUN/DHT Server: Request external IP & Port mapping
    STUN/DHT Server-->>Peer A: Return mapped public address
    Peer B->>STUN/DHT Server: Request external IP & Port mapping
    STUN/DHT Server-->>Peer B: Return mapped public address

    Note over Peer A, Peer B: Phase 2: Connection Initiation (User inputs Peer B's IP)
    Peer A->>Peer B: Attempt direct UDP (QUIC) hole punching connection
    Peer B->>Peer A: Simultaneous hole punching attempt
    
    alt Hole Punching Succeeds
        Peer A<->>Peer B: Direct P2P TCP/UDP connection established
    else Hole Punching Fails (Symmetric NAT)
        Note over Peer A, Peer B: Fallback to Libp2p Circuit Relay (TURN equivalent)
        Peer A->>STUN/DHT Server: Connect via relay node
        Peer B->>STUN/DHT Server: Connect via relay node
        Peer A<->>Peer B: Relayed connection (E2E SSH encrypted)
    end

    Note over Peer A, Peer B: Phase 3: SSH Handshake
    Peer A->>Peer B: Initiate SSH Session (Key Exchange, Ed25519)
    Peer B->>Peer A: SSH Host Key Exchange & Authentication
    Peer A->>Peer B: Open SSH Interactive Session Channel
```

### Addressing Mechanics
When both systems run the application:
1. The app starts a **libp2p host** listening on a random port.
2. The app contacts public STUN servers/libp2p bootnodes to discover its own public IP and NAT type.
3. To connect, User A enters User B's public IP.
4. The libp2p layer uses the IP to route a connection attempt. It queries the libp2p DHT (Distributed Hash Table) using the target IP or performs direct connection attempts on standard ports (e.g. `4001`) to initiate hole punching (DCUtR).

---

## 3. Cryptography & Security Architecture

SSH naturally provides strong cryptographic primitives. ChatSSH leverages them to ensure no third party can read messages.

### SSH Session Setup
1. **Identity Generation:** Upon first launch, the app generates a unique SSH Host Key (Ed25519) and a client key. These are stored locally in the application's config directory.
2. **Key Exchange (KEX):** Diffie-Hellman or ECDH (Curve25519) is performed to establish a shared session key.
3. **Session Encryption:** Symmetric encryption using `AES-GCM` or `ChaCha20-Poly1305` protects all terminal inputs and outputs.
4. **Peer Authentication:**
   * **Direct Auth:** Peers authenticate using public key authentication. Users verify each other's Host Key fingerprint (shown as a visual randomart / emoji sequence) in the UI to prevent Man-in-the-Middle (MITM) attacks.

### Relayed Security
If a direct P2P connection fails, libp2p routes traffic through a **Circuit Relay v2** node.
* **Security Guard:** The relay node only sees packet metadata and encrypted SSH frames. It does *not* possess the SSH private keys. Therefore, it is mathematically impossible for the relay to intercept or read the chat content.

---

## 4. Local Message Storage & Privacy

Users are given two storage options when launching a chat session:

### Option A: Local Storage (Secure Logs)
* Messages are saved in a local SQLite database.
* To protect the database from local access, it is encrypted using **SQLCipher (AES-256)**.
* The encryption key is derived from a passphrase entered by the user at startup, stretched using PBKDF2/Argon2.

### Option B: Ghost Mode (Erase Forever)
* Message history is kept exclusively in-memory (RAM) in a volatile buffer.
* No data is written to disk.
* Upon closing the session or exiting the application, the memory buffers are overwritten with zeroes (memory scrubbing) before the process exits to prevent cold-boot memory extraction attacks.

---

## 5. Group Chat (Room Chat) Protocol

Group chats are implemented using a **Star Topology** mediated by one of the peers acting as the "Room Host".

```
        +---------------+
        |  Peer C       | (SSH Client)
        +-------+-------+
                |
                | (SSH Conn)
        +-------+-------+         (SSH Conn)        +---------------+
        |  Peer A       |<--------------------------|  Peer B       | (SSH Client)
        |  (Room Host)  |                           +---------------+
        +-------+-------+
                |
                | (SSH Conn)
        +-------+-------+
        |  Peer D       | (SSH Client)
        +---------------+
```

1. **Host Configuration:** Peer A selects "Create Public Room". Their app starts a multiplexed SSH Server and registers the room's multiaddress.
2. **Joining the Room:** Peer B, C, and D connect to Peer A's IP address.
3. **Session Multiplexing:**
   * Peer A allocates a virtual `pty` session for each connecting peer.
   * When Peer B sends a message over their SSH channel, Peer A's server intercepts it and broadcasts it to Peer C and Peer D's active SSH stdout channels.
4. **Decentralized Hub:** Peer A is the central hub. If Peer A leaves, the room closes, or the room host can migrate ownership by promoting another peer to server hosting.

---

## 6. Recommended Tech Stack

| Component | Technology | Rationale |
| :--- | :--- | :--- |
| **Language** | Go (Golang) | High-performance networking, first-class libp2p support, excellent SSH libraries, compilation to a single binary. |
| **P2P Networking** | `go-libp2p` | Handles NAT hole punching (STUN/TURN equivalents), relays, and stream multiplexing. |
| **SSH Protocol** | `golang.org/x/crypto/ssh` | The standard, robust Go implementation of SSH client and server protocols. |
| **Terminal UI (TUI)** | `charmbracelet/bubbletea` | A modern Elm-architecture-based TUI framework for Go. Allows creating rich, smooth animations and UI layouts in terminal. |
| **TUI Styling** | `charmbracelet/lipgloss` | Style builder for terminal apps (colors, borders, layouts). |
| **Local Storage** | `modernc.org/sqlite` / SQLCipher | Zero-dependency SQLite driver for Go, combined with SQLCipher for local database encryption. |

---

## 7. Implementation Flow & Phases

### Phase 1: Local SSH TUI Chat (Prototype)
* **Goal:** Implement the Terminal UI chat layout and message rendering.
* **Steps:**
  1. Build a CLI app using `bubbletea` with a split-screen layout (top: chat feed, bottom: text input).
  2. Implement a local-only SSH server using `crypto/ssh` that serves this TUI.
  3. Verify connection locally using standard `ssh localhost -p 2222`.

### Phase 2: Peer-to-Peer SSH Transport
* **Goal:** Replace raw TCP sockets with libp2p streams.
* **Steps:**
  1. Initialize libp2p host nodes.
  2. Write a custom network dialer/listener that wraps libp2p streams into `net.Conn` interfaces.
  3. Pass these libp2p stream connections into the SSH client and server handshakes.

### Phase 3: NAT Traversal & Hole Punching
* **Goal:** Enable connections over the internet.
* **Steps:**
  1. Integrate libp2p's AutoNAT and DCUtR protocols.
  2. Configure bootnodes to assist with address lookup.
  3. Add Circuit Relay v2 configuration so connections succeed even behind strict symmetric NATs.

### Phase 4: Room Chat (Multi-user)
* **Goal:** Support group rooms.
* **Steps:**
  1. Build room hosting logic in the SSH server component to multiplex messages.
  2. Maintain a list of active connection channels and broadcast incoming lines.

### Phase 5: Encryption & Storage Polish
* **Goal:** Secure the messages.
* **Steps:**
  1. Implement SQLCipher for database logging.
  2. Build "Ghost Mode" memory buffer zeroing.
  3. Render the host public-key randomart verification screen upon connection startup.
