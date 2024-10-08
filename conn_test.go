package simnet_test

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/picatz/simnet"
)

func ExampleConn() {
	cfg := &simnet.Config{
		Latency:       100 * time.Millisecond,
		Jitter:        50 * time.Millisecond,
		Bandwidth:     256 * 1024, // 256 KBps
		LossRate:      0.05,       // 5% packet loss
		ReorderRate:   0.05,       // 5% packet reordering
		DuplicateRate: 0.05,       // 5% packet duplication
		Seed:          42,         // Seed for deterministic behavior
	}

	dialer := simnet.NewDialer(cfg)

	// Start a server
	go startServer()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Use the simulated dialer to connect to the server
	conn, err := dialer.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("Dial error: %v\n", err)
		return
	}
	defer conn.Close()

	// Send data
	message := "Hello, simnet!"
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Printf("Write error: %v\n", err)
		return
	}

	// Wait for the data to be received
	time.Sleep(200 * time.Millisecond)

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			return
		}
		fmt.Printf("Read error: %v\n", err)
		return
	}
	fmt.Printf("Received: %s\n", string(buf[:n]))

	// Output:
	// Server received: Hello, simnet!
}

func startServer() {
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("Listen error: %v\n", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			return
		}

		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 1024)
			n, err := c.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Printf("Server read error: %v\n", err)
				return
			}
			fmt.Printf("Server received: %s\n", string(buf[:n]))

			// Echo back the message
			c.Write(buf[:n])
		}(conn)
	}
}
