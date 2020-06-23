package etcd

import (
	"context"
	"sync"
	"time"

	"github.com/timmy21/sd"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

var (
	minInitRate = 10 * time.Second
)

func NewBuilder(svc *Service) resolver.Builder {
	return &etcdBuilder{
		svc: svc,
	}
}

type etcdBuilder struct {
	svc *Service
}

func (b *etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(context.Background())
	e := &etcdResolver{
		svc:    b.svc,
		ctx:    ctx,
		cancel: cancel,
		cc:     cc,
	}
	e.wg.Add(1)
	go e.watcher()
	return e, nil
}

func (b *etcdBuilder) Scheme() string {
	return b.svc.name
}

type etcdResolver struct {
	svc    *Service
	ctx    context.Context
	cancel context.CancelFunc
	cc     resolver.ClientConn
	wg     sync.WaitGroup
}

func (e *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (e *etcdResolver) Close() {
	e.cancel()
	e.wg.Wait()
}

func (e *etcdResolver) watcher() {
	defer e.wg.Done()

	// initial
	for {
		instances, err := e.svc.client.GetInstances(e.svc.name)
		if err == nil {
			e.UpdateState(instances)
			break
		}
		if err != nil {
			grpclog.Infof("grpc: failed to get etcd instances, %+v\n", err)
		}
		t := time.NewTimer(minInitRate)
		select {
		case <-t.C:
		case <-e.ctx.Done():
			t.Stop()
			return
		}
	}

	// subscribe events
	ch := make(chan sd.Event)
	e.svc.Subscribe(ch)
	defer e.svc.Unsubscribe(ch)

	for {
		select {
		case <-e.ctx.Done():
			return
		case event := <-ch:
			if event.Err != nil {
				grpclog.Infof("grpc: failed to get etcd event, %+v\n", event.Err)
				continue
			}
			e.UpdateState(event.Instances)
		}
	}
}

func (e *etcdResolver) UpdateState(instances []sd.Instance) {
	addresses := make([]resolver.Address, len(instances))
	for i, inst := range instances {
		addresses[i] = resolver.Address{
			Addr: inst.Addr,
		}
	}
	state := resolver.State{
		Addresses: addresses,
	}
	e.cc.UpdateState(state)
}
