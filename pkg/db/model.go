package db

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/event"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

type Model struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *Model) GetID() int64 {
	return m.ID
}

func (m *Model) TableName() string {
	return ""
}

func (m *Model) OnEvent(ctx context.Context, tx *gorm.DB, model schema.Tabler, eventType event.EventType) error {
	return fireModelEvent(ctx, tx, model, eventType)
}

type SoftDeleteModel struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (m *SoftDeleteModel) GetID() int64 {
	return m.ID
}

func (m *SoftDeleteModel) TableName() string {
	return ""
}

func (m *SoftDeleteModel) OnEvent(ctx context.Context, tx *gorm.DB, model schema.Tabler, eventType event.EventType) error {
	return fireModelEvent(ctx, tx, model, eventType)
}
