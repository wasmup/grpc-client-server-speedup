package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"time"

	pb "app/pingpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type server struct {
	pb.UnimplementedPingServiceServer
}

func (s *server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Reply: "Pong: " + req.Payload}, nil
}

func main() {
	address := ":50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error(`tcp listen`, `address`, address, `error`, err)
		return
	}

	// Enable HTTP/2 keepalive and related settings to avoid idle-connection teardown.
	// These settings help keep latency low for ping-like RPCs over idle periods.
	srv := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    60 * time.Second, // send keepalive ping every 60s
			Timeout: 20 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	pb.RegisterPingServiceServer(srv, &server{})

	slog.Info(`server listening on`, `address`, address)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
