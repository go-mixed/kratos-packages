package stream

import (
	"errors"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	"io"
	"strconv"
	"time"
)

type Stream struct {
	ctx         kratosHttp.Context
	contentType string

	pipeReader *io.PipeReader
	pipeWriter *io.PipeWriter
}

func NewStream(ctx kratosHttp.Context, contentType string) *Stream {
	pipeReader, pipeWriter := io.Pipe()
	return &Stream{
		ctx:         ctx,
		contentType: contentType,
		pipeReader:  pipeReader,
		pipeWriter:  pipeWriter,
	}

}

func (s *Stream) Close() error {
	err1 := s.pipeWriter.Close()
	err2 := s.pipeReader.Close()

	if err1 != nil {
		return err1
	} else {
		return err2
	}
}

// Write writes the data to the stream. MUST-run in a separate goroutine.
func (s *Stream) Write(data []byte) (int, error) {
	n, err := s.pipeWriter.Write(data)
	if err != nil {
		s.Close()
	}
	return n, err
}

func (s *Stream) WriteString(val string) (int, error) {
	return s.Write([]byte(val))
}

func (s *Stream) WriteEvent(event string) error {
	_, err := s.WriteString("event: " + event + "\n")
	return err
}

func (s *Stream) WriteData(data string) error {
	_, err := s.WriteString("data: " + data + "\n\n")
	return err
}

func (s *Stream) WriteId(id string) error {
	_, err := s.WriteString("id: " + id + "\n")
	return err
}

func (s *Stream) WriteRetry(retry time.Duration) error {
	_, err := s.WriteString("id: " + strconv.FormatInt(retry.Milliseconds(), 10) + "\n")
	return err
}

func (s *Stream) WriteComment(comment string) error {
	_, err := s.WriteString(": " + comment + "\n")
	return err
}

// WriteAndClose writes the data to the stream and closes the stream. MUST-run in a separate goroutine.
func (s *Stream) WriteAndClose(data []byte) error {
	_, _ = s.Write(data)
	return s.Close()
}

// WriteStringAndClose writes the data to the stream and closes the stream. MUST-run in a separate goroutine.
func (s *Stream) WriteStringAndClose(data string) error {
	_, _ = s.WriteString(data)
	return s.Close()
}

func (s *Stream) Wait() error {
	if err := s.ctx.Stream(200, s.contentType, s.pipeReader); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		return err
	}
	return nil
}
