package repo

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/cache"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/clause"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/cnd"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/event"
)

type IOrm[T db.Tabler] interface {
	// Transaction 开启事务
	// example:
	// repo.Transaction(ctx, func(ctx context.Context) error {
	//     m, err := repo.First(ctx, db.ID(1).LockingForUpdate())
	//	   return err
	//}, func(ctx context.Context) error {
	//	   return repo.Create(ctx, &model)
	//}...)
	//	注意：在Transaction之前设置的Clauses、Select、Omit、Preloads、PreloadWithBuilder，在steps中是不生效的
	Transaction(ctx context.Context, steps ...func(ctx context.Context) error) error

	// GetDB 获取db 从上下文中取出db
	GetDB(ctx context.Context) *db.DB

	IOrmOperation[T]
	IOrmSetter[T]
	IOrmGetter[T]
	IOrmScanner
}

type IOrmOperation[T db.Tabler] interface {
	// Clauses 条件：冲突解决、读写分离等
	//   - 比如：Clauses(clause.Write).Find(ctx, cnd.Eq("id", 1))，表示强制使用主库查询
	Clauses(cnds ...clause.Expression) IOrm[T]
	// Preloads 不带条件的预加载关联，可以一次设置多个关联名
	Preloads(preloads ...string) IOrm[T]
	// PreloadWithBuilder 带条件的预加载，一次只能设置一个关联
	PreloadWithBuilder(preload string, args ...any) IOrm[T]
	// Select 指定查询/更新/创建字段。比如：Select("name", "age").Find(ctx, cnd.Eq("id", 1))，表示只查询name和age字段
	//
	//	FieldName：如果后面传递是Struct，使用Struct的字段名；如果后面传递是map，使用map的key
	Select(query any, args ...any) IOrm[T]
	// Omit 排除字段。比如：Omit("name", "age").Find(ctx, cnd.Eq("id", 1))，表示排除name和age字段
	//
	//	FieldName：如果后面传递是Struct，使用Struct的字段名；如果后面传递是map，使用map的key
	Omit(columns ...string) IOrm[T]
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
	// Update 批量更新资源。如果没有主键，为了避免批量更新，会返回ErrMissingWhereClaus。
	// 必须要指定更新的字段，否则只会更新非零值字段，如果要更新全部字段，可以设置Select("*")
	// example: repo.Select("Name", "Gender").Update(ctx, &User{ID: 1, Name: "tom", Gender: "male"})
	// T必须为指针类型
	Update(ctx context.Context, models ...T) error
	// UpdateColumns 更新资源的多个字段
	UpdateColumns(ctx context.Context, query *cnd.QueryBuilder, attributes db.Columns) error
	// UpdateColumn 更新资源的单个字段
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
	First(ctx context.Context, query *cnd.QueryBuilder) (T, error)
	// FirstOrFail 查询第一个资源，如果没有找到返回gorm.ErrRecordNotFound错误
	// query 查询条件
	FirstOrFail(ctx context.Context, query *cnd.QueryBuilder) (T, error)

	// Find 通过主键查询第一个资源，如果没有找到【不会】gorm.ErrRecordNotFound
	// primary 主键值
	Find(ctx context.Context, primary any) (T, error)
	// FindOrFail 通过主键查询第一个资源，如果没有找到返回gorm.ErrRecordNotFound错误
	FindOrFail(ctx context.Context, primary any) (T, error)

	// FindMany 通过主键查询多个资源，如果没有找到【不会】gorm.ErrRecordNotFound
	FindMany(ctx context.Context, primary []any) ([]T, error)

	// Get 查询获取资源集合，如果没有找到【不会】gorm.ErrRecordNotFound
	// query 查询条件
	Get(ctx context.Context, query *cnd.QueryBuilder) ([]T, error)

	// Paginate 资源分页
	Paginate(ctx context.Context, query *cnd.QueryBuilder, pagination *db.Pagination) ([]T, error)
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
	getCacheDriver() *cache.Cache
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

type IRemember[T db.Tabler] interface {
	IOrmGetter[T]
	// Do 自定义返回内容
	Do(ctx context.Context, callback func(context.Context, IRepository[T]) (any, error)) (any, error)
}
