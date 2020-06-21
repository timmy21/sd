package etcd

import "time"

var (
	defaultOptions = ClientOptions{
		DialTimeout:       5 * time.Second,
		DialKeepAliveTime: 5 * time.Second,
	}
)

type ClientOptions struct {
	DialTimeout          time.Duration
	DialKeepAliveTime    time.Duration
	DialKeepAliveTimeout time.Duration
}

type ClientOption func(*ClientOptions)

func WithDailTimeout(v time.Duration) ClientOption {
	return func(opts *ClientOptions) {
		opts.DialTimeout = v
	}
}

func WithDailKeepAliveTime(v time.Duration) ClientOption {
	return func(opts *ClientOptions) {
		opts.DialKeepAliveTime = v
	}
}

func WithDialKeepAliveTimeout(v time.Duration) ClientOption {
	return func(opts *ClientOptions) {
		opts.DialKeepAliveTimeout = v
	}
}
