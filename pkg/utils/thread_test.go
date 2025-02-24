package utils

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestThreadPool_BasicFunctionality(t *testing.T) {
	pool := NewThreadPool(2)
	defer pool.Stop()

	var counter int32
	const tasks = 10

	// 提交10个任务
	for i := 0; i < tasks; i++ {
		pool.Submit(func() {
			atomic.AddInt32(&counter, 1)
		})
	}

	// 等待所有任务完成
	pool.Join()

	if atomic.LoadInt32(&counter) != tasks {
		t.Errorf("Expected %d tasks executed, got %d", tasks, counter)
	}
}

func TestThreadPool_MaxThreads(t *testing.T) {
	pool := NewThreadPool(2)
	defer pool.Stop()

	var (
		maxConcurrent int32
		current       int32
		wg            sync.WaitGroup
	)

	const tasks = 5

	for i := 0; i < tasks; i++ {
		wg.Add(1)
		pool.Submit(func() {
			defer wg.Done()
			curr := atomic.AddInt32(&current, 1)
			defer atomic.AddInt32(&current, -1)

			if curr > atomic.LoadInt32(&maxConcurrent) {
				atomic.StoreInt32(&maxConcurrent, curr)
			}
			time.Sleep(100 * time.Millisecond) // 模拟耗时任务
		})
	}

	wg.Wait()

	if maxConcurrent > 2 {
		t.Errorf("Max concurrent goroutines should be <= 2, got %d", maxConcurrent)
	}
}

func TestThreadPool_PauseResume(t *testing.T) {
	pool := NewThreadPool(1)
	defer pool.Stop()

	var counter int32

	// 先暂停
	pool.Pause()

	// 提交任务
	pool.Submit(func() {
		atomic.StoreInt32(&counter, 1)
	})

	// 等待一段时间验证任务未执行
	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt32(&counter) != 0 {
		t.Error("Task executed when pool was paused")
	}

	// 恢复执行
	pool.Resume()
	pool.Join()

	if atomic.LoadInt32(&counter) != 1 {
		t.Error("Task not executed after resume")
	}
}

func TestThreadPool_SubmitWaitFor(t *testing.T) {
	pool := NewThreadPool(2)
	defer pool.Stop()

	var counter int32

	start := time.Now()
	pool.SubmitWaitFor(func() {
		time.Sleep(500 * time.Millisecond)
		atomic.StoreInt32(&counter, 1)
	})

	duration := time.Since(start)

	if duration < 500*time.Millisecond {
		t.Error("SubmitWaitFor didn't block properly")
	}

	if atomic.LoadInt32(&counter) != 1 {
		t.Error("Task not executed in SubmitWaitFor")
	}
}

func TestThreadPool_Stop(t *testing.T) {
	pool := NewThreadPool(1)

	var counter int32
	pool.Submit(func() {
		time.Sleep(100 * time.Millisecond)
		atomic.StoreInt32(&counter, 1)
	})
	// 这里需要等待运行结束，不然stop会优先于job
	pool.Join()
	pool.Stop()

	// 尝试提交新任务
	pool.Submit(func() {
		t.Error("[TestThreadPool_Stop]New task executed after Stop()")
		atomic.StoreInt32(&counter, 123)
	})

	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&counter) != 1 {
		t.Errorf("New task executed after Stop(): %d", counter)
	}
}

// 测试动态调整最大线程数
func TestThreadPool_DynamicMaxThreads(t *testing.T) {
	pool := NewThreadPool(2)
	defer pool.Stop()

	var (
		maxConcurrent int32
	)

	// 初始阶段验证2并发
	const phase1Tasks = 4
	for i := 0; i < phase1Tasks; i++ {
		pool.Submit(func() {
			current := atomic.AddInt32(&maxConcurrent, 1)
			defer atomic.AddInt32(&maxConcurrent, -1)

			if current > 2 {
				t.Error("Concurrent workers exceeded initial max threads")
			}
			time.Sleep(200 * time.Millisecond)
		})
	}

	pool.Join()

	// 动态调整到4个线程
	pool.SetMaxThreads(4)

	// 第二阶段验证4并发
	const phase2Tasks = 8
	start := time.Now()
	for i := 0; i < phase2Tasks; i++ {
		pool.Submit(func() {
			current := atomic.AddInt32(&maxConcurrent, 1)
			defer atomic.AddInt32(&maxConcurrent, -1)

			if current > 4 {
				t.Error("Concurrent workers exceeded adjusted max threads")
			}
			time.Sleep(200 * time.Millisecond)
		})
	}

	pool.Join()

	// 验证调整后任务完成时间（应比固定2线程快）
	if duration := time.Since(start); duration > 650*time.Millisecond {
		t.Errorf("Expected parallel execution with 4 threads, took %v", duration)
	}
}

// 测试最大生存时间功能
func TestThreadPool_MaxLifetime(t *testing.T) {
	pool := NewThreadPool(5, WithMaxLifetime(100*time.Millisecond))
	defer pool.Stop()

	// 首次提交任务启动工作线程
	var ran atomic.Int32
	pool.Submit(func() {
		ran.Store(1)
	})
	pool.Join()

	// 等待超过生存时间
	time.Sleep(300 * time.Millisecond)

	// 检查工作线程是否已回收
	if running := pool.RunningCount(); running != 0 {
		t.Errorf("Expected workers to expire, but got %d running", running)
	}

	// 验证线程池仍然可用
	pool.Submit(func() {
		ran.Store(2)
	})

	pool.Join()

	if ran.Load() != 2 {
		t.Error("Thread pool failed to restart workers after expiration")
	}
}
