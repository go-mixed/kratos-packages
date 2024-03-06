package exec

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type process struct {
	logger *log.Helper

	name       string
	executable string
	workDir    string
	args       []string
	env        map[string]string
	stdout     io.Writer
	stderr     io.Writer

	options processOptions
}

func (p *process) buildOptions(options ...optionFunc) {
	for _, option := range options {
		option(p)
	}
}

func (p *process) makeRunServerContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if p.options.alwaysBlock { // 一直阻塞
		return context.WithCancel(ctx)
	} else if p.options.blockTimeout > 0 { // 等待超时退出
		return context.WithTimeout(ctx, p.options.blockTimeout)
	} else if p.options.detectingText != "" { // 等到某个文本出现时退出
		return context.WithCancel(ctx)
	}

	return ctx, func() {} // 不会阻塞
}

func (p *process) makeRunOnceContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if p.options.runTimeout <= 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, p.options.runTimeout)
}

func (p *process) getName() string {
	if p.name != "" {
		return p.name
	}
	return p.executable
}

func (p *process) getEnv() []string {
	return append(os.Environ(), lo.MapToSlice(p.env, func(key string, value string) string {
		return fmt.Sprintf("%s=%s", key, value)
	})...)
}

func (p *process) getStdout(blockCtx context.Context, blockCancel context.CancelFunc) io.Writer {
	var writer io.Writer = p.stdout
	if p.stdout == nil {
		writer = os.Stdout
	}

	if p.options.detectingText != "" {
		return detectProcessSuccess(writer, blockCtx, blockCancel, p.options.detectingText)
	}

	return writer
}

func (p *process) getStderr(ctx context.Context, cancel context.CancelFunc) io.Writer {
	var writer io.Writer = p.stderr
	if p.stderr == nil {
		writer = os.Stderr
	}

	if p.options.detectingText != "" {
		return detectProcessSuccess(writer, ctx, cancel, p.options.detectingText)
	}

	return writer
}

func (p *process) RunOnce(ctx context.Context, options ...onceOptionFunc) error {
	p.buildOptions(options...)
	p.options.detectingText = ""
	runCtx, runCancel := p.makeRunOnceContext(ctx)
	defer runCancel()

	var pid int

	cmd := exec.CommandContext(runCtx, p.executable, p.args...)
	cmd.Dir = p.workDir
	cmd.Stdout = p.getStdout(runCtx, runCancel)
	cmd.Stderr = p.getStderr(runCtx, runCancel)
	cmd.Env = p.getEnv()

	if err := cmd.Start(); err != nil {
		p.logger.WithContext(ctx).Errorf("[%s]start error: %v", p.getName(), err)
		return err
	}
	pid = cmd.Process.Pid

	if err := cmd.Wait(); err != nil {
		if strings.Contains(err.Error(), "signal: killed") {
			return nil
		}

		p.logger.WithContext(ctx).Errorf("[%s]unexpectedly abort, error: %+v", p.getName(), err)
		return err
	}

	p.logger.WithContext(ctx).Infof("[%s]started successfully, pid = %d\n", p.getName(), pid)
	return nil
}

func (p *process) RunServer(ctx context.Context, options ...serverOptionFunc) {

	p.buildOptions(options...)

	// 如果设置了blockTimeout, 或者detectingText, 则表示在阻塞到某个条件时退出。否则不会阻塞
	blockCtx, blockCancel := p.makeRunServerContext(ctx)
	defer blockCancel()

	var pid int

	go func() {
		defer blockCancel()

		var i int
		for {
			i += 1
			if p.options.runLimits > 0 && i > p.options.runLimits {
				p.logger.WithContext(ctx).Errorf("[%s]exceeds the maximum number of retries", p.getName())
				return
			}

			select {
			case <-ctx.Done():
				return
			default:
				cmd := exec.CommandContext(ctx, p.executable, p.args...)
				cmd.Dir = p.workDir
				cmd.Stdout = p.getStdout(blockCtx, blockCancel)
				cmd.Stderr = p.getStderr(blockCtx, blockCancel)
				cmd.Env = p.getEnv()

				if err := cmd.Start(); err != nil {
					p.logger.WithContext(ctx).Errorf("[%s]start error: %v", p.getName(), err)
					time.Sleep(5 * time.Second)
					continue
				}
				pid = cmd.Process.Pid

				p.logger.WithContext(ctx).Infof("[%s]start, pid = %d", p.getName(), pid)

				if err := cmd.Wait(); err != nil {
					if strings.Contains(err.Error(), "signal: killed") {
						return
					}

					p.logger.WithContext(ctx).Errorf("[%s]unexpectedly abort, pid = %d, retry after 5s, error: %+v", p.getName(), pid, err)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()

	// 如果设置了BlockUntil、或者BlockTimeout, blockCtx != ctx
	if blockCtx != ctx {
		<-blockCtx.Done()
	}

	p.logger.WithContext(ctx).Infof("[%s]started successfully, pid = %d\n", p.getName(), pid)
}
