package main

import (
	"fmt"
	"log"
	"net"
	"slices"
	"time"
)

const (
	socketPath = "/tmp/health.sock"
	pingByte   = byte(0x01)
	pongByte   = byte(0x02)
	numPings   = 10000 // Number of calls for accurate statistics
)

func main() {
	fmt.Printf("Running %d UDS health checks against %s...\n", numPings, socketPath)

	// 1. Establish a persistent connection
	// UDS connections are very fast to establish.
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to dial UDS: %v\n(Is the server running?)", err)
	}
	defer conn.Close()

	// 2. Prepare buffers
	pingBuf := []byte{pingByte}
	pongBuf := make([]byte, 1)

	// 3. Run benchmark and record latencies
	latencies := make([]time.Duration, 0, numPings)

	for range numPings {
		start := time.Now()

		// Send Ping
		if _, err := conn.Write(pingBuf); err != nil {
			log.Fatalf("Write error: %v", err)
		}

		// Read Pong
		_, err := conn.Read(pongBuf)
		if err != nil {
			log.Fatalf("Read error: %v", err)
		}

		// Validate (optional, but good for integrity)
		if pongBuf[0] != pongByte {
			log.Fatalf("Received invalid pong byte: %v", pongBuf[0])
		}

		// Record RTT
		latencies = append(latencies, time.Since(start))
	}

	// 4. Calculate statistics
	slices.Sort(latencies)

	mean := func() time.Duration {
		var total time.Duration
		for _, l := range latencies {
			total += l
		}
		return total / time.Duration(len(latencies))
	}()

	p95Index := int(0.95 * float64(len(latencies)))
	p99Index := int(0.99 * float64(len(latencies)))

	fmt.Println("--- UDS Latency Results ---")
	fmt.Printf("Samples: %d\n", numPings)
	fmt.Printf("Min Latency: %s\n", latencies[0])
	fmt.Printf("Average (Mean): %s\n", mean)
	fmt.Printf("P95 Latency: %s\n", latencies[p95Index])
	fmt.Printf("P99 Latency: %s\n", latencies[p99Index])
}
