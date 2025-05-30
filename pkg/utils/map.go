package utils

import (
	"iter"
	"sync"
)

type ConcurrentMap[K comparable, V any] struct {
	m sync.Map
}

func (m *ConcurrentMap[K, V]) Len() int {
	var l int
	m.Range(func(key K, value V) bool {
		l += 1
		return true
	})
	return l
}

// Delete 删除Key
func (m *ConcurrentMap[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// Has 判断Key是否存在
func (m *ConcurrentMap[K, V]) Has(key K) bool {
	_, ok := m.m.Load(key)
	return ok
}

// Load 获取Key
func (m *ConcurrentMap[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return value, ok
	}
	return v.(V), ok
}

// LoadAndDelete 原子性的，获取之后，删除
func (m *ConcurrentMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		return value, loaded
	}
	return v.(V), loaded
}

// LoadOrStore 原子性的，存储或获取，如果key不存在则存储，如果key存在则返回
func (m *ConcurrentMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	a, loaded := m.m.LoadOrStore(key, value)
	return a.(V), loaded
}

// Range 原子性的遍历
func (m *ConcurrentMap[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool { return f(key.(K), value.(V)) })
}

// Iterator 原子性的迭代器
func (m *ConcurrentMap[K, V]) Iterator(fns ...MapFilterFunc[K, V]) iter.Seq2[K, V] {
	fn := WrapMapFilterFunc(fns...)
	return func(yield func(K, V) bool) {
		m.Range(func(key K, value V) bool {
			// 如果被过滤，则不调用yield，但是还是需要继续循环
			if !fn(key, value) {
				return true
			}
			return yield(key, value)
		})
	}
}

// Store 存储
func (m *ConcurrentMap[K, V]) Store(key K, value V) { m.m.Store(key, value) }

// Swap 原子性的替换
func (m *ConcurrentMap[K, V]) Swap(key K, value V) (old V, loaded bool) {
	o, loaded := m.m.Swap(key, value)
	if !loaded {
		return old, loaded
	}
	return o.(V), loaded
}

// CompareAndSwap 原子性的替换，如果key存在则替换，如果key不存在则存储
func (m *ConcurrentMap[K, V]) CompareAndSwap(key K, old, new V) (swapped bool) {
	return m.m.CompareAndSwap(key, old, new)
}

// CompareAndDelete 原子性的替换，如果key存在且等于old则替换，如果key不存在则存储
func (m *ConcurrentMap[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.m.CompareAndDelete(key, old)
}

// Clear 原子性的移除所有key
func (m *ConcurrentMap[K, V]) Clear() {
	m.m.Clear()
}

// RemoveExcept 原子性的移除不存在的key
func (m *ConcurrentMap[K, V]) RemoveExcept(keys ...K) {
	// 将keys转为map[K]，便于快速判断key是否存在
	mm := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		mm[k] = struct{}{}
	}

	m.Range(func(key K, value V) bool {
		// 如果key不存在，则删除
		if _, ok := mm[key]; !ok {
			m.Delete(key)
		}
		return true
	})
}
