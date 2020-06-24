# sd
SD automatically registers and deregisters services. SD supports pluggable service registries, which currently includes etcd.

# Usage

## register service
```go
package main

import (
	"context"
	"log"

	"github.com/timmy21/sd"
	"github.com/timmy21/sd/etcd"
)

func main() {
	keyPrefix := "discovery"
	endpoints := []string{"127.0.0.1:2379"}
	ctx := context.Background()
	client, err := etcd.NewClient(ctx, endpoints, keyPrefix)
	if err != nil {
		log.Fatal(err)
	}
	service := "test"
	svc, err := etcd.NewService(client, service)
	if err != nil {
		log.Fatal(err)
	}
	inst := sd.Instance{
		ID:   "127.0.0.1:8000",
		Addr: "127.0.0.1:8000",
	}
	cleanup, err := svc.Register(inst, 60)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()
}

```

## discovery service
```go
package main

import (
	"context"
	"log"

	"github.com/timmy21/sd"
	"github.com/timmy21/sd/etcd"
)

func main() {
	keyPrefix := "discovery"
	endpoints := []string{"127.0.0.1:2379"}
	ctx := context.Background()
	client, err := etcd.NewClient(ctx, endpoints, keyPrefix)
	if err != nil {
		log.Fatal(err)
	}
	service := "test"
	svc, err := etcd.NewService(client, service)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan sd.Event)
	svc.Subscribe(ch)
	defer svc.Unsubscribe(ch)

	for evt := range ch {
		if evt.Err != nil {
			log.Fatal(evt.Err)
		}
		log.Printf("instances: %v\n", evt.Instances)
	}
}
```
