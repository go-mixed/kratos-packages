package stream

import (
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	"io"
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

// WriteAndClose writes the data to the stream and closes the stream. MUST-run in a separate goroutine.
func (s *Stream) WriteAndClose(data []byte) error {
	_, _ = s.Write(data)
	return s.Close()
}

// Write writes the data to the stream. MUST-run in a separate goroutine.
func (s *Stream) Write(data []byte) (int, error) {
	n, err := s.pipeWriter.Write(data)
	if err != nil {
		s.Close()
	}
	return n, err
}

func (s *Stream) Wait() error {
	return s.ctx.Stream(200, s.contentType, s.pipeReader)
}
