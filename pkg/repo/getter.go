package repo

import (
	"context"
	"github.com/pkg/errors"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/cnd"
)

// Count 查询资源数量，但是数据库错误了也返回0，则会让程序
func (repo *Repository[T]) Count(ctx context.Context, query *cnd.QueryBuilder) (count int64, err error) {
	orm := repo.GetDB(ctx).Model(repo.modelCreator())
	err = query.Build(orm).Count(&count).Error
	return count, errors.Wrapf(err, "repo Count method of table \"%s\" failed", repo.modelCreator().TableName())
}

// First 查询第一个资源，如果没有找到【不会】返回ErrRecordNotFound
func (repo *Repository[T]) First(ctx context.Context, query *cnd.QueryBuilder) (T, error) {
	var model T
	var nilModel T
	orm := repo.GetDB(ctx).Model(repo.modelCreator())

	if err := query.Build(orm).First(&model).Error; err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nilModel, nil
		}
		return nilModel, errors.Wrapf(err, "repo First method of table \"%s\" failed", repo.modelCreator().TableName())
	}
	return model, nil
}

// Get 查询获取资源集合，如果没有找到【不会】返回ErrRecordNotFound
func (repo *Repository[T]) Get(ctx context.Context, query *cnd.QueryBuilder) ([]T, error) {
	var models []T
	orm := repo.GetDB(ctx).Model(repo.modelCreator())

	if err := query.Build(orm).Find(&models).Error; err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "repo Get method of table \"%s\" failed", repo.modelCreator().TableName())
	}
	return models, nil
}

// Pluck 获取资源单个字段集合，如果没有找到【不会】返回ErrRecordNotFound
func (repo *Repository[T]) Pluck(ctx context.Context, queries *cnd.QueryBuilder, field string, scanner any) error {
	orm := repo.GetDB(ctx).Model(repo.modelCreator())

	if err := queries.Build(orm).Pluck(field, scanner).Error; err != nil && !errors.Is(err, db.ErrRecordNotFound) {
		return errors.Wrapf(err, "repo Pluck method of table \"%s\" failed", repo.modelCreator().TableName())
	}

	return nil
}

// Paginate 资源分页，如果没有找到【不会】返回ErrRecordNotFound
func (repo *Repository[T]) Paginate(ctx context.Context, query *cnd.QueryBuilder, pagination *db.Pagination) ([]T, error) {
	var total int64
	var err error
	var models []T

	orm := repo.GetDB(ctx).Model(repo.modelCreator())

	cond := query.Build(orm)

	if err = cond.Count(&total).Error; err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "repo Paginate method of table \"%s\" get count failed", repo.modelCreator().TableName())
	}

	pagination.Total = total

	if err = cond.Limit(pagination.Limit).Offset(pagination.GetOffset()).Find(&models).Error; err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "repo Paginate method of table \"%s\" get data failed", repo.modelCreator().TableName())
	}

	return models, nil
}

// Find 查询id，如果没有找到【不会】返回ErrRecordNotFound
func (repo *Repository[T]) Find(ctx context.Context, id any) (T, error) {
	var model T
	var nilModel T

	if err := repo.GetDB(ctx).Model(repo.modelCreator()).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nilModel, nil
		}
		return nilModel, errors.Wrapf(err, "repo Find method of table [%s:%d] failed", repo.modelCreator().TableName(), id)
	}
	return model, nil
}

// FindOrFail 查询id，如果没有找到返回ErrRecordNotFound错误
func (repo *Repository[T]) FindOrFail(ctx context.Context, id any) (T, error) {
	var model T
	var nilModel T

	if err := repo.GetDB(ctx).Model(repo.modelCreator()).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nilModel, errors.Wrapf(err, "table [%s:%d] is not found", repo.modelCreator().TableName(), id)
		}
		return nilModel, errors.Wrapf(err, "repo FindOrFail method of table [%s:%d] failed", repo.modelCreator().TableName(), id)
	}
	return model, nil
}

// FindMany 查询获取ids的资源集合
func (repo *Repository[T]) FindMany(ctx context.Context, ids []any) ([]T, error) {
	var models []T
	if err := repo.GetDB(ctx).Model(repo.modelCreator()).Where("id in ?", ids).Find(&models).Error; err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "repo FindMany method of table \"%s\" failed", repo.modelCreator().TableName())
	}
	return models, nil
}
