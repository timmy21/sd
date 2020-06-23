/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/timmy21/sd/etcd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/resolver"
)

const (
	defaultName = "world"
)

func main() {
	var (
		help      = flag.Bool("-help", false, "print usage")
		service   = flag.String("service", "test", "service name")
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
	resolver.Register(etcd.NewBuilder(svc))

	// Set up a connection to the server.
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:///", *service),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithBalancerName(roundrobin.Name),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
