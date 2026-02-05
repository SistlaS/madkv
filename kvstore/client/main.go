package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	kvpb "madkv/kvstore/gen/kvpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	serverAddr := flag.String("server", "127.0.0.1:50051", "server ip:port")
	msg := flag.String("msg", "hello", "message to ping")
	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	c := kvpb.NewKVClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.Ping(ctx, &kvpb.PingRequest{Msg: *msg})
	if err != nil {
		log.Fatalf("Ping RPC failed: %v", err)
	}

	fmt.Println(resp.Msg)
}
