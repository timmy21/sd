package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/timmy21/sd"
	"github.com/timmy21/sd/etcd"
)

func main() {
	var (
		svc       = flag.String("service", "", "service name")
		nodeID    = flag.String("node-id", "", "node id")
		nodeAddr  = flag.String("node-addr", "127.0.0.1", "node address")
		nodeTTL   = flag.Int64("node-ttl", 10, "node ttl (seconds)")
		endpoints = flag.String("etcd-endpoints", "127.0.0.1:2379", "etcd endpoints")
		prefix    = flag.String("etcd-key-prefix", "discovery", "etcd discovery key prefix")
	)
	flag.Parse()
	if *svc == "" {
		log.Fatal("service name is empty")
	}
	if *nodeID == "" {
		log.Fatal("node id is empty")
	}
	ctx := context.Background()
	strings.Split(*endpoints, ",")
	client, err := etcd.NewClient(ctx, strings.Split(*endpoints, ","), *prefix)
	if err != nil {
		log.Fatal(err)
	}
	s, err := etcd.NewService(client, *svc)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan sd.Event)
	s.Subscribe(ch)

	node := sd.Node{
		ID:   *nodeID,
		Addr: *nodeAddr,
	}
	err = s.Register(node, *nodeTTL)
	if err != nil {
		log.Fatal(err)
	}

	for evt := range ch {
		if evt.Err != nil {
			log.Fatal(evt.Err)
		}
		fmt.Println("nodes:")
		for _, node := range evt.Nodes {
			fmt.Printf("    node id=%s addr=%s\n", node.ID, node.Addr)
		}
		fmt.Println()
	}
}
