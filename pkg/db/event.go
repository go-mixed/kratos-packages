package db

import (
	"context"
	"fmt"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/event"
	"gorm.io/gorm/schema"
)

type onEventFunc func(ctx context.Context, tx *DB, model schema.Tabler, event event.EventType, args ...any) error

var globalEventListeners map[string]onEventFunc

func init() {
	globalEventListeners = make(map[string]onEventFunc)
}

func getEventName(t schema.Tabler) string {
	return fmt.Sprintf("model.%s", t.TableName())
}

// fireModelEvent 由Model调用，触发事件
func fireModelEvent(ctx context.Context, tx *DB, model schema.Tabler, event event.EventType) error {
	if model.TableName() == "" {
		return nil
	}

	listener, exists := globalEventListeners[getEventName(model)]
	if exists {
		return listener(ctx, tx, model, event)
	}
	return nil
}

// BindModelEvents 绑定T的所有事件
func BindModelEvents[T schema.Tabler](t T, callback onEventFunc) {
	if t.TableName() == "" {
		panic("model.TableName() must not be empty")
	}
	globalEventListeners[getEventName(t)] = callback
}
