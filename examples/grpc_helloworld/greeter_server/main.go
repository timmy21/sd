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

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"strings"

	"github.com/timmy21/sd"
	"github.com/timmy21/sd/etcd"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	var (
		help      = flag.Bool("-help", false, "print usage")
		service   = flag.String("service", "test", "service name")
		addr      = flag.String("address", "127.0.0.1:50051", "listen address")
		endpoints = flag.String("etcd-endpoints", "127.0.0.1:2379", "etcd endpoints")
		instTTL   = flag.Int64("instance-ttl", 10, "service register ttl (seconds)")
		prefix    = flag.String("etcd-key-prefix", "discovery", "etcd discovery key prefix")
	)
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// register server
	ctx := context.Background()
	client, err := etcd.NewClient(ctx, strings.Split(*endpoints, ","), *prefix)
	if err != nil {
		log.Fatal(err)
	}
	svc, err := etcd.NewService(client, *service)
	if err != nil {
		log.Fatal(err)
	}
	inst := sd.Instance{
		ID:   *addr,
		Addr: *addr,
	}
	err = svc.Register(inst, *instTTL)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
