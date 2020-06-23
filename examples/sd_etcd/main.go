package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/timmy21/sd"
	"github.com/timmy21/sd/etcd"
)

func main() {
	var (
		help      = flag.Bool("-help", false, "print usage")
		service   = flag.String("service", "test", "service name")
		instID    = flag.String("instance-id", "ds1", "instance id")
		instAddr  = flag.String("instance-addr", "127.0.0.1", "instance address")
		instTTL   = flag.Int64("instance-ttl", 10, "instance ttl (seconds)")
		endpoints = flag.String("etcd-endpoints", "127.0.0.1:2379", "etcd endpoints")
		prefix    = flag.String("etcd-key-prefix", "discovery", "etcd discovery key prefix")
	)
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	ctx := context.Background()
	client, err := etcd.NewClient(ctx, strings.Split(*endpoints, ","), *prefix)
	if err != nil {
		log.Fatal(err)
	}

	svc, err := etcd.NewService(client, *service)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan sd.Event)
	svc.Subscribe(ch)
	defer svc.Unsubscribe(ch)

	inst := sd.Instance{
		ID:   *instID,
		Addr: *instAddr,
	}
	cleanup, err := svc.Register(inst, *instTTL)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

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
