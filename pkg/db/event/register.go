package event

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
)

type iModelEvent interface {
	OnEvent(context.Context, *gorm.DB, schema.Tabler, EventType) error
	schema.Tabler
}

func RegisterGormEvents(db *gorm.DB) {
	createCallback := db.Callback().Create()
	createCallback.Register("gorm:before_create", beforeCreate)
	createCallback.Register("gorm:after_create", afterCreate)

	queryCallback := db.Callback().Query()
	queryCallback.Register("gorm:after_query", afterQuery)

	deleteCallback := db.Callback().Delete()
	deleteCallback.Register("gorm:before_delete", beforeDelete)
	deleteCallback.Register("gorm:after_delete", afterDelete)

	updateCallback := db.Callback().Update()
	updateCallback.Register("gorm:before_update", beforeUpdate)
	updateCallback.Register("gorm:after_update", afterUpdate)
}

// beforeCreate before create hooks
func beforeCreate(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks {
		callMethod(db, func(value any, tx *gorm.DB) (called bool) {
			ctx := db.Statement.Context
			if i, ok := value.(iModelEvent); ok {
				called = true
				db.AddError(i.OnEvent(ctx, tx, i, Saving))
				db.AddError(i.OnEvent(ctx, tx, i, Creating))

			}
			return called
		})
	}
}

// afterCreate after create hooks
func afterCreate(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks {
		callMethod(db, func(value any, tx *gorm.DB) (called bool) {
			ctx := db.Statement.Context
			if i, ok := value.(iModelEvent); ok {
				called = true
				db.AddError(i.OnEvent(ctx, tx, i, Created))
				db.AddError(i.OnEvent(ctx, tx, i, Saved))
			}

			return called
		})
	}
}

func afterQuery(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks && db.RowsAffected > 0 {
		callMethod(db, func(value any, tx *gorm.DB) bool {
			ctx := db.Statement.Context
			if i, ok := value.(iModelEvent); ok {
				db.AddError(i.OnEvent(ctx, tx, i, Found))
				return true
			}
			return false
		})
	}
}

func beforeDelete(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks {
		callMethod(db, func(value any, tx *gorm.DB) bool {
			ctx := db.Statement.Context
			if i, ok := value.(iModelEvent); ok {
				db.AddError(i.OnEvent(ctx, tx, i, Deleting))
				return true
			}

			return false
		})
	}
}

func afterDelete(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks {
		callMethod(db, func(value any, tx *gorm.DB) bool {
			ctx := db.Statement.Context
			if i, ok := value.(iModelEvent); ok {
				db.AddError(i.OnEvent(ctx, tx, i, Deleted))
				return true
			}
			return false
		})
	}
}

// beforeUpdate before update hooks
func beforeUpdate(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks {
		callMethod(db, func(value any, tx *gorm.DB) (called bool) {
			ctx := db.Statement.Context
			if i, ok := value.(iModelEvent); ok {
				called = true
				db.AddError(i.OnEvent(ctx, tx, i, Saving))
				db.AddError(i.OnEvent(ctx, tx, i, Updating))
			}

			return called
		})
	}
}

// afterUpdate after update hooks
func afterUpdate(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks {
		callMethod(db, func(value any, tx *gorm.DB) (called bool) {
			ctx := db.Statement.Context
			if db.Statement.Schema.AfterUpdate {
				if i, ok := value.(iModelEvent); ok {
					called = true
					db.AddError(i.OnEvent(ctx, tx, i, Updated))
					db.AddError(i.OnEvent(ctx, tx, i, Saved))
				}
			}

			return called
		})
	}
}

func callMethod(db *gorm.DB, fc func(value any, tx *gorm.DB) bool) {
	tx := db.Session(&gorm.Session{NewDB: true})
	if called := fc(db.Statement.ReflectValue.Interface(), tx); !called {
		switch db.Statement.ReflectValue.Kind() {
		case reflect.Slice, reflect.Array:
			db.Statement.CurDestIndex = 0
			for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
				if value := reflect.Indirect(db.Statement.ReflectValue.Index(i)); value.CanAddr() {
					fc(value.Addr().Interface(), tx)
				} else {
					db.AddError(gorm.ErrInvalidValue)
					return
				}
				db.Statement.CurDestIndex++
			}
		case reflect.Struct:
			if db.Statement.ReflectValue.CanAddr() {
				fc(db.Statement.ReflectValue.Addr().Interface(), tx)
			} else {
				db.AddError(gorm.ErrInvalidValue)
			}
		}
	}
}
