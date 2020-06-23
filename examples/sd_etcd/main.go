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
		instID    = flag.String("instance-id", "", "instance id")
		instAddr  = flag.String("instance-addr", "127.0.0.1", "instance address")
		instTTL   = flag.Int64("instance-ttl", 10, "instance ttl (seconds)")
		endpoints = flag.String("etcd-endpoints", "127.0.0.1:2379", "etcd endpoints")
		prefix    = flag.String("etcd-key-prefix", "discovery", "etcd discovery key prefix")
	)
	flag.Parse()

	if *svc == "" {
		log.Fatal("service name is empty")
	}
	if *instID == "" {
		log.Fatal("inst id is empty")
	}

	ctx := context.Background()
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

	inst := sd.Instance{
		ID:   *instID,
		Addr: *instAddr,
	}
	err = s.Register(inst, *instTTL)
	if err != nil {
		log.Fatal(err)
	}

	for evt := range ch {
		if evt.Err != nil {
			log.Fatal(evt.Err)
		}
		fmt.Println("instances:")
		for _, inst := range evt.Instances {
			fmt.Printf("    instance id=%s addr=%s\n", inst.ID, inst.Addr)
		}
		fmt.Println()
	}
}
