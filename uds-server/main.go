package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	socketPath = "/tmp/health.sock"
	pingByte   = byte(0x01)
	pongByte   = byte(0x02)
)

func main() {
	// 1. Clean up old socket file if it exists
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.RemoveAll(socketPath); err != nil {
			log.Fatalf("Failed to remove old socket: %v", err)
		}
	}

	// 2. Set up listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on UDS: %v", err)
	}
	defer listener.Close()

	fmt.Printf("UDS Server listening at %s\n", socketPath)

	// Set up graceful shutdown (removes the socket file)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nServer shutting down...")
		listener.Close()
		os.Remove(socketPath)
		os.Exit(0)
	}()

	// 3. Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("Accept error: %v", err)
			return
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Use minimal buffer for performance (1 byte)
	buf := make([]byte, 1)

	for {
		// Read the ping byte (0x01)
		n, err := conn.Read(buf)
		if err != nil {
			// Connection closed or error
			return
		}
		if n == 0 {
			// End of file / zero read
			return
		}

		// Check if it's the expected ping
		if buf[0] != pingByte {
			log.Printf("Received unexpected byte: %v", buf[0])
			return
		}

		// Write the pong byte (0x02) back
		buf[0] = pongByte
		if _, err := conn.Write(buf); err != nil {
			// Write error
			return
		}
	}
}
