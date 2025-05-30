package utils

import (
	"github.com/samber/lo"
	"iter"
)

// FilterFunc 过滤函数，返回true则表示通过
type FilterFunc[P any] func(P) bool

// WrapFilterFunc 将多个用于slice的过滤条件逻辑组合成一个过滤器
func WrapFilterFunc[P any, T FilterFunc[P]](fns ...T) T {
	return lo.Reduce(fns, func(r T, fn T, _ int) T {
		return func(p P) bool {
			return r(p) && fn(p)
		}
	}, func(P) bool {
		return true
	})
}

// FilterWithError 过滤函数，返回nil则表示通过，否则返回错误
type FilterWithError[P any] func(P) error

// WrapFilterWithError 将多个用于slice的过滤条件逻辑组合成一个过滤器
func WrapFilterWithError[P any, T FilterWithError[P]](fns ...T) T {
	return lo.Reduce(fns, func(r T, fn T, _ int) T {
		return func(p P) error {
			if err := r(p); err != nil {
				return err
			}
			return fn(p)
		}
	}, func(P) error {
		return nil
	})
}

// MapFilterFunc 用于map的过滤函数，返回true则表示通过
type MapFilterFunc[K comparable, V any] func(K, V) bool

// WrapMapFilterFunc 将多个用于map的过滤条件逻辑组合成一个过滤器
func WrapMapFilterFunc[K comparable, V any, T MapFilterFunc[K, V]](fns ...T) T {
	return lo.Reduce(fns, func(r T, fn T, _ int) T {
		return func(k K, v V) bool {
			return r(k, v) && fn(k, v)
		}
	}, func(K, V) bool {
		return true
	})
}

// MapFilterWithErrorFunc 用于map的过滤函数，返回nil则表示通过，否则返回错误
type MapFilterWithErrorFunc[K comparable, V any] func(K, V) error

// WrapMapFilterWithErrorFunc 将多个用于map的过滤条件逻辑组合成一个过滤器
func WrapMapFilterWithErrorFunc[K comparable, V any, T MapFilterWithErrorFunc[K, V]](fns ...T) T {
	return lo.Reduce(fns, func(r T, fn T, _ int) T {
		return func(k K, v V) error {
			if err := r(k, v); err != nil {
				return err
			}
			return fn(k, v)
		}
	}, func(K, V) error {
		return nil
	})
}

// MapIterator map带filter的迭代器
//   - （注意：没有可以输入FilterMapWithErrorFunc的迭代器，因为错误无法在for中返回）
func MapIterator[K comparable, V any, T MapFilterFunc[K, V]](m map[K]V, fns ...T) iter.Seq2[K, V] {
	fn := WrapMapFilterFunc(fns...)
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if fn(k, v) && !yield(k, v) {
				break
			}
		}
	}
}
