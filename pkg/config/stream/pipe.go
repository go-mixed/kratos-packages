package stream

type ReaderPipe chan Reader

func NewReaderPipe() ReaderPipe {
	return make(ReaderPipe)
}

func (pipe ReaderPipe) Send(reader Reader) {
	pipe <- reader
}

func (pipe ReaderPipe) Receive() Reader {
	return <-pipe
}
