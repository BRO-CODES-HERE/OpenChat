package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BRO-CODES-HERE/OpenChat/internal/app"
	"github.com/BRO-CODES-HERE/OpenChat/internal/storage"
)

func main() {
	if len(os.Args) == 1 {
		opts, err := app.RunWizard()
		if err != nil {
			if err == app.ErrWizardAborted {
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "setup wizard error: %v\n", err)
			fmt.Println("\nPress Enter to exit...")
			var temp string
			_, _ = fmt.Scanln(&temp)
			os.Exit(1)
		}
		if err := app.Run(opts); err != nil {
			fmt.Fprintf(os.Stderr, "\nchatssh error: %v\n", err)
			fmt.Println("\nPress Enter to exit...")
			var temp string
			_, _ = fmt.Scanln(&temp)
			os.Exit(1)
		}
		return
	}

	mode := flag.String("mode", "server", "Run mode: server or connect")
	addr := flag.String("addr", ":2222", "Listen address (server) or remote host:port (connect)")
	connect := flag.String("connect", "", "Shorthand to connect to host:port")
	user := flag.String("user", "me", "Local display name")
	p2p := flag.Bool("p2p", false, "Use libp2p transport with NAT traversal")
	room := flag.Bool("room", false, "Create a public room (server mode only)")
	roomName := flag.String("room-name", "public", "Room name when --room is set")
	ghost := flag.Bool("ghost", false, "Ghost mode: messages kept in RAM only, erased on exit")
	passphrase := flag.String("passphrase", "chatssh", "Passphrase for local encrypted storage")
	port := flag.Int("port", 4001, "libp2p listen port when --p2p is set")
	flag.Parse()

	runMode := *mode
	remote := *addr
	if *connect != "" {
		runMode = "connect"
		remote = *connect
	}

	storageMode := app.ParseStorageMode(*ghost)
	if storageMode == storage.ModeLocal && *passphrase == "" {
		fmt.Fprintln(os.Stderr, "passphrase required for local storage (or use --ghost)")
		os.Exit(1)
	}

	opts := app.Options{
		Mode:       runMode,
		Addr:       remote,
		UseP2P:     *p2p,
		Room:       *room,
		RoomName:   *roomName,
		Storage:    storageMode,
		Passphrase: *passphrase,
		LocalUser:  *user,
		ListenPort: *port,
		Bootnodes:  app.DefaultBootnodes(),
	}

	if err := app.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "chatssh: %v\n", err)
		os.Exit(1)
	}
}
