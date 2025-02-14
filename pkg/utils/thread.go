package utils

import "sync"

type ThreadPool struct {
	jobs    *List[func()]
	cond    *sync.Cond
	stopped bool // 停止
	paused  bool // 暂停

	maxThreads int // 最大线程数
	running    int // 当前正在运行的线程数
}

// NewThreadPool creates a new thread pool with max thread size .
func NewThreadPool(maxThreads int) *ThreadPool {
	t := &ThreadPool{
		jobs:       NewList[func()](),
		cond:       sync.NewCond(&sync.Mutex{}),
		maxThreads: maxThreads,
	}

	t.Reset()
	return t
}

// worker is the main loop of the thread.
func (t *ThreadPool) worker() {
	for {
		t.cond.L.Lock()
		if t.stopped {
			t.cond.L.Unlock()
			return
		}

		// 如果当前消费者 >= maxThreads || >= 任务队列长度 || 暂停，则等待
		for t.running >= t.maxThreads || t.running >= t.jobs.Len() || t.paused {
			t.cond.Wait()
		}

		// 启动新的消费者
		for i := 0; i < min(t.maxThreads, t.jobs.Len())-t.running; i++ {
			go t.consuming(t.running)
			t.running++
		}
		t.cond.L.Unlock()
	}
}

// consuming consumes jobs.
func (t *ThreadPool) consuming(index int) {
	for {
		t.cond.L.Lock()
		// 如果当前消费者 > maxThreads || 停，则退出
		if index >= t.maxThreads || t.stopped {
			t.running--
			t.cond.L.Unlock()
			return
		}

		// 如果任务队列为空 || 暂停，则等待
		for t.jobs.Len() == 0 || t.paused {
			t.cond.Wait()
		}

		element := t.jobs.PopFront()
		t.cond.L.Unlock()

		if element != nil {
			element.Value()
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
	t.cond.L.Unlock()
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
	}
	t.cond.L.Unlock()
}
