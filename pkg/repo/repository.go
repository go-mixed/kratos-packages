package repo

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/cache"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/event"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"reflect"
)

type Repository[T db.Tabler] struct {
	db           *db.DB
	cache        *cache.Cache
	modelCreator func() T
	logger       *log.Helper

	events event.Events[T]
}

func NewRepository[T db.Tabler](
	_db *db.DB,
	_cache *cache.Cache,
	modelCreator func() T,
	logger log.Logger) *Repository[T] {

	repo := &Repository[T]{
		db:           _db,
		cache:        _cache,
		modelCreator: modelCreator,
		logger:       log.NewModuleHelper(logger, "repo/base"),

		events: make(event.Events[T]),
	}

	typeOf := reflect.TypeOf(modelCreator())
	if typeOf == nil || typeOf.Kind() != reflect.Ptr {
		panic("modelCreator[T] must return a pointer of model")
	}

	// Hook当前model的事件
	db.BindModelEvents(modelCreator(), repo.onModelEvent)
	return repo
}

// GetDB 获取db，如果是事务，并将ctx附加到gorm中
func (repo *Repository[T]) GetDB(ctx context.Context) *db.DB {
	if ctx != nil {
		if tx := ctx.Value(&transactionKey{}); tx != nil {
			return tx.(*db.DB).WithContext(ctx)
		}
	}

	return repo.db.WithContext(ctx)
}

// Transaction 开启事务
func (repo *Repository[T]) Transaction(ctx context.Context, steps ...func(ctx context.Context) error) error {
	var err error
	tx := repo.db.Begin()
	defer func() {
		if err != nil {
			repo.logger.WithContext(ctx).Error(err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx = newTxContext(ctx, tx)

	for _, step := range steps {
		if err = step(ctx); err != nil {
			return err
		}
	}

	return nil
}
