package stream

import (
	"context"
	"encoding/json"
	protojson "github.com/go-kratos/kratos/v2/encoding/json"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"io"
)

type StreamWriter struct {
	ctx         kratosHttp.Context
	contentType string

	ioChanel   chan []byte
	quitCtx    context.Context
	quitCancel context.CancelCauseFunc
}

// NewStreamWriter creates a new stream writer from kratos http context.
// contentType is the content type of the stream. eg: "text/event-stream"
func NewStreamWriter(ctx kratosHttp.Context, contentType string) *StreamWriter {
	quitCtx, quitCancel := context.WithCancelCause(ctx)
	return &StreamWriter{
		ctx:         ctx,
		contentType: contentType,
		ioChanel:    make(chan []byte),
		quitCtx:     quitCtx,
		quitCancel:  quitCancel,
	}
}

// Streaming quickly creates a stream writer, and calls the callback to write data to the stream.
func Streaming(ctx kratosHttp.Context, contentType string, callback func(s *StreamWriter) error) error {
	stream := NewStreamWriter(ctx, contentType)
	go func() {
		// Close the stream writer when the callback returns. It'll stop the second goroutine.
		defer stream.Close()

		if err := callback(stream); err != nil {
			stream.quitCancel(err)
			return
		}
	}()
	return stream.Wait()
}

// Close closes the stream writer. You MUST call this method when you finish writing to the stream.
func (s *StreamWriter) Close() error {
	select {
	case <-s.quitCtx.Done(): // 已经退出了
	default:
		s.quitCancel(nil) // 无错退出
	}

	return nil
}

// Write writes the data to the stream.
func (s *StreamWriter) Write(data []byte) (int, error) {
	select {
	case s.ioChanel <- data:
	case <-s.quitCtx.Done(): // 已经退出了
		return 0, io.EOF
	}
	return len(data), nil
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
	return err
}

// WriteProto turn the proto buffer message to json and write it to the stream.
func (s *StreamWriter) WriteProto(data proto.Message) error {
	j, err := protojson.MarshalOptions.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.Write(j)
	return err
}

// WriteSse writes the SSE to the stream.
// https://www.ruanyifeng.com/blog/2017/05/server-sent_events.html
func (s *StreamWriter) WriteSse(sse Sse) error {
	_, err := s.WriteString(sse.String())
	return err
}

// Wait blocks until the stream.pipeReader is closed. Run it in the main goroutine.
func (s *StreamWriter) Wait() error {
	response := s.ctx.Response()
	response.Header().Set("Content-Type", s.contentType)
	response.WriteHeader(200)

for1:
	for {
		select {
		case data := <-s.ioChanel:
			if _, err := s.ctx.Response().Write(data); err != nil {
				s.quitCancel(err)
				break for1
			}

			if flush, ok := response.(kratosHttp.Flusher); ok {
				flush.Flush()
			}
		case <-s.quitCtx.Done(): // 已经退出了
			break for1
		}
	}

	if err := s.quitCtx.Err(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
