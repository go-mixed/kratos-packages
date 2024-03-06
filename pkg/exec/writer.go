package exec

import (
	"context"
	"io"
	"strings"
)

type detectAndQuitWriter struct {
	originalWriter io.Writer
	buf            []byte
	detectingText  string

	monitoring  bool // 是否还在监控
	blockCtx    context.Context
	blockCancel context.CancelFunc
}

func (w *detectAndQuitWriter) Write(p []byte) (n int, err error) {
	n, err = w.originalWriter.Write(p)
	if err != nil {
		return n, err
	} else if !w.monitoring { // 已经退出了，直接返回
		return n, nil
	}

	select {
	case <-w.blockCtx.Done(): // 如果超时、或者进程退出，直接返回
		w.blockCancel()
		w.buf = nil
		w.monitoring = false // 关闭监控
		return n, nil
	default: // 检测buf中的关键字
		w.buf = append(w.buf, p[:n]...)
		if strings.Contains(string(w.buf), w.detectingText) {
			w.blockCancel()
			w.buf = nil
			w.monitoring = false // 关闭监控
		}
	}

	return n, nil
}

func detectProcessSuccess(writer io.Writer, ctx context.Context, cancel context.CancelFunc, detectingText string) io.Writer {
	if detectingText == "" {
		cancel()
		return writer
	}
	return &detectAndQuitWriter{
		originalWriter: writer,
		buf:            make([]byte, 0),
		monitoring:     true,
		blockCtx:       ctx,
		blockCancel:    cancel,
		detectingText:  detectingText,
	}
}
