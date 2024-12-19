package utils

import (
	"github.com/samber/lo"
	"iter"
	"reflect"
	"sync"
)

type IEventListeners[T any] interface {
	// Add 添加监听器
	Add(listeners ...T) IEventListeners[T]
	// Remove 移除监听器
	Remove(listeners ...T) IEventListeners[T]
	// Clear 清空监听器
	Clear() IEventListeners[T]
	// Iterator 迭代器
	Iterator() iter.Seq2[int, T]
	// Listeners 获取监听器列表（副本）
	Listeners() []T
}

type EventListeners[T any] struct {
	listeners []T
	mutex     sync.RWMutex
}

var _ IEventListeners[any] = (*EventListeners[any])(nil)

func NewEventListeners[T any]() IEventListeners[T] {
	return &EventListeners[T]{
		mutex: sync.RWMutex{},
	}
}

// Add 添加监听器
func (e *EventListeners[T]) Add(listeners ...T) IEventListeners[T] {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	// 避免重复添加监听器，通过对比Ptr来实现
	_oldPtrs := lo.Map(e.listeners, func(listener T, _ int) uintptr {
		ptr := reflect.ValueOf(listener)
		return ptr.Pointer()
	})

	for _, listener := range listeners {
		ptr := reflect.ValueOf(listener).Pointer()
		if lo.Contains(_oldPtrs, ptr) {
			continue
		}
		_oldPtrs = append(_oldPtrs, ptr) // 避免listeners中有重复
		e.listeners = append(e.listeners, listener)
	}
	return e
}

// Remove 移除监听器
func (e *EventListeners[T]) Remove(listeners ...T) IEventListeners[T] {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	// 获取待移除的监听器的Ptr
	deletingPtrs := lo.Map(listeners, func(listener T, _ int) uintptr {
		return reflect.ValueOf(listener).Pointer()
	})
	// 移除指定监听器
	e.listeners = lo.Filter(e.listeners, func(listener T, _ int) bool {
		return !lo.Contains(deletingPtrs, reflect.ValueOf(listener).Pointer())
	})
	return e
}

// Clear 清空监听器
func (e *EventListeners[T]) Clear() IEventListeners[T] {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	clear(e.listeners)
	e.listeners = nil
	return e
}

// Iterator 迭代器
func (e *EventListeners[T]) Iterator() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		e.mutex.RLock()
		defer e.mutex.RUnlock()
		for i, listener := range e.listeners {
			if !yield(i, listener) {
				break
			}
		}
	}
}

// Listeners 获取监听器列表（副本）
func (e *EventListeners[T]) Listeners() []T {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.listeners
}
