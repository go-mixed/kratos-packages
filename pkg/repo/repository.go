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

	operations repositoryOperations[T]
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
		operations: repositoryOperations[T]{
			preloads: make(map[string][]any),
		},
	}

	typeOf := reflect.TypeOf(modelCreator())
	if typeOf == nil || typeOf.Kind() != reflect.Ptr {
		panic("modelCreator[T] must return a pointer of model")
	}

	// Hook当前model的事件
	db.BindModelEvents(modelCreator(), repo.onModelEvent)
	return repo
}

// Clone 这是一个不在interface中导出的方法，如果要在IOrm[T]后使用，需要断言转化
func (repo *Repository[T]) Clone() *Repository[T] {
	clone := *repo
	clone.operations = repo.operations
	clone.operations.preloads = make(map[string][]any)
	for k, v := range repo.operations.preloads {
		clone.operations.preloads[k] = v
	}
	return &clone
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
