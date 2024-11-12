package config

import (
	"github.com/go-kratos/kratos/v2/config"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/config/stream"
)

type Option func(*configure)

type ApolloOption struct {
	AppID     string
	Cluster   string
	Endpoint  string
	Namespace string
	Secret    string
}

type StreamOption struct {
	Reader stream.Reader
	Pipe   stream.ReaderPipe
}

func WithDriver(driver string) Option {
	return func(w *configure) {
		w.driver = driver
	}
}

func WithPath(paths ...string) Option {
	return func(w *configure) {
		w.paths = paths
	}
}

func WithApolloOption(opts ...ApolloOption) Option {
	return func(w *configure) {
		if len(opts) > 0 {
			w.apolloOpts = opts[0]
		}
	}
}

func WithStreamOption(opts ...StreamOption) Option {
	return func(w *configure) {
		if len(opts) > 0 {
			w.streamOpts = opts[0]
		}
	}
}

func WithSource(sources ...config.Source) Option {
	return func(w *configure) {
		w.sources = append(w.sources, sources...)
	}
}

func WithOption(opts ...config.Option) Option {
	return func(w *configure) {
		w.nativeOpts = append(w.nativeOpts, opts...)
	}
}
