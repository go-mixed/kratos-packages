package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// ListenStopSignal 监听停止Ctrl+C信号
func ListenStopSignal(parentCtx context.Context, cancelFunc context.CancelFunc) {

	go func() {
		exitSign := make(chan os.Signal)
		//监听指定信号: 终端断开, ctrl+c, kill, ctrl+/
		signal.Notify(exitSign, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		defer close(exitSign)

		select {
		case <-parentCtx.Done():
			// 正常退出协程
		case <-exitSign:
			cancelFunc()
		}
	}()

}

func WithStopSignalContext(parentCtx context.Context) (context.Context, context.CancelFunc) {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(parentCtx)

	ListenStopSignal(ctx, cancel)

	return ctx, cancel
}
