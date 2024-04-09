package repo

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/cache"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/cnd"
)

func (repo *Repository[T]) getCacheDriver() *cache.Cache {
	return repo.cache
}

// GetCache 获取某key的cache，并转化为T对象
func (repo *Repository[T]) GetCache(ctx context.Context, key string) (bool, T, error) {
	return cache.AsModernCache[T](repo.cache).Get(ctx, key)
}

// ForgetCache 删除某keys的cache
func (repo *Repository[T]) ForgetCache(ctx context.Context, keys ...string) error {
	return repo.cache.Forget(ctx, keys...)
}

// GetCacheForList 获取某key的cache，并转化为[]T对象列表
func (repo *Repository[T]) GetCacheForList(ctx context.Context, key string) ([]T, error) {
	_, res, err := cache.AsModernCache[[]T](repo.cache).Get(ctx, key)
	return res, err
}

type rememberCacheGetter[T db.Tabler] struct {
	repository IRepository[T]
	cache      *cache.Cache
	cacheKey   string
}

// Remember 如果有缓存，则返回缓存，不然就执行后面的动作
// **** 注意：(repo *Repository[T]).Remember 在编译时，会非常非常慢，所以改成了这种方式 ****
// 此函数不能单独使用，需要：Remember(repo)("key-1", ...).Get(ctx, cnd.Where(...))
// 如果只想得到Cache，可以使用GetCache、GetCacheForModel、GetCacheForModelList；如果只想设置Cache，可以使用SetCache
// 注意：默认情况下，没有查询到记录（包括Count()==0），不会设置缓存。
// 当options传入WithSaveEmptyOnRemember()，可以强制保存空值
// 如果要修改缓存的过期时间，可以传递WithExpiration()，如果要修改缓存的key前缀，可以WithKeyPrefix()
func Remember[T db.Tabler](repo IRepository[T], key string, options ...cache.Option) IRemember[T] {
	c := &rememberCacheGetter[T]{
		repository: repo,
	}

	c.cacheKey = key
	_cache := repo.getCacheDriver().Clone()
	// 由于cache.Option是用于cache.New的。这里借用这些方法，将只会设置到c.cache.options，之后需要调用WithOptions才会生效
	for _, option := range options {
		option(_cache)
	}
	c.cache = _cache.WithOptions(_cache.GetOptions())

	return c
}

// ------------ rememberCacheGetter ------------

// Do 自定义返回内容
func (c *rememberCacheGetter[T]) Do(ctx context.Context, callback func(context.Context, IRepository[T]) (any, error)) (any, error) {
	return cache.AsModernCache[any](c.cache).Remember(ctx, c.cacheKey, func(ctx context.Context) (any, error) {
		return callback(ctx, c.repository)
	})
}

// Count 统计数量。
// 会先尝试获取缓存，如果没有获取到，则数据库查询，并设置缓存。
func (c *rememberCacheGetter[T]) Count(ctx context.Context, query *cnd.QueryBuilder) (int64, error) {
	return cache.AsModernCache[int64](c.cache).Remember(ctx, c.cacheKey, func(ctx context.Context) (int64, error) {
		return c.repository.Count(ctx, query)
	})
}

// First 通过查询条件获取第一个记录。
// 会先尝试获取缓存，如果没有获取到，则数据库查询，并设置缓存。
// 注意：如果没有查询到记录，不会设置缓存。
func (c *rememberCacheGetter[T]) First(ctx context.Context, query *cnd.QueryBuilder) (T, error) {
	return cache.AsModernCache[T](c.cache).Remember(ctx, c.cacheKey, func(ctx context.Context) (T, error) {
		return c.repository.First(ctx, query)
	})
}

// FirstOrFail 通过查询条件获取第一个记录，没有找到则返回gorm.ErrRecordNotFound错误。
// 会先尝试获取缓存，如果没有获取到，则数据库查询，并设置缓存。
// 注意：如果没有查询到记录，不会设置缓存。
func (c *rememberCacheGetter[T]) FirstOrFail(ctx context.Context, query *cnd.QueryBuilder) (T, error) {
	return cache.AsModernCache[T](c.cache).Remember(ctx, c.cacheKey, func(ctx context.Context) (T, error) {
		return c.repository.FirstOrFail(ctx, query)
	})
}

// Find 通过主键查询第一个资源。
// 会先尝试获取缓存，如果没有获取到，则数据库查询，并设置缓存。
// 注意：如果没有查询到记录，不会设置缓存。
func (c *rememberCacheGetter[T]) Find(ctx context.Context, primary any) (T, error) {
	return cache.AsModernCache[T](c.cache).Remember(ctx, c.cacheKey, func(ctx context.Context) (T, error) {
		return c.repository.Find(ctx, primary)
	})
}

// FindOrFail 通过主键查询第一个资源，没有找到则返回gorm.ErrRecordNotFound错误。
// 会先尝试获取缓存，如果没有获取到，则数据库查询，并设置缓存。
// 注意：如果没有查询到记录，不会设置缓存。
func (c *rememberCacheGetter[T]) FindOrFail(ctx context.Context, primary any) (T, error) {
	return cache.AsModernCache[T](c.cache).Remember(ctx, c.cacheKey, func(ctx context.Context) (T, error) {
		return c.repository.FindOrFail(ctx, primary)
	})
}

// FindMany 通过主键查询多个资源。
// 会先尝试获取缓存，如果没有获取到，则数据库查询，并设置缓存。
// 注意：如果没有查询到记录，不会设置缓存。
func (c *rememberCacheGetter[T]) FindMany(ctx context.Context, primary []any) ([]T, error) {
	return cache.AsModernCache[[]T](c.cache).Remember(ctx, c.cacheKey, func(ctx context.Context) ([]T, error) {
		return c.repository.FindMany(ctx, primary)
	})
}

// Get 通过查询条件获取记录列表。
// 会先尝试获取缓存，如果没有获取到，则数据库查询，并设置缓存。
// 注意：如果没有查询到记录，不会设置缓存。
func (c *rememberCacheGetter[T]) Get(ctx context.Context, query *cnd.QueryBuilder) ([]T, error) {
	return cache.AsModernCache[[]T](c.cache).Remember(ctx, c.cacheKey, func(ctx context.Context) ([]T, error) {
		return c.repository.Get(ctx, query)
	})
}

// Paginate 根据查询条件和分页条件获取记录的分页列表。
// 会先尝试获取缓存，如果没有获取到，则数据库查询，并设置缓存。
// 注意：如果没有查询到记录，不会设置缓存。
func (c *rememberCacheGetter[T]) Paginate(ctx context.Context, query *cnd.QueryBuilder, pagination *db.Pagination) ([]T, error) {
	return cache.AsModernCache[[]T](c.cache).Remember(ctx, c.cacheKey, func(ctx context.Context) ([]T, error) {
		return c.repository.Paginate(ctx, query, pagination)
	})
}
