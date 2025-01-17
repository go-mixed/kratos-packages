package future

import (
	"context"
	"errors"
	"github.com/samber/lo"
)

var ErrPromiseArgument = errors.New("promise argument typing error")

// Job 带ctx的Promise任务函数，用于 PromiseEx PromiseExAll PromiseExAny，可以在其它任务完成/报错的情况下，结束当前任务
type Job[P any, R any] func(context.Context, P) (R, error)

// All 并发运行所有任务，并返回所有任务的结果，传递给下一个Then的参数是[]R，
//
//	注意：如果1个任务报错，会立即调用reject，并且ctx会Done()，其它任务需要处理ctx来退出
//	Example: PromiseAll(job1, job2)).Finally(...)
func All[P any, R any](ctx context.Context, fns ...Job[P, R]) *FutureContext {
	arg, ok := FromContext[P](ctx)
	if !ok {
		return ErrorFuture(ErrPromiseArgument)
	}
	return NewFutureContext(ctx, nil).ThenAllEx(lo.Map(fns, func(fn Job[P, R], index int) func(ctx context.Context) (context.Context, error) {
		return func(ctx context.Context) (context.Context, error) {
			res, err := fn(ctx, arg)
			if err != nil {
				return ctx, err
			}
			return WithContext(ctx, res), nil
		}
	})...)
	//return NewFutureContext(ctx, func(resolve func(context.Context), reject func(error)) {
	//
	//	results := make([]R, len(fns))
	//	wg, innerCtx := utils.NewShortCircuitWaitGroupWithContext(ctx)
	//
	//	for i, fn := range fns {
	//		i := i
	//		fn := fn
	//		wg.Go(func() error {
	//			var innerErr error
	//			results[i], innerErr = fn(innerCtx, arg)
	//
	//			return innerErr
	//		})
	//	}
	//
	//	err := wg.Wait()
	//
	//	if err != nil {
	//		reject(err)
	//	} else {
	//		resolve(WithContext(ctx, results))
	//	}
	//})
}

// Any 运行其中一个任务成功就返回
//
//	注意：如果1个任务完成/报错，ctx会Done()，其它任务需要处理ctx来退出
//	Example: NewPromise().Then(PromiseAny(job1, job2)).Then(...)
func Any[P any, R any](ctx context.Context, fns ...Job[P, R]) *FutureContext {
	arg, ok := FromContext[P](ctx)
	if !ok {
		return ErrorFuture(ErrPromiseArgument)
	}
	return NewFutureContext(ctx, nil).ThenAnyEx(lo.Map(fns, func(fn Job[P, R], index int) func(ctx context.Context) (context.Context, error) {
		return func(ctx context.Context) (context.Context, error) {
			res, err := fn(ctx, arg)
			if err != nil {
				return ctx, err
			}
			return WithContext(ctx, res), nil
		}
	})...)
	//return mo.NewFuture(func(resolve func(context.Context), reject func(error)) {
	//	arg, ok := FromContext[P](ctx)
	//	if !ok {
	//		reject(ErrPromiseArgument)
	//		return
	//	}
	//
	//	results := make([]R, len(fns))
	//	firstIndex := atomic.Int32{} // 谁是第一个返回的任务
	//	firstIndex.Store(-1)
	//	wg, innerCtx := utils.NewShortCircuitWaitGroupWithContext(ctx)
	//
	//	for i, fn := range fns {
	//		i := i
	//		fn := fn
	//		wg.Go(func() error {
	//			var err error
	//			results[i], err = fn(innerCtx, arg)
	//			firstIndex.CompareAndSwap(-1, int32(i)) // 首个任务完成，设置firstIndex
	//			if err != nil {
	//				return err
	//			}
	//			return ErrAnyPromiseDone // 任意1个任务完成，则退出其它任务
	//		})
	//	}
	//
	//	err := wg.Wait()
	//
	//	if err != nil && !errors.Is(err, ErrAnyPromiseDone) {
	//		reject(err)
	//	} else {
	//		var result R
	//		if firstIndex.Load() >= 0 {
	//			result = results[firstIndex.Load()]
	//		}
	//		resolve(WithContext(ctx, result))
	//	}
	//})
}

// Chain 串行的Promise任务，前面的任务不报错，后面的任务才会执行
//
//	Example: Promise(job1, job2).Finally()
func Chain[P any, R any](ctx context.Context, fns ...Job[P, R]) *FutureContext {
	arg, ok := FromContext[P](ctx)
	if !ok {
		return ErrorFuture(ErrPromiseArgument)
	}
	return NewFutureContext(ctx, nil).ThenEx(lo.Map(fns, func(fn Job[P, R], index int) func(ctx context.Context) (context.Context, error) {
		return func(ctx context.Context) (context.Context, error) {
			res, err := fn(ctx, arg)
			if err != nil {
				return ctx, err
			}
			return WithContext(ctx, res), nil
		}
	})...)
	//f := mo.NewFuture(func(resolve func(context.Context), reject func(error)) {
	//	resolve(ctx)
	//})
	//
	//for _, fn := range fns {
	//	fn := fn // 其实 golang 1.22 已经修复了循环中参数的引用问题
	//	f = f.Then(func(ctx context.Context) (context.Context, error) {
	//		arg, ok := FromContext[P](ctx)
	//		if !ok {
	//			return ctx, ErrPromiseArgument
	//		}
	//
	//		res, err := fn(ctx, arg)
	//		if err != nil { // 以fn的错误优先
	//			return ctx, err
	//		} else if err = context.Cause(ctx); err != nil { // 判断是否是ctx cancel导致的退出
	//			return ctx, err
	//		}
	//		return WithContext(ctx, res), nil
	//	})
	//}
	//
	//return f
}
