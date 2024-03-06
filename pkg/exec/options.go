package exec

import "time"

type processOptions struct {
	blockTimeout  time.Duration
	alwaysBlock   bool
	runTimeout    time.Duration
	detectingText string

	runLimits int
}

type optionFunc func(*process)
type onceOptionFunc = optionFunc
type serverOptionFunc = optionFunc

// BlockUtil 当遇到某个文本时，退出阻塞，用于RunServer
func BlockUtil(detectText string) serverOptionFunc {
	return func(p *process) {
		p.options.alwaysBlock = false
		p.options.detectingText = detectText
	}
}

// BlockTimeout 阻塞超时时间，用于RunServer
func BlockTimeout(timeout time.Duration) serverOptionFunc {
	return func(p *process) {
		p.options.alwaysBlock = false
		p.options.blockTimeout = timeout
	}
}

// RunLimits 进程运行次数，用于RunServer
func RunLimits(limits int) serverOptionFunc {
	return func(p *process) {
		p.options.runLimits = limits
	}
}

func AlwaysBlock() serverOptionFunc {
	return func(p *process) {
		p.options.blockTimeout = 0
		p.options.detectingText = ""
		p.options.alwaysBlock = true
	}
}

// RunTimeout 进程退出超时时间，用于RunOnce
func RunTimeout(timeout time.Duration) onceOptionFunc {
	return func(p *process) {
		p.options.runTimeout = timeout
	}
}
