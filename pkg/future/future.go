package future

import (
	"context"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"sync"
)

type FutureContext struct {
	future *mo.Future[context.Context]
	ctx    context.Context
}

func ErrorFutureContext(ctx context.Context, err error) *FutureContext {
	return NewFutureContext(ctx, func(resolve func(context.Context), reject func(error)) {
		reject(err)
	})
}

func NewFutureContext(ctx context.Context, cb func(resolve func(context.Context), reject func(error))) *FutureContext {
	if cb == nil {
		cb = func(resolve func(context.Context), reject func(error)) {
			resolve(ctx)
		}
	}
	return &FutureContext{
		future: mo.NewFuture[context.Context](cb),
		ctx:    ctx,
	}
}

// Then invoke callbacks by chain. It returns a new Future.
func (f *FutureContext) Then(cbs ...func() error) *FutureContext {
	future := f.future
	for _, cb := range cbs {
		future = future.Then(func(ctx context.Context) (context.Context, error) {
			return ctx, cb()
		})
	}
	return &FutureContext{future: future, ctx: f.ctx}
}

// ThenEx invoke callbacks by chain. It returns a new Future.
func (f *FutureContext) ThenEx(cbs ...func(context.Context) (context.Context, error)) *FutureContext {
	future := f.future
	for _, cb := range cbs {
		future = future.Then(cb)
	}
	return &FutureContext{future: future, ctx: f.ctx}
}

// ThenFuture invoke the Future of futureFn. It returns a new Future.
func (f *FutureContext) ThenFuture(futureFn func(ctx context.Context) *FutureContext) *FutureContext {
	future := f.future.Then(func(ctx context.Context) (context.Context, error) {
		_f := futureFn(ctx)
		return _f.Collect()
	})
	return &FutureContext{future: future, ctx: f.ctx}
}

// ThenAny return if any of the callbacks succeed. It returns a new Future.
func (f *FutureContext) ThenAny(cbs ...func() error) *FutureContext {
	return f.ThenAnyEx(lo.Map(cbs, func(cb func() error, index int) func(context.Context) (context.Context, error) {
		return func(ctx context.Context) (context.Context, error) {
			return ctx, cb()
		}
	})...)
}

// ThenAnyEx return if any of the callbacks succeed. It returns a new Future.
func (f *FutureContext) ThenAnyEx(cbs ...func(context.Context) (context.Context, error)) *FutureContext {
	future := f.future.Then(func(ctx context.Context) (context.Context, error) {
		return NewFutureContext(ctx, func(resolve func(context.Context), reject func(error)) {
			for _, fn := range cbs {
				go func(fn func(context.Context) (context.Context, error)) {
					res, err := fn(ctx)
					if err != nil {
						reject(err)
						return
					}
					resolve(res)
				}(fn)
			}
		}).Collect()
	})
	return &FutureContext{future: future, ctx: f.ctx}
}

// ThenAll return if all of the callbacks succeed. It returns a new Future.
func (f *FutureContext) ThenAll(cbs ...func() error) *FutureContext {
	return f.ThenAllEx(lo.Map(cbs, func(cb func() error, index int) func(context.Context) (context.Context, error) {
		return func(ctx context.Context) (context.Context, error) {
			return ctx, cb()
		}
	})...)
}

// ThenAllEx return if all of the callbacks succeed. It returns a new Future.
func (f *FutureContext) ThenAllEx(cbs ...func(context.Context) (context.Context, error)) *FutureContext {
	future := f.future.Then(func(ctx context.Context) (context.Context, error) {
		return NewFutureContext(ctx, func(resolve func(context.Context), reject func(error)) {
			var results []any
			wg := sync.WaitGroup{}
			wg.Add(len(cbs))
			for i, fn := range cbs {
				go func(i int, fn func(context.Context) (context.Context, error)) {
					res, err := fn(ctx)
					if err != nil {
						reject(err)
						return
					}
					results[i], _ = FromContext[any](res)
				}(i, fn)
			}
			wg.Wait()
			resolve(WithContext(ctx, results))
		}).Collect()
	})
	return &FutureContext{future: future, ctx: f.ctx}
}

// Finally is called when Future is processed either resolved or rejected. It returns a new Future.
func (f *FutureContext) Finally(cb func(context.Context, error) (context.Context, error)) *FutureContext {
	future := f.future.Finally(cb)
	return &FutureContext{future: future, ctx: f.ctx}
}

// Catch is called when Future is rejected. It returns a new Future.
func (f *FutureContext) Catch(cb func(error) error) *FutureContext {
	future := f.future.Catch(func(err error) (context.Context, error) {
		return f.ctx, cb(err)
	})
	return &FutureContext{future: future, ctx: f.ctx}
}

func (f *FutureContext) Collect() (context.Context, error) {
	return f.future.Collect()
}

func (f *FutureContext) Cancel() {
	f.future.Cancel()
}

func (f *FutureContext) Result() mo.Result[context.Context] {
	return f.future.Result()
}

func (f *FutureContext) Either() mo.Either[error, context.Context] {
	return f.future.Either()
}
