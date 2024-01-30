package repo

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/event"
	"gorm.io/gorm/schema"
)

// RegisterEventListener 注册单个Model的事件
func (repo *Repository[T]) RegisterEventListener(eventType event.EventType, callback event.EventListenerFunc[T]) {
	repo.RegisterEventListeners([]event.EventType{eventType}, callback)
}

// RegisterEventListeners 注册多个Model的事件
func (repo *Repository[T]) RegisterEventListeners(eventTypes []event.EventType, callback event.EventListenerFunc[T]) {
	for _, eventType := range eventTypes {
		repo.events.RegisterEventListener(eventType, callback)
	}
}

// onModelEvent 模型事件回调
func (repo *Repository[T]) onModelEvent(ctx context.Context, tx *db.DB, model schema.Tabler, eventType event.EventType, args ...any) error {
	var m T
	if model != nil {
		m = model.(T)
	}
	return repo.events.FireEvent(ctx, tx, eventType, m)
}

// FireEvent 手动触发事件
func (repo *Repository[T]) FireEvent(ctx context.Context, model T, args ...any) error {
	return repo.onModelEvent(ctx, nil, model, event.Customer, args...)
}
