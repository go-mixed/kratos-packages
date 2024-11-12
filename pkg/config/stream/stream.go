package stream

import (
	"io"

	"github.com/go-kratos/kratos/v2/config"
)

type IStream interface {
	Load() ([]*config.KeyValue, error)
	Watch() (config.Watcher, error)
}

type stream struct {
	reader Reader
	pipe   ReaderPipe
}

func NewSource(reader Reader, pipe ReaderPipe) config.Source {
	return &stream{
		reader: reader,
		pipe:   pipe,
	}
}

func (r *stream) Load() ([]*config.KeyValue, error) {
	return r.loadStream(r.reader)
}

func (r *stream) loadStream(reader Reader) ([]*config.KeyValue, error) {
	content, err := io.ReadAll(reader.GetStream())
	if err != nil {
		return nil, err
	}
	kv := &config.KeyValue{
		Key:    "stream",
		Format: reader.Format(),
		Value:  content,
	}
	return []*config.KeyValue{kv}, nil
}

func (r *stream) Watch() (config.Watcher, error) {
	return newWatcher(r)
}
