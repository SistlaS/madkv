package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	kvpb "madkv/kvstore/gen/kvpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  client --server <ip:port> --op put    --key <k> --value <v>
  client --server <ip:port> --op get    --key <k>
  client --server <ip:port> --op swap   --key <k> --value <v>
  client --server <ip:port> --op delete --key <k>
  client --server <ip:port> --op scan   --start <k1> --end <k2>

Examples:
  client --server 127.0.0.1:50051 --op put --key a --value 1
  client --server 127.0.0.1:50051 --op get --key a
  client --server 127.0.0.1:50051 --op scan --start a --end z
`)
}

func main() {
	serverAddr := flag.String("server", "127.0.0.1:50051", "server ip:port")
	op := flag.String("op", "", "operation: put|get|swap|delete|scan")
	key := flag.String("key", "", "key for put/get/swap/delete")
	value := flag.String("value", "", "value for put/swap")
	start := flag.String("start", "", "scan start key")
	end := flag.String("end", "", "scan end key")
	timeout := flag.Duration("timeout", 2*time.Second, "rpc timeout")
	flag.Usage = usage
	flag.Parse()

	if *op == "" {
		usage()
		os.Exit(2)
	}

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// If your proto service name is KV, change to kvpb.NewKVClient(conn)
	c := kvpb.NewKVSClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	switch strings.ToLower(*op) {
	case "put":
		if *key == "" || *value == "" {
			log.Fatalf("put requires --key and --value")
		}
		resp, err := c.Put(ctx, &kvpb.PutRequest{Key: *key, Value: *value})
		if err != nil {
			log.Fatalf("Put RPC failed: %v", err)
		}
		fmt.Printf("PUT %s %s (found=%v)\n", *key, *value, resp.Found)

	case "get":
		if *key == "" {
			log.Fatalf("get requires --key")
		}
		resp, err := c.Get(ctx, &kvpb.GetRequest{Key: *key})
		if err != nil {
			log.Fatalf("Get RPC failed: %v", err)
		}
		if !resp.Found {
			fmt.Printf("GET %s null\n", *key)
		} else {
			fmt.Printf("GET %s %s\n", *key, resp.Value)
		}

	case "swap":
		if *key == "" || *value == "" {
			log.Fatalf("swap requires --key and --value")
		}
		resp, err := c.Swap(ctx, &kvpb.SwapRequest{Key: *key, Value: *value})
		if err != nil {
			log.Fatalf("Swap RPC failed: %v", err)
		}
		if !resp.Found {
			fmt.Printf("SWAP %s null\n", *key)
		} else {
			fmt.Printf("SWAP %s old=%s new=%s\n", *key, resp.OldValue, *value)
		}

	case "delete":
		if *key == "" {
			log.Fatalf("delete requires --key")
		}
		resp, err := c.Delete(ctx, &kvpb.DeleteRequest{Key: *key})
		if err != nil {
			log.Fatalf("Delete RPC failed: %v", err)
		}
		fmt.Printf("DELETE %s (found=%v)\n", *key, resp.Found)

	case "scan":
		if *start == "" || *end == "" {
			log.Fatalf("scan requires --start and --end")
		}
		resp, err := c.Scan(ctx, &kvpb.ScanRequest{StartKey: *start, EndKey: *end})
		if err != nil {
			log.Fatalf("Scan RPC failed: %v", err)
		}
		fmt.Printf("SCAN %s %s (%d pairs)\n", *start, *end, len(resp.Pairs))
		for _, p := range resp.Pairs {
			fmt.Printf("  %s %s\n", p.Key, p.Value)
		}

	default:
		log.Fatalf("unknown --op %q (expected put|get|swap|delete|scan)", *op)
	}
}
