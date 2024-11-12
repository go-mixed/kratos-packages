package stream

import (
	"context"

	"github.com/go-kratos/kratos/v2/config"
)

type watcher struct {
	stream *stream

	ctx    context.Context
	cancel context.CancelFunc
}

func newWatcher(stream *stream) (config.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &watcher{
		ctx:    ctx,
		cancel: cancel,
		stream: stream,
	}, nil
}

func (w *watcher) Next() ([]*config.KeyValue, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case reader := <-w.stream.pipe:
		return w.stream.loadStream(reader)
	}
}

func (w *watcher) Stop() error {
	w.cancel()
	close(w.stream.pipe)
	return nil
}
