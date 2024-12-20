package utils

import (
	"iter"
	"sync"
)

type ConcurrentMap[K comparable, V any] struct {
	m sync.Map
}

// Delete 删除Key
func (m *ConcurrentMap[K, V]) Delete(key K) { m.m.Delete(key) }

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
func (m *ConcurrentMap[K, V]) Iterator() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.Range(func(key K, value V) bool {
			return yield(key, value)
		})
	}
}

// Store 存储
func (m *ConcurrentMap[K, V]) Store(key K, value V) { m.m.Store(key, value) }

// Clear 原子性的移除所有key
func (m *ConcurrentMap[K, V]) Clear() {
	m.Range(func(key K, value V) bool {
		m.Delete(key)
		return true
	})
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
