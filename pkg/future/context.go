package future

import "context"

type promiseResult struct{}

func WithContext(ctx context.Context, val any) context.Context {
	return context.WithValue(ctx, promiseResult{}, val)
}
func FromContext[T any](ctx context.Context) (T, bool) {
	val := ctx.Value(promiseResult{})
	if val == nil {
		var nilT T
		return nilT, false
	}
	_v, ok := val.(T)
	return _v, ok
}
