package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	pb "app/pingpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	_ "time/tzdata"
)

func Setup(ctx context.Context, wg *sync.WaitGroup, address string) (client pb.PingServiceClient, err error) {
	if address != `` {
		address = os.Getenv(address)
	}
	if address == `` {
		address = "localhost:50051"
	}

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    60 * time.Second,
			Timeout: 5 * time.Second,
		}),
	)
	if err != nil {
		slog.Error(`gRPC new client`, `error`, err)
		return
	}

	wg.Go(func() {
		<-ctx.Done()
		conn.Close()
	})

	client = pb.NewPingServiceClient(conn)
	slog.Info("gRPC new client", "address", address)
	return
}

func main() {
	if loc, err := time.LoadLocation(`UTC`); err != nil {
		panic(err)
	} else {
		time.Local = loc // sets the global timezone to UTC
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo})))
	slog.Info(`Go`, `Version`, runtime.Version(), `OS`, runtime.GOOS, `ARCH`, runtime.GOARCH, `now`, time.Now(), `Local`, time.Local)

	var wg sync.WaitGroup
	var ctx = context.Background()

	client, err := Setup(ctx, &wg, ``)
	if err != nil {
		return
	}

	warmup(client)

	count := 0
	var sum time.Duration
	t0 := time.Now()
	for time.Since(t0) < 60*time.Second {
		var ctx = context.Background()
		start := time.Now()
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		_, err := client.Ping(ctx, &pb.PingRequest{Payload: "ping"})
		cancel()
		if err != nil {
			log.Printf("ping error: %v", err)
		} else {
			sum += time.Since(start)
			count++
		}
		time.Sleep(10 * time.Millisecond)
	}

	var avg time.Duration
	if count > 0 {
		avg = sum / time.Duration(count)
	}

	fmt.Printf("Ping run complete: count=%d, avg_latency=%v \n", count, avg)
}

func warmup(client pb.PingServiceClient) {
	var ctx = context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	for range 10 {
		_, err := client.Ping(ctx, &pb.PingRequest{Payload: "warmup"})
		if err != nil {
			slog.Error(`ping`, `error`, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
