package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	kvpb "madkv/kvstore/gen/kvpb"
	"google.golang.org/grpc"
)
// Entry point for the KV Store server application
// initialise an in memory kv store and start listening for client requests
type kvServer struct {
	kvpb.UnimplementedKVServer
}

func (s *kvServer) Ping(ctx context.Context, req *kvpb.PingRequest) (*kvpb.PingReply, error) {
	return &kvpb.PingReply{Msg: "pong: " + req.Msg}, nil
}

func main() {
	listenAddr := flag.String("listen", "127.0.0.1:50051", "ip:port to listen on")
	flag.Parse()

	lis, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}

	grpcServer := grpc.NewServer()
	kvpb.RegisterKVServer(grpcServer, &kvServer{})

	fmt.Printf("server listening on %s\n", *listenAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve failed: %v", err)
	}
}