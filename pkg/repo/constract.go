package repo

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/cache"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/clause"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/cnd"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/event"
)

type Columns = map[string]any

type IOrm[T db.Tabler] interface {
	// Transaction 开启事务
	// example:
	// repo.Transaction(ctx, func(ctx context.Context) error {
	//     m, err := repo.First(ctx, db.ID(1).LockingForUpdate())
	//	   return err
	//}, func(ctx context.Context) error {
	//	   return repo.Create(ctx, &model)
	//}...)
	Transaction(ctx context.Context, steps ...func(ctx context.Context) error) error

	// GetDB 获取db 从上下文中取出db
	GetDB(ctx context.Context) *db.DB

	Clauses(cnds ...clause.Expression) IOrm[T]

	IOrmSetter[T]
	IOrmGetter[T]
	IOrmScanner
}

type IOrmScanner interface {
	// Pluck 获取资源单个字段集合，如果没有找到【不会】gorm.ErrRecordNotFound
	Pluck(ctx context.Context, queries *cnd.QueryBuilder, field string, scanner any) error
}

type IOrmSetter[T db.Tabler] interface {
	// Create 批量创建资源
	// example: repo.Create(ctx, &User{Name: "tom"}, &User{Name: "jerry"})
	// T必须为指针类型
	Create(ctx context.Context, models ...T) error
	// Save 保存资源，如果主键为空，则创建，否则更新。注意：零值【会】更新。
	// 注意：官方建议使用Create、Update，而不是Save。假如期望是Update行为，但是ID并未设置，这会导致创建了新的记录
	// example: repo.Save(ctx, &User{ID: 1, Name: "tom"}, &User{Name: "jerry"})
	// T必须为指针类型
	Save(ctx context.Context, models ...T) error
	// Update 批量更新资源，注意：零值不会更新。如果没有主键，为了避免批量更新，会返回ErrMissingWhereClaus
	// example: repo.Update(ctx, &User{ID: 1, Name: "tom"}, &User{ID: 2, Name: "jerry"})
	// T必须为指针类型
	Update(ctx context.Context, models ...T) error
	// UpdateColumns 更新资源多个字段
	UpdateColumns(ctx context.Context, query *cnd.QueryBuilder, attributes Columns) error
	// UpdateColumn 更新资源单个字段
	UpdateColumn(ctx context.Context, query *cnd.QueryBuilder, key string, value any) error
	// Delete 使用model删除资源。如果没有主键，为了避免批量删除，会返回ErrMissingWhereClause
	// example: repo.Delete(ctx, &User{ID: 1}, &User{ID: 2})
	// T必须为指针类型
	Delete(ctx context.Context, models ...T) error
	// DeleteWithBuilder 使用query删除资源
	// example: repo.Delete(ctx, db.ID(1))、repo.Delete(ctx, cnd.Where("name", "tom"))
	DeleteWithBuilder(ctx context.Context, query *cnd.QueryBuilder) error
	// DeletePrimary 通过主键删除资源
	// example: repo.DeletePrimary(ctx, 1, 2, 3)
	DeletePrimary(ctx context.Context, primary ...any) error

	// Incr 递增某字段
	Incr(ctx context.Context, query *cnd.QueryBuilder, field string, val any) error
	// Decr 递减某字段
	Decr(ctx context.Context, query *cnd.QueryBuilder, field string, val any) error
}

type IOrmGetter[T db.Tabler] interface {
	// Count 查询资源数量
	// query 查询条件
	Count(ctx context.Context, query *cnd.QueryBuilder) (int64, error)

	// First 查询第一个资源，如果没有找到【不会】gorm.ErrRecordNotFound，但是scanner仍然可能被空值实例化
	// query 查询条件
	// scanner 结构指针
	First(ctx context.Context, query *cnd.QueryBuilder) (T, error)

	// Find 通过主键查询第一个资源，如果没有找到【不会】gorm.ErrRecordNotFound
	// primary 主键值
	// scanner 结构指针
	Find(ctx context.Context, primary any) (T, error)
	// FindOrFail 通过主键查询第一个资源，如果没有找到返回gorm.ErrRecordNotFound错误
	FindOrFail(ctx context.Context, primary any) (T, error)

	// FindMany 通过主键查询多个资源，如果没有找到【不会】gorm.ErrRecordNotFound
	FindMany(ctx context.Context, primary []any) ([]T, error)

	// Get 查询获取资源集合，如果没有找到【不会】gorm.ErrRecordNotFound
	// query 查询条件
	// scanner 结构指针
	Get(ctx context.Context, query *cnd.QueryBuilder) ([]T, error)

	// Paginate 资源分页
	Paginate(ctx context.Context, query *cnd.QueryBuilder, pagination *db.Pagination) ([]T, error)
}

type IRemember[T db.Tabler] interface {
	IOrmGetter[T]
	Do(ctx context.Context, callback func(context.Context, *Repository[T]) (any, error)) (any, error)
}

type IModelEvent[T db.Tabler] interface {
	// RegisterEventListener 注册单个Model的事件
	RegisterEventListener(eventType event.EventType, callback event.EventListenerFunc[T])
	// RegisterEventListeners 注册多个Model的事件
	RegisterEventListeners(eventTypes []event.EventType, callback event.EventListenerFunc[T])
	// FireEvent 手动触发事件
	FireEvent(ctx context.Context, model T, args ...any) error
}

type IRepositoryCache[T db.Tabler] interface {
	// Remember 如果有缓存，则返回缓存，不然就执行后面的动作
	// 此函数不能单独使用，需要：repo.Remember("key-1", 0).Get(ctx, cnd.Where(...))
	// 如果只想得到Cache，可以使用GetCache、GetCacheForModel、GetCacheForModelList；如果只想设置Cache，可以使用SetCache
	// 注意：默认情况下，没有查询到记录（包括Count()==0），不会设置缓存。
	// 当options传入WithSaveEmptyOnRemember()，可以强制保存空值
	// 如果要修改缓存的过期时间，可以传递WithExpiration()，如果要修改缓存的key前缀，可以WithKeyPrefix()
	Remember(key string, options ...cache.Option) IRemember[T]
	// GetCache 获取某key的cache，并转化为T对象
	GetCache(ctx context.Context, key string) (bool, T, error)
	// ForgetCache 删除某keys的cache
	ForgetCache(ctx context.Context, keys ...string) error
	// GetCacheForList 获取某key的cache，并转化为[]T对象列表
	GetCacheForList(ctx context.Context, key string) ([]T, error)
}

type IRepository[T db.Tabler] interface {
	IOrm[T]
	IModelEvent[T]
	IRepositoryCache[T]
}
