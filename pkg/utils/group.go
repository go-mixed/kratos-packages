package utils

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
)

var ErrWaitGroupStopped = errors.New("wait group stopped")

// ShortCircuitWaitGroup 只要有一个错误出现，就会退出Wait，运行中的任务需要自行监听ctx来退出，排队的任务会跳过
type ShortCircuitWaitGroup struct {
	wg sync.WaitGroup

	running atomic.Int32
	ctx     context.Context
	cancel  context.CancelCauseFunc
	sem     chan struct{}
}

// NewShortCircuitWaitGroup creates a new ShortCircuitWaitGroup
//
//	注意：1个任务失败，则Wait退出，正在运行的任务需要自行监听ctx来退出，排队的任务会跳过
func NewShortCircuitWaitGroup() *ShortCircuitWaitGroup {
	wg, _ := NewShortCircuitWaitGroupWithContext(context.Background())
	return wg
}

// NewShortCircuitWaitGroupWithContext creates a new ShortCircuitWaitGroup with the given context.
//
//	注意：1个任务失败，则Wait退出，正在运行的任务需要自行监听ctx来退出，排队的任务会跳过
func NewShortCircuitWaitGroupWithContext(ctx context.Context) (*ShortCircuitWaitGroup, context.Context) {
	wg := &ShortCircuitWaitGroup{}
	wg.ctx, wg.cancel = context.WithCancelCause(ctx)
	return wg, wg.ctx
}

func (g *ShortCircuitWaitGroup) add(block bool) bool {
	if g.IsStopped() {
		return false
	}

	if g.sem != nil {
		if block {
			select {
			case g.sem <- struct{}{}:
			}
		} else {
			select {
			case g.sem <- struct{}{}:
			default:
				return false
			}
		}
	}

	g.wg.Add(1)
	return g.running.Add(1) > 0
}

func (g *ShortCircuitWaitGroup) done() {
	if val := g.running.Add(-1); val >= 0 {
		g.wg.Done()
		if g.sem != nil {
			select {
			case <-g.sem:
			default:

			}
		}
	}
}

// Stop stops the group's execution
//
//	注意：排队的任务会跳过
func (g *ShortCircuitWaitGroup) Stop(err error) {
	if old := g.running.Swap(-1); old > 0 {
		for j := int32(0); j < old; j++ {
			g.done()
		}

		g.cancel(err)
	}
}

// IsStopped returns true if the group has been stopped.
func (g *ShortCircuitWaitGroup) IsStopped() bool {
	return g.running.Load() < 0
}

// SetLimit limits the number of active goroutines in this group to at most n.
// A negative value indicates no limit.
//
// Any subsequent call to the Go method will block until it can add an active
// goroutine without exceeding the configured limit.
//
// The limit must not be modified while any goroutines in the group are active.
func (g *ShortCircuitWaitGroup) SetLimit(n int) {
	if n < 0 {
		g.sem = nil
		return
	}

	if len(g.sem) != 0 {
		panic(fmt.Errorf("errgroup: modify limit while %v goroutines in the group are still active", len(g.sem)))
	}
	g.sem = make(chan struct{}, n)
}

// Go calls the given function in a new goroutine.
// It blocks until the new goroutine can be added without the number of
// active goroutines in the group exceeding the configured limit.
//
// The first call to return a non-nil error cancels the group's context, if the
// group was created by calling WithContext. The error will be returned by Wait.
func (g *ShortCircuitWaitGroup) Go(fn func() error) {
	// 队列满，则阻塞，如果已经停止，则不执行并返回
	if !g.add(true) {
		return
	}
	g._go(fn)
}

// TryGo calls the given function in a new goroutine only if the number of
// active goroutines in the group is currently below the configured limit.
//
// The return value reports whether the goroutine was started.
func (g *ShortCircuitWaitGroup) TryGo(fn func() error) bool {
	// 队列满或已经停止，则不执行并返回false
	if !g.add(false) {
		return false
	}
	g._go(fn)
	return true
}

func (g *ShortCircuitWaitGroup) _go(fn func() error) {
	go func() {
		defer g.done()

		if err := fn(); err != nil {
			g.Stop(err)
		}
	}()
}

func (g *ShortCircuitWaitGroup) Wait() error {
	if g.IsStopped() {
		return ErrWaitGroupStopped
	}

	g.wg.Wait()

	err := context.Cause(g.ctx)
	// context没有被取消, err 会是 nil
	if err == nil {
		// 避免 memory leak
		g.cancel(nil)
	}

	return err
}
