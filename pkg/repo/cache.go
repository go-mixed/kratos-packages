package repo

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/cache"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/cnd"
)

type rememberCacheGetter[T db.Tabler] struct {
	repository *Repository[T]
	cache      *cache.Cache
	cacheKey   string
}

func (repo *Repository[T]) Remember(key string, options ...cache.Option) IRemember[T] {
	c := &rememberCacheGetter[T]{
		repository: repo,
		cacheKey:   key,
	}

	_cache := repo.cache.Clone()
	// 由于cache.Option是用于cache.New的。这里借用这些方法，将只会设置到c.cache.options，之后需要调用WithOptions才会生效
	for _, option := range options {
		option(_cache)
	}
	c.cache = _cache.WithOptions(_cache.GetOptions())

	return c
}

func (repo *Repository[T]) GetCache(ctx context.Context, key string) (bool, T, error) {
	return cache.AsModernCache[T](repo.cache).Get(ctx, key)
}

func (repo *Repository[T]) GetCacheForList(ctx context.Context, key string) ([]T, error) {
	_, res, err := cache.AsModernCache[[]T](repo.cache).Get(ctx, key)
	return res, err
}

// ------------ rememberCacheGetter ------------

// Do 自定义返回内容
func (c *rememberCacheGetter[T]) Do(ctx context.Context, callback func(context.Context, *Repository[T]) (any, error)) (any, error) {
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
