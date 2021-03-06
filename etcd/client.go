package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/timmy21/sd"
	"go.etcd.io/etcd/clientv3"
)

type Client struct {
	prefix string
	ctx    context.Context
	cli    *clientv3.Client
}

func NewClient(ctx context.Context, endpoints []string, prefix string, options ...ClientOption) (*Client, error) {
	opts := defaultOptions
	for _, opt := range options {
		opt(&opts)
	}
	cfg := clientv3.Config{
		Endpoints:            endpoints,
		Context:              ctx,
		DialTimeout:          opts.DialTimeout,
		DialKeepAliveTime:    opts.DialKeepAliveTime,
		DialKeepAliveTimeout: opts.DialKeepAliveTimeout,
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		prefix: prefix,
		ctx:    ctx,
		cli:    cli,
	}, err
}

func (c *Client) Register(svc string, inst sd.Instance, ttl int64) (func(), error) {
	val, err := json.Marshal(inst)
	if err != nil {
		return nil, nil
	}
	key := c.makeKey(svc, inst)
	resp, err := c.cli.Grant(c.ctx, ttl)
	if err != nil {
		return nil, err
	}
	leaseID := resp.ID
	_, err = c.cli.Put(c.ctx, key, string(val), clientv3.WithLease(leaseID))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(c.ctx)
	kch, err := c.cli.KeepAlive(ctx, leaseID)
	if err != nil {
		cancel()
		return nil, err
	}
	go func() {
		for {
			select {
			case _, ok := <-kch:
				if !ok {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return func() {
		c.cli.Revoke(c.ctx, leaseID)
		cancel()
	}, nil
}

func (c *Client) Watch(svc string, ch chan<- struct{}) func() {
	ctx, cancel := context.WithCancel(c.ctx)
	wch := c.cli.Watch(c.ctx, c.servicePrefix(svc), clientv3.WithPrefix())
	go func() {
		for {
			select {
			case _, ok := <-wch:
				if !ok {
					return
				}
				ch <- struct{}{}
			case <-ctx.Done():
				return
			}
		}
	}()
	return func() {
		cancel()
	}
}

func (c *Client) GetInstances(svc string) ([]sd.Instance, error) {
	resp, err := c.cli.Get(c.ctx, c.servicePrefix(svc), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	result := make([]sd.Instance, 0)
	for _, kv := range resp.Kvs {
		var inst sd.Instance
		err := json.Unmarshal(kv.Value, &inst)
		if err != nil {
			return nil, err
		}
		result = append(result, inst)
	}
	return result, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) makeKey(svc string, inst sd.Instance) string {
	return c.servicePrefix(svc) + inst.ID
}

func (c *Client) servicePrefix(svc string) string {
	return fmt.Sprintf("%s/%s/", c.prefix, svc)
}
