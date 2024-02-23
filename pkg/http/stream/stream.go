package stream

import (
	"encoding/json"
	"errors"
	protojson "github.com/go-kratos/kratos/v2/encoding/json"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/protobuf/proto"
	"io"
)

type Stream struct {
	ctx         kratosHttp.Context
	contentType string

	pipeReader *io.PipeReader
	pipeWriter *io.PipeWriter
}

// NewStream creates a new stream from kratos http context.
// contentType is the content type of the stream. eg: "text/event-stream"
func NewStream(ctx kratosHttp.Context, contentType string) *Stream {
	pipeReader, pipeWriter := io.Pipe()
	return &Stream{
		ctx:         ctx,
		contentType: contentType,
		pipeReader:  pipeReader,
		pipeWriter:  pipeWriter,
	}
}

// Streaming quickly creates a stream, and calls the callback to write data to the stream.
func Streaming(ctx kratosHttp.Context, contentType string, callback func(s *Stream)) error {
	stream := NewStream(ctx, contentType)
	go func() {
		defer stream.Close()

		callback(stream)
	}()
	return stream.Wait()
}

// Close closes the stream. You MUST call this method when you finish writing to the stream.
func (s *Stream) Close() error {
	err1 := s.pipeWriter.Close()
	err2 := s.pipeReader.Close()

	if err1 != nil {
		return err1
	} else {
		return err2
	}
}

// Write writes the data to the stream.
// MUST-run in a separate goroutine different from Wait's goroutine.
func (s *Stream) Write(data []byte) (int, error) {
	n, err := s.pipeWriter.Write(data)
	if err != nil {
		s.Close()
	}
	return n, err
}

// WriteString writes the string data to the stream.
// MUST-run in a separate goroutine different from Wait's goroutine.
func (s *Stream) WriteString(data string) (int, error) {
	return s.Write([]byte(data))
}

// WriteJson turn the data to json and write it to the stream.
func (s *Stream) WriteJson(data any) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.Write(j)
	return err
}

// WriteProto turn the proto buffer message to json and write it to the stream.
// MUST-run in a separate goroutine different from Wait's goroutine.
func (s *Stream) WriteProto(data proto.Message) error {
	j, err := protojson.MarshalOptions.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.Write(j)
	return err
}

// WriteSse writes the SSE to the stream.
// MUST-run in a separate goroutine different from Wait's goroutine.
// https://www.ruanyifeng.com/blog/2017/05/server-sent_events.html
func (s *Stream) WriteSse(sse Sse) error {
	_, err := s.WriteString(sse.String())
	return err
}

// Wait blocks until the stream is closed. Run it in the main goroutine.
func (s *Stream) Wait() error {
	if err := s.ctx.Stream(200, s.contentType, s.pipeReader); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		return err
	}
	return nil
}
