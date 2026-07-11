# Get Started — 3 Friends Chat

> **Scenario:** Alice (Windows), Bob (Mac), Charlie (Linux).  
> Alice hosts the room, Bob and Charlie connect.

---

## Step 1: Install Go (Everyone)

Download and install **Go 1.22+** :

- **Windows**: https://go.dev/dl/ (`.msi` installer)
- **Mac**: https://go.dev/dl/ (`.pkg` installer) or `brew install go`
- **Linux**: https://go.dev/dl/ (`.tar.gz`) or `sudo apt install golang-go`

---

## Step 2: Get the Code

Open a **terminal** on each machine and run:

```
git clone https://github.com/BRO-CODES-HERE/OpenChat
cd OpenChat
```
On Windows replace `\` with `\` inside the path.

*(If `git` is not installed, Alice can build the binary and share the file — see Step 2b below.)*

### Step 2b (Optional) — Share the Binary Instead

Alice builds the app for all platforms:

```
go build -o chatssh.exe .\cmd\chatssh                          # for Windows
go build -o chatssh-mac .\cmd\chatssh                           # for Mac (Intel)
GOOS=darwin GOARCH=arm64 go build -o chatssh-mac-m1 .\cmd\chatssh  # for Mac (M1/M2)
GOOS=linux go build -o chatssh-linux .\cmd\chatssh              # for Linux
```

Alice shares these files with Bob and Charlie. They just double-click or run `./chatssh-mac` in terminal.

---

## Step 3: Alice (Host) — Start the Chat Room

**Windows (PowerShell):**
```
go run .\cmd\chatssh --mode server --addr :2222 --user alice
```

**Mac / Linux:**
```
go run ./cmd/chatssh --mode server --addr :2222 --user alice
```

Keep this terminal open. You'll see a chat screen appear.

---

## Step 4: Alice — Find Your IP

Open a **new** terminal.

**Windows:**
```
ipconfig
```
Look for **IPv4 Address** (e.g. `192.168.1.100`).

**Mac:**
```
ifconfig
```
Look for `inet` under `en0` or `en1` (e.g. `192.168.1.100`).

**Linux:**
```
ip a
```
Look for `inet` under your network interface (e.g. `192.168.1.100`).

Share this IP with Bob and Charlie.

---

## Step 5: Bob & Charlie — Connect

In their terminal (inside the `OpenChat` folder):

```
go run ./cmd/chatssh --connect "192.168.1.100:2222" --user bob
```

Replace `192.168.1.100` with Alice's actual IP. Use their own name as `--user`.

> **Note:** Windows PowerShell users **must** put the IP in quotes.  
> Mac / Linux users can use quotes or write it directly: `--connect 192.168.1.100:2222`

---

## Step 6: Chat

Type any message and press **Enter**. Everyone in the room sees it.

To exit, press **Ctrl+C**.

---

## Troubleshooting

| Problem | Fix |
|---------|------|
| `connection refused` | Allow port `2222` in **Windows Firewall** on Alice's machine |
| `The system cannot find the file specified` | Windows: put IP in quotes `--connect "10.0.0.5:2222"` |
| `permission denied` | Mac/Linux: run `chmod +x chatssh-mac` or use `go run` |
| Can't connect from outside the house | Use `--p2p` mode (see README) or set up port forwarding on Alice's router |
