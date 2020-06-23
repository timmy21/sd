package etcd

import (
	"sync"

	"github.com/timmy21/sd"
)

var (
	_ sd.Service = &Service{}
)

func NewService(client *Client, name string) (*Service, error) {
	s := &Service{
		client:      client,
		name:        name,
		stopCh:      make(chan struct{}),
		subscribers: make(map[chan<- sd.Event]struct{}),
	}
	go s.run()
	return s, nil
}

type Service struct {
	mu          sync.RWMutex
	client      *Client
	name        string
	watchStop   func() error
	stopCh      chan struct{}
	subscribers map[chan<- sd.Event]struct{}
}

func (s *Service) Register(inst sd.Instance, ttl int64) (func(), error) {
	return s.client.Register(s.name, inst, ttl)
}

func (s *Service) Subscribe(ch chan<- sd.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscribers[ch] = struct{}{}
}

func (s *Service) Unsubscribe(ch chan<- sd.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subscribers, ch)
}

func (s *Service) run() {
	ch := make(chan struct{})
	cleanup := s.client.Watch(s.name, ch)
	defer cleanup()
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
			instances, err := s.client.GetInstances(s.name)
			event := sd.Event{
				Instances: instances,
				Err:       err,
			}
			s.broadcast(event)
		case <-s.stopCh:
			return
		}
	}
}

func (s *Service) broadcast(evt sd.Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.subscribers {
		ch <- evt.Copy()
	}
}

func (s *Service) Stop() error {
	close(s.stopCh)
	return nil
}
