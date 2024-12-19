package utils

import (
	"sync/atomic"
)

type ObserveListener[T comparable] func(preValue, newValue T)

type IObserve[T comparable] interface {
	// Set Atomic set new value
	Set(value T)
	// SetAndTrigger Atomic set new value, force triggers listeners
	SetAndTrigger(value T)
	// CompareAndSwap Atomic compare and swap value, triggers listeners ONLY oldValue != newValue
	CompareAndSwap(oldValue T, newValue T) (swapped bool)
	// Value Get current value(atomic)
	Value() T
	// Watch call listeners if value changed
	Watch(listeners ...ObserveListener[T]) IObserve[T]
}

type observer[T comparable] struct {
	value     atomic.Value
	listeners IEventListeners[ObserveListener[T]]
}

// NewObserver Create new observer
func NewObserver[T comparable](initialValue T) IObserve[T] {
	value := atomic.Value{}
	value.Store(initialValue)
	o := &observer[T]{
		value:     value,
		listeners: NewEventListeners[ObserveListener[T]](),
	}
	return o
}

// trigger triggers listeners
func (o *observer[T]) trigger(oldValue T, newValue T) {
	for _, listener := range o.listeners.Iterator() {
		listener(oldValue, newValue)
	}
}

// Set Atomic set new value, triggers listeners if value changed
func (o *observer[T]) Set(value T) {
	oldValue := o.value.Swap(value)
	oldValueT := oldValue.(T)
	// trigger listeners if value changed
	if oldValueT != value {
		o.trigger(oldValueT, value)
	}
}

// SetAndTrigger Atomic set new value, force triggers listeners
func (o *observer[T]) SetAndTrigger(value T) {
	oldValue := o.value.Swap(value)
	oldValueT := oldValue.(T)

	o.trigger(oldValueT, value)
}

// CompareAndSwap Atomic compare and swap value, triggers listeners ONLY oldValue != newValue
func (o *observer[T]) CompareAndSwap(oldValue, newValue T) (swapped bool) {
	swapped = o.value.CompareAndSwap(oldValue, newValue)

	if swapped && oldValue != newValue {
		o.trigger(oldValue, newValue)
	}
	return swapped
}

// Value Get current value(atomic)
func (o *observer[T]) Value() T {
	return o.value.Load().(T)
}

// Watch call listeners if value changed
func (o *observer[T]) Watch(listeners ...ObserveListener[T]) IObserve[T] {
	o.listeners.Add(listeners...)
	return o
}

// BoolObserver Observer for bool type
type BoolObserver = IObserve[bool]

// NewBoolObserver Create new bool observer
func NewBoolObserver(val bool) BoolObserver {
	return NewObserver[bool](val)
}

// IntObserver Observer for int type
type IntObserver = IObserve[int]

// NewIntObserver Create new int observer
func NewIntObserver(val int) IntObserver {
	return NewObserver[int](val)
}

// Int64Observer Observer for int64 type
type Int64Observer = IObserve[int64]

// NewInt64Observer Create new int64 observer
func NewInt64Observer(val int64) Int64Observer {
	return NewObserver[int64](val)
}

// Int32Observer Observer for int32 type
type Int32Observer = IObserve[int32]

// NewInt32Observer Create new int32 observer
func NewInt32Observer(val int32) Int32Observer {
	return NewObserver[int32](val)
}

// Int8Observer Observer for int8 type
type Int8Observer = IObserve[int8]

// NewInt8Observer Create new int8 observer
func NewInt8Observer(val int8) Int8Observer {
	return NewObserver[int8](val)
}

// UintObserver Observer for uint type
type UintObserver = IObserve[uint]

// NewUintObserver Create new uint observer
func NewUintObserver(val uint) UintObserver {
	return NewObserver[uint](val)
}

// Uint64Observer Observer for uint64 type
type Uint64Observer = IObserve[uint64]

// NewUint64Observer Create new uint64 observer
func NewUint64Observer(val uint64) Uint64Observer {
	return NewObserver[uint64](val)
}

// Uint32Observer Observer for uint32 type
type Uint32Observer = IObserve[uint32]

// NewUint32Observer Create new uint32 observer
func NewUint32Observer(val uint32) Uint32Observer {
	return NewObserver[uint32](val)
}

// Uint8Observer Observer for uint8 type
type Uint8Observer = IObserve[uint8]

// NewUint8Observer Create new uint8 observer
func NewUint8Observer(val uint8) Uint8Observer {
	return NewObserver[uint8](val)
}

// Float64Observer Observer for float64 type
type Float64Observer = IObserve[float64]

// NewFloat64Observer Create new float64 observer
func NewFloat64Observer(val float64) Float64Observer {
	return NewObserver[float64](val)
}

// Float32Observer Observer for float32 type
type Float32Observer = IObserve[float32]

// NewFloat32Observer Create new float32 observer
func NewFloat32Observer(val float32) Float32Observer {
	return NewObserver[float32](val)
}

// StringObserver Observer for string type
type StringObserver = IObserve[string]

// NewStringObserver Create new string observer
func NewStringObserver(val string) StringObserver {
	return NewObserver[string](val)
}
