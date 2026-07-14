# Debugging Report (Report01.md)

**Date:** July 12, 2026  
**Project:** ChatSSH (OpenChat)  
**Status:** All issues resolved. 100% tests passing.  

---

## 1. Executive Summary

This report documents the comprehensive debugging, diagnostic testing, and resolution of critical structural bugs within the ChatSSH codebase. Prior to this intervention, the application suffered from test flakiness, memory/goroutine leaks, terminal UI rendering issues on multibyte inputs, and a fatal design bug in the P2P connection layer that prevented nodes from dialing each other via raw IP/ports without pre-shared cryptographic Peer IDs.

Following the diagnostic sweep, **five core bugs** were identified and corrected. A new unit test suite has been introduced to cover the P2P connection layer. The entire codebase builds cleanly, and the test suite reports a 100% success rate.

---

## 2. Bug Analysis and Resolutions

### Bug 1: Global Path Pollution & Database Test Flakiness
* **Component:** `internal/storage` & `internal/keys`
* **Symptoms:** Running `go test ./...` repeatedly caused the test `TestLocalEncryptedStorage` to fail with `cipher: message authentication failed`.
* **Root Cause:** Both the `storage` and `keys` packages had their configurations hardcoded to write straight into the user's home directory under `~/.chatssh/messages.db`. Because unit tests ran on the host machine, they polluted this global path. If a database already existed with records encrypted under a different passphrase or key, the test would query the old rows during `.Load()` and fail to decrypt them, causing the cipher authentication failure.
* **Resolution:**
  - Modified [storage.go](internal/storage/storage.go) and [keys.go](internal/keys/keys.go) to support overriding the configuration directory via the `CHATSSH_HOME` environment variable.
  - Modified [storage_test.go](internal/storage/storage_test.go) to set `CHATSSH_HOME` to `t.TempDir()`. This isolates every test run's data into an ephemeral directory that is cleaned up automatically by Go's testing framework.

### Bug 2: Fatal P2P Peer ID Dial Failure
* **Component:** `internal/p2p`
* **Symptoms:** Starting the app in P2P mode and connecting via a raw IP/port (e.g. `--connect 192.168.1.100:4001`) failed.
* **Root Cause:** `go-libp2p` requires a cryptographically secure `Peer ID` (which is a hash of the host's public key) to perform connection handshakes and secure protocol negotiations. In the original `DialPeerByIP` method, `addrInfo` was created without a Peer ID. This caused `Host.Connect` to fail immediately. The fallback logic attempted to dial random bootstrap nodes from the peer pool, resulting in failure or routing to an incorrect server.
* **Resolution:**
  - Implemented the **Handshake-Mismatch Discovery technique** in [host.go](internal/p2p/host.go).
  - The client now dynamically generates a temporary dummy Peer ID and attempts to connect to the target `IP:port`.
  - This intentionally triggers a cryptographic handshake mismatch error, which contains the target's actual Peer ID (e.g., `expected <dummy_id>, but remote key matches <actual_id>`).
  - We extract the actual Peer ID from the error using string parsing, decode it, and perform a second, successful connect using the correct Peer ID.
  - Added a new automated unit test in [p2p_test.go](internal/p2p/p2p_test.go) verifying this end-to-end flow.

### Bug 3: P2P Goroutine and Listener Leak on Shutdown
* **Component:** `internal/app` & `internal/p2p`
* **Symptoms:** Exiting a P2P server session left background goroutines hanging on P2P stream acceptance.
* **Root Cause:** When running in P2P mode, `srv.Serve(ln)` is run in a separate goroutine. However, `ln` (the `streamListener`) was never closed during application shutdown. Since `streamListener.Accept()` blocks indefinitely on a channel read (`<-l.ch`) until the listener is closed, the goroutine running `srv.Serve` leaked and would block forever.
* **Resolution:**
  - Modified [app.go](internal/app/app.go) to add the P2P listener `ln.Close` to the application's cleanups. Closing `l.closed` causes `Accept()` to return `net.ErrClosed` immediately, allowing the server goroutine to terminate cleanly.

### Bug 4: Chat Hub Subscription Memory Leak
* **Component:** `internal/tui` & `internal/chat`
* **Symptoms:** Program memory usage grows continuously as chat messages are sent or received.
* **Root Cause:** In the Bubble Tea TUI, the `listen()` command is a one-shot task. To continuously listen to messages, `listen()` was returning a command that called `hub.Subscribe()` on every message. This created a new channel and appended it to the `hub.subs` slice on every iteration. These channels were never removed or closed, causing `hub.subs` to grow indefinitely, leaking memory for every single message.
* **Resolution:**
  - Modified [model.go](internal/tui/model.go) to store a single, persistent channel `sub <-chan chat.Message` in the TUI `Model` struct.
  - The subscription is now established once during model initialization in `New()`. The one-shot `listen()` command simply reads from this persistent channel, eliminating the memory leak.

### Bug 5: TUI UTF-8 Backspace Slicing Bug
* **Component:** `internal/tui`
* **Symptoms:** Pressing backspace after typing multi-byte characters (such as emojis or accented characters) outputs scrambled unicode characters or corrupts input rendering.
* **Root Cause:** The key handler for `backspace` sliced the string using byte indices: `m.input = m.input[:len(m.input)-1]`. Since multi-byte characters consist of 2 to 4 bytes in UTF-8, slicing by a single byte cuts the character in half, generating invalid UTF-8 bytes.
* **Resolution:**
  - Modified [model.go](internal/tui/model.go) to convert `m.input` to a rune slice, slice the rune slice, and convert it back to a string, ensuring safe UTF-8 character deletion.

---

## 3. Verification Results

### Automated Tests
The complete test suite runs successfully with zero cached test failures. The test logs are as follows:

```
$ go test -count=1 ./...
?   	github.com/BRO-CODES-HERE/OpenChat/cmd/chatssh	[no test files]
?   	github.com/BRO-CODES-HERE/OpenChat/internal/app	[no test files]
ok  	github.com/BRO-CODES-HERE/OpenChat/internal/chat	1.179s
ok  	github.com/BRO-CODES-HERE/OpenChat/internal/crypto	1.058s
?   	github.com/BRO-CODES-HERE/OpenChat/internal/keys	[no test files]
ok  	github.com/BRO-CODES-HERE/OpenChat/internal/p2p	8.175s
?   	github.com/BRO-CODES-HERE/OpenChat/internal/room	[no test files]
?   	github.com/BRO-CODES-HERE/OpenChat/internal/sshclient	[no test files]
?   	github.com/BRO-CODES-HERE/OpenChat/internal/sshserver	[no test files]
ok  	github.com/BRO-CODES-HERE/OpenChat/internal/storage	2.639s
ok  	github.com/BRO-CODES-HERE/OpenChat/internal/tui	0.474s
```

### Compiler Verification
The command `go build ./cmd/...` successfully compiles the binaries with no warnings or syntax errors.

---

## 4. Recommendations for Future Code Quality

1. **Continuous Integration (CI):** Integrate `go test -race ./...` into Github Actions to detect any potential concurrency issues early, especially in the multi-threaded Chat Hub.
2. **Linting Rules:** Introduce static code analyzers like `golangci-lint` to prevent raw byte slicing of strings, helping to catch unicode bugs in future TUI additions.
3. **Structured Logging:** Replace `fmt.Fprintln` or basic print loops with structured logging (e.g. `uber-go/zap` or standard library `slog`) to allow proper verbose logging in production environments while keeping production stdout clean.
