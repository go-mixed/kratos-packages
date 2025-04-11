package stream

import (
	"encoding/json"
	protojson "github.com/go-kratos/kratos/v2/encoding/json"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

type StreamWriter struct {
	ctx     kratosHttp.Context
	options options

	// 是否已经发送了header
	sentHeader atomic.Bool
	// 是否已经关闭
	stop       atomic.Bool
	bufferSize atomic.Int32

	// lastSentAt is the time when the last data was sent.
	lastSentAt time.Time
	// lastSentBytes is the number of bytes sent in the last data.
	lastSentBytes int
	// sentBytes is the number of bytes sent in total.
	sentBytes int
}

var _ io.WriteCloser = (*StreamWriter)(nil)
var _ http.Flusher = (*StreamWriter)(nil)

// NewStreamWriter creates a new stream writer from kratos http context.
func NewStreamWriter(ctx kratosHttp.Context, contentType string, opts ...StreamOption) *StreamWriter {
	sw := &StreamWriter{
		ctx: ctx,

		sentHeader: atomic.Bool{},
		stop:       atomic.Bool{},
		bufferSize: atomic.Int32{},

		options: options{
			flushInterval: 0,
			flushSize:     1024,
			contentType:   contentType,
		},
	}

	for _, option := range opts {
		option(&sw.options)
	}

	if sw.options.flushInterval > 0 {
		sw.flushInterval()
	}

	return sw
}

// Streaming quickly creates a stream writer, and calls the callback to write data to the stream.
func Streaming(ctx kratosHttp.Context, contentType string, callback func(s *StreamWriter) error, opts ...StreamOption) error {
	stream := NewStreamWriter(ctx, contentType, opts...)
	// Close the stream writer when the callback returns. It'll stop the second goroutine.
	defer stream.Close()

	if err := callback(stream); err != nil {
		return err
	}

	return nil
}

func (s *StreamWriter) sendHeader() {
	if !s.sentHeader.Swap(true) {
		response := s.ctx.Response()
		if s.options.contentType != "" {
			response.Header().Set("Content-Type", s.options.contentType)
		}
		response.WriteHeader(200)
	}
}

func (s *StreamWriter) Flush() {
	// If the buffer is empty, return directly.
	if s.bufferSize.Load() <= 0 {
		return
	}
	// Flush the buffer and reset the buffer size.
	if flush, ok := s.ctx.Response().(kratosHttp.Flusher); ok {
		flush.Flush()
	}
	s.bufferSize.Store(0)
}

func (s *StreamWriter) flushInterval() {
	time.AfterFunc(s.options.flushInterval, func() {
		if !s.stop.Load() {
			s.Flush()
			// If the stream is not closed, continue to Flush.
			s.flushInterval()
		}
	})
}

// Close closes the stream writer. You MUST call this method when you finish writing to the stream.
func (s *StreamWriter) Close() error {
	s.stop.Store(true)

	return nil
}

// Write writes the data to the stream.
func (s *StreamWriter) Write(data []byte) (int, error) {
	s.sendHeader()

	// If the stream is closed, return EOF.
	if s.stop.Load() {
		return 0, io.EOF
	}

	n, err := s.ctx.Response().Write(data)
	s.sentBytes += n
	s.lastSentBytes = n
	if err != nil {
		return n, err
	}

	if s.bufferSize.Add(int32(n)) >= s.options.flushSize {
		s.Flush()
	}

	s.lastSentAt = time.Now()

	return n, nil
}

// WriteString writes the string data to the stream.
func (s *StreamWriter) WriteString(data string) (int, error) {
	return s.Write([]byte(data))
}

// WriteJson turn the data to json and write it to the stream.
func (s *StreamWriter) WriteJson(data any) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.Write(j)
	// Flush the buffer to send the data immediately.
	s.Flush()
	return err
}

// WriteProto turn the proto buffer message to json and write it to the stream.
func (s *StreamWriter) WriteProto(data proto.Message) error {
	j, err := protojson.MarshalOptions.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.Write(j)
	// Flush the buffer to send the data immediately.
	s.Flush()
	return err
}

// WriteSse writes the SSE to the stream.
// https://www.ruanyifeng.com/blog/2017/05/server-sent_events.html
func (s *StreamWriter) WriteSse(sse Sse) error {
	_, err := s.WriteString(sse.String())
	// Flush the buffer to send the data immediately.
	s.Flush()
	return err
}

// LastSentAt returns the time when the last data was sent.
func (s *StreamWriter) LastSentAt() time.Time {
	return s.lastSentAt
}

// LastSentBytes returns the number of bytes sent in the last data.
func (s *StreamWriter) LastSentBytes() int {
	return s.lastSentBytes
}

// SentBytes returns the number of bytes sent in total.
func (s *StreamWriter) SentBytes() int {
	return s.sentBytes
}
