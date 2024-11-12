package stream

import "io"

type Reader interface {
	GetStream() io.Reader
	Format() string
}

type ReaderImpl struct {
	src    io.Reader
	format string
}

func NewReader(src io.Reader, format string) Reader {
	return &ReaderImpl{
		src:    src,
		format: format,
	}
}

func (sr *ReaderImpl) GetStream() io.Reader {
	return sr.src
}

func (sr *ReaderImpl) Format() string {
	return sr.format
}
