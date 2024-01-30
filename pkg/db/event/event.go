package event

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
)

type EventType string

const (
	Creating     EventType = "creating"
	Created      EventType = "created"
	Updating     EventType = "updating"
	Updated      EventType = "updated"
	BatchUpdated EventType = "batch_updated"
	// Saving 包含Creating/Updating
	Saving EventType = "saving"
	// Saved 包含Created/Updated
	Saved        EventType = "saved"
	Deleting     EventType = "deleting"
	Deleted      EventType = "deleted"
	BatchDeleted EventType = "batch_deleted"
	Found        EventType = "found"
	Customer     EventType = "customer"
)

type ModelEvent[T schema.Tabler] struct {
	EventType EventType
	Tx        *gorm.DB
	Model     T
	Arguments []any
}

// EventListenerFunc 事件监听器函数
type EventListenerFunc[T schema.Tabler] func(ctx context.Context, modelEvent ModelEvent[T]) error

// Events 事件监听器
type Events[T schema.Tabler] map[EventType][]EventListenerFunc[T]

func (e Events[T]) FireEvent(ctx context.Context, tx *gorm.DB, event EventType, model T, arguments ...any) error {
	if callbacks, ok := e[event]; ok {
		for _, callback := range callbacks {
			if err := callback(ctx, ModelEvent[T]{
				EventType: event,
				Tx:        tx,
				Model:     model,
				Arguments: arguments,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

// RegisterEventListener 注册事件监听器，当事件触发时，会调用callback。注意：不是协程安全的。（因为大部分情况下，注册事件监听器是在初始化时完成的）
func (e Events[T]) RegisterEventListener(eventType EventType, callback EventListenerFunc[T]) {
	if _, ok := e[eventType]; !ok {
		e[eventType] = make([]EventListenerFunc[T], 0)
	}
	e[eventType] = append(e[eventType], callback)
}

// RemoveEventListener 移除eventType的callback监听器，使用的是reflect.ValueOf(callback).Pointer()来判断callback是否相等。如存在返回true，否则返回false
func (e Events[T]) RemoveEventListener(eventType EventType, callback EventListenerFunc[T]) bool {
	fp := reflect.ValueOf(callback).Pointer()
	if callbacks, ok := e[eventType]; ok {
		for i, cb := range callbacks {
			if reflect.ValueOf(cb).Pointer() == fp {
				e[eventType] = append(e[eventType][:i], e[eventType][i+1:]...)
				return true
			}
		}
	}
	return false
}

// RemoveAllEventListeners 移除eventType的所有监听器
func (e Events[T]) RemoveAllEventListeners(eventType EventType) {
	e[eventType] = make([]EventListenerFunc[T], 0)
}
