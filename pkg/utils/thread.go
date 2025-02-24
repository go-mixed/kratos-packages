package utils

import (
	"sync"
	"time"
)

type ThreadPool struct {
	jobs    *List[func()]
	cond    *sync.Cond
	stopped bool // 停止
	paused  bool // 暂停

	maxThreads  int // 最大线程数
	maxLifetime time.Duration
	running     int // 当前正在运行的线程数
}

type threadPoolOption func(t *ThreadPool)

func WithMaxLifetime(maxLifeTime time.Duration) threadPoolOption {
	return func(t *ThreadPool) {
		t.maxLifetime = maxLifeTime
	}
}

// NewThreadPool creates a new thread pool with max thread size .
//
//	maxThreads: the max threads of the thread pool.
//	maxLifetime: the max lifetime of the thread. (default: 10 minutes), Only exit when the thread is idle
func NewThreadPool(maxThreads int, options ...threadPoolOption) *ThreadPool {
	t := &ThreadPool{
		jobs:        NewList[func()](),
		cond:        sync.NewCond(&sync.Mutex{}),
		maxThreads:  maxThreads,
		maxLifetime: 10 * time.Minute,
	}

	for _, option := range options {
		option(t)
	}

	t.Reset()
	return t
}

// worker is the main loop of the thread.
func (t *ThreadPool) worker() {
	ticker := time.NewTicker(t.maxLifetime)
	defer ticker.Stop()

	// lazy remove expire thread
	go func() {
		for {
			select {
			case _, ok := <-ticker.C:
				if !ok { // ticker 已经关闭
					return
				}

				t.cond.L.Lock()
				if t.stopped {
					t.cond.L.Unlock()
					return
				}
				// 唤醒所有消费者，因为可能存在 maxLifetime 超过 maxThreads 的情况
				t.cond.Broadcast()
				t.cond.L.Unlock()
			}
		}
	}()

	for {
		t.cond.L.Lock()
		if t.stopped {
			t.cond.L.Unlock()
			return
		}

		// 如果当前消费者 >= maxThreads || >= 任务队列长度 || 暂停，则等待
		for t.running >= t.maxThreads || t.running >= t.jobs.Len() || t.paused {
			t.cond.Wait()
			// 唤醒之后，检测退出条件
			if t.stopped {
				t.cond.L.Unlock()
				return
			}
		}

		// 启动新的消费者，数量不超过 maxThreads 或 任务队列长度
		// 此举可以保证：不盲目扩展
		for ; t.running < min(t.maxThreads, t.jobs.Len()); t.running++ {
			go t.consuming()
		}
		t.cond.L.Unlock()
	}
}

// consuming consumes jobs.
func (t *ThreadPool) consuming() {
	runAt := time.Now()

	tryExit := func() bool {
		// 如果当前消费者 > maxThreads || 已经停止 || 超过 maxLifetime，则退出
		if t.running > t.maxThreads || t.stopped || time.Since(runAt) > t.maxLifetime {
			t.running--
			t.cond.Broadcast() // 退出时，需要唤醒worker，不然可能会因为 maxLifetime 导致无 consuming 运行

			return true
		}
		return false
	}

	for {
		t.cond.L.Lock()
		// 如果当前消费者 > maxThreads || 停，则退出
		// 此举可以保证 SetMaxThreads 之后，自动缩减 线程数
		if tryExit() {
			t.cond.L.Unlock()
			return
		}

		// 如果任务队列为空 || 暂停，则等待
		for t.jobs.Len() == 0 || t.paused {
			t.cond.Wait()
			// 唤醒之后，如果符合退出条件，则退出
			if tryExit() {
				t.cond.L.Unlock()
				return
			}
		}

		element := t.jobs.PopFront()
		remainingJobs := t.jobs.Len()
		t.cond.L.Unlock()

		if element != nil {
			element.Value()
		}

		// 运行任务完毕 唤醒 t.Join 不然Join不会退出
		if remainingJobs == 0 {
			t.cond.L.Lock()
			t.cond.Broadcast()
			t.cond.L.Unlock()
		}
	}
}

// Len returns the length of the jobs.
func (t *ThreadPool) Len() int {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	return t.jobs.Len()
}

// MaxThreads returns the max threads of the thread pool.
func (t *ThreadPool) MaxThreads() int {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	return t.maxThreads
}

// SetMaxThreads sets the max threads of the thread pool.
func (t *ThreadPool) SetMaxThreads(cap int) {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	t.maxThreads = cap
	t.cond.Broadcast()
}

// RunningCount returns the number of running jobs.
func (t *ThreadPool) RunningCount() int {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	return t.running
}

// Submit submits a job to the thread pool.
func (t *ThreadPool) Submit(jobs ...func()) {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	for _, job := range jobs {
		t.jobs.PushBack(job)
	}
	t.cond.Broadcast()
}

func (t *ThreadPool) SubmitWaitFor(jobs ...func()) {
	if len(jobs) == 0 {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(len(jobs))

	t.cond.L.Lock()

	for _, job := range jobs {
		job := job
		t.jobs.PushBack(func() {
			defer wg.Done()

			job()
		})
	}
	t.cond.Broadcast()
	t.cond.L.Unlock()

	wg.Wait()
}

// Reset resets the thread pool. (the running jobs won't be killed)
func (t *ThreadPool) Reset() {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	t.jobs.Init()
	t.stopped = false
	t.paused = false
	t.running = 0
	go t.worker() // 重新启动worker
	t.cond.Broadcast()
}

// Resume resumes the thread.
func (t *ThreadPool) Resume() {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	t.paused = false
	t.cond.Broadcast()
}

// Pause pauses the thread. (it'll pause after the running job is finished)
func (t *ThreadPool) Pause() {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	t.paused = true
	t.cond.Broadcast()
}

// Stop stops the thread. (it'll stop after the running job is finished)
func (t *ThreadPool) Stop() {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	t.stopped = true
	t.cond.Broadcast()
}

// Join blocks until all jobs are done.
func (t *ThreadPool) Join() {
	t.cond.L.Lock()
	for t.jobs.Len() > 0 {
		t.cond.Wait()
		if t.stopped {
			break
		}
	}

	t.cond.L.Unlock()
}
