# ChatSSH CLI Flags Reference (Flags.md)

This document provides a comprehensive guide to all command-line flags supported by the **ChatSSH** application. You can use these flags to configure your display name, storage type, network transport, and group room hosting.

---

## Quick Reference Table

| Flag | Default Value | Description | Supported Modes |
| :--- | :--- | :--- | :--- |
| [`--mode`](#--mode) | `server` | Running mode: `server` (host) or `connect` (client) | Both |
| [`--addr`](#--addr) | `:2222` | Listening address (for server) or target host:port (for client) | Both |
| [`--connect`](#--connect) | `""` | Shorthand alias to connect to a target server (auto-sets `--mode connect`) | Client |
| [`--user`](#--user) | `me` | Your local chat display name | Both |
| [`--p2p`](#--p2p) | `false` | Enables Peer-to-Peer transport with hole-punching and NAT traversal | Both |
| [`--port`](#--port) | `4001` | Local libp2p protocol listener port (when `--p2p` is active) | Both |
| [`--room`](#--room) | `false` | Starts a multiplexed star-topology public group chat room | Server |
| [`--room-name`](#--room-name) | `public` | Name of the hosted group room (when `--room` is active) | Server |
| [`--ghost`](#--ghost) | `false` | Ephemeral mode: holds messages in RAM only and erases them on exit | Both |
| [`--passphrase`](#--passphrase) | `chatssh` | Passphrase used to encrypt/decrypt local history storage | Both |

---

## Flag Details & Usage Examples

### `--mode`
* **Type:** `string`
* **Supported Values:** `server`, `connect`
* **Description:** Determines the running role of the application. 
  * `server`: The application hosts a chat room on a specific port and waits for incoming connections.
  * `connect`: The application joins an active chat room hosted on a remote server.
* **Example:**
  ```bash
  go run .\cmd\chatssh --mode connect --addr "192.168.1.100:2222"
  ```

### `--addr`
* **Type:** `string`
* **Description:** The address parameter depending on the active `--mode`.
  * In **Server Mode**: The bind port/address to listen on (e.g. `:2222` binds to port 2222 on all network interfaces).
  * In **Connect Mode**: The destination socket address of the server (e.g. `10.0.0.5:2222`).
* **Example:**
  ```bash
  go run .\cmd\chatssh --mode server --addr "127.0.0.1:3333"
  ```

### `--connect`
* **Type:** `string`
* **Description:** A convenient shorthand for client connections. Specifying `--connect "IP:port"` automatically switches the application's role to client mode and sets the connection address.
* **Example:**
  ```bash
  go run .\cmd\chatssh --connect "172.29.160.1:2222" --user bob
  ```

### `--user`
* **Type:** `string`
* **Description:** The display username that other chat participants see. Standard system labels (like `system`) are reserved.
* **Example:**
  ```bash
  go run .\cmd\chatssh --connect "localhost:2222" --user Alice_In_Wonderland
  ```

### `--p2p`
* **Type:** `boolean` (flag presence sets to `true`)
* **Description:** Enables decentralized transport powered by **libp2p**. This manages automatic hole punching (NAT-Dcutr), address advertisements, DHT peer routing, and Circuit Relay v2 fallbacks. Use this mode when connecting over the internet where servers are behind routers/firewalls and port forwarding is not configured.
* **Example:**
  * Host:
    ```bash
    go run .\cmd\chatssh --mode server --p2p --user host_user
    ```
  * Client:
    ```bash
    go run .\cmd\chatssh --p2p --connect "/ip4/192.168.1.100/tcp/4001/p2p/Qm..." --user client_user
    ```

### `--port`
* **Type:** `int`
* **Description:** The port used by the P2P network stack for inter-node communication when `--p2p` is active.
* **Example:**
  ```bash
  go run .\cmd\chatssh --mode server --p2p --port 5005
  ```

### `--room`
* **Type:** `boolean` (flag presence sets to `true`)
* **Description:** Configures the server to host a multiplexed group chat room (star-topology). If disabled, the server only hosts a direct one-on-one session. When enabled, any number of clients can connect and chat simultaneously in a shared feed.
* **Example:**
  ```bash
  go run .\cmd\chatssh --mode server --room --room-name "developers-hub"
  ```

### `--room-name`
* **Type:** `string`
* **Description:** Sets a custom name for the group chat room when `--room` is enabled. This name is displayed in the TUI header banner of all connected clients.
* **Example:**
  ```bash
  go run .\cmd\chatssh --mode server --room --room-name "General"
  ```

### `--ghost`
* **Type:** `boolean` (flag presence sets to `true`)
* **Description:** Activates **Ghost Mode**. All incoming and outgoing messages are stored only in volatile memory (RAM) and are wiped cleanly when the application terminates. No files or directories are written to local disk.
* **Example:**
  ```bash
  go run .\cmd\chatssh --connect "localhost:2222" --user shadow --ghost
  ```

### `--passphrase`
* **Type:** `string`
* **Description:** The passphrase used to derive an AES-256 key (via PBKDF2) to encrypt your local chat log storage database (`messages.db`). If you restart a session using the same passphrase, your past chat logs are decrypted and loaded back into the feed automatically.
* **Example:**
  ```bash
  go run .\cmd\chatssh --mode server --passphrase "my-secure-key-phrase"
  ```
