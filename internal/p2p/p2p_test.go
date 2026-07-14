package p2p

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDialPeerByIP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// 1. Create Node 1 (listener/server)
	node1, err := NewNode(ctx, Config{
		ListenPort:  0, // automatic free port
		Bootnodes:   nil,
		EnableRelay: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer node1.Close()

	ln := node1.ListenSSH()
	defer ln.Close()

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		t.Log("Server: waiting for Accept...")
		conn, err := ln.Accept()
		if err != nil {
			t.Logf("Server: Accept failed: %v", err)
			return
		}
		t.Log("Server: Accept succeeded!")
		defer conn.Close()

		// Read data to keep connection open
		buf := make([]byte, 10)
		_, _ = conn.Read(buf)
	}()

	// 2. Create Node 2 (client)
	node2, err := NewNode(ctx, Config{
		ListenPort:  0,
		Bootnodes:   nil,
		EnableRelay: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer node2.Close()

	// Extract the port of Node 1
	addrs := node1.Host.Addrs()
	if len(addrs) == 0 {
		t.Fatal("no addrs on node1")
	}
	var port int
	for _, addr := range addrs {
		if strings.HasPrefix(addr.String(), "/ip4/127.0.0.1/tcp/") {
			fmt.Sscanf(addr.String(), "/ip4/127.0.0.1/tcp/%d", &port)
			break
		}
	}
	if port == 0 {
		for _, addr := range addrs {
			if strings.Contains(addr.String(), "/tcp/") {
				parts := strings.Split(addr.String(), "/tcp/")
				if len(parts) > 1 {
					fmt.Sscanf(parts[1], "%d", &port)
					break
				}
			}
		}
	}
	if port == 0 {
		t.Fatalf("could not extract port from addrs: %v", addrs)
	}

	t.Logf("Client: Dialing 127.0.0.1:%d", port)
	// Dial node 1 by IP and port (without Peer ID)
	conn, err := node2.DialPeerByIP(ctx, "127.0.0.1", port)
	if err != nil {
		t.Fatalf("failed to dial peer by IP: %v", err)
	}
	t.Log("Client: Dial succeeded!")
	defer conn.Close()

	// Write data to flush stream to server
	t.Log("Client: Writing data to connection...")
	_, err = conn.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("failed to write data: %v", err)
	}
	t.Log("Client: Write succeeded!")

	select {
	case <-serverDone:
		t.Log("Test: serverDone received!")
	case <-time.After(15 * time.Second):
		t.Fatal("timeout waiting for server to accept connection")
	}
}
