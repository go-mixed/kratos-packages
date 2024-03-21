package repo

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/cnd"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/event"
)

// Create 批量创建资源
// example: repo.Create(ctx, &User{Name: "tom"}, &User{Name: "jerry"})
// T必须为指针类型
func (repo *Repository[T]) Create(ctx context.Context, models ...T) error {
	// 每次INSERT 100条
	return repo.GetDB(ctx).CreateInBatches(models, 100).Error
}

// Save 保存资源，如果主键为空，则创建，否则更新。注意：零值【会】更新
// 注意：官方建议使用Create、Update，而不是Save。假如期望是Update的行为，但是ID并未设置，这会导致创建了新的记录
// example: repo.Save(ctx, &User{ID: 1, Name: "tom"}, &User{Name: "jerry"})
// T必须为指针类型
func (repo *Repository[T]) Save(ctx context.Context, models ...T) error {
	// Save 不支持[]T，需要遍历
	for _, model := range models {
		if err := repo.GetDB(ctx).Save(model).Error; err != nil {
			return err
		}
	}
	return nil
}

// Update 批量更新资源，注意：零值不会更新。如果没有主键，为了避免批量更新，会返回ErrMissingWhereClaus
// example: repo.Update(ctx, &User{ID: 1, Name: "tom"}, &User{ID: 2, Name: "jerry"})
// T必须为指针类型
func (repo *Repository[T]) Update(ctx context.Context, models ...T) error {
	// Updates 不支持[]T，需要遍历
	for _, model := range models {
		if err := repo.GetDB(ctx).Model(model).Updates(model).Error; err != nil {
			return err
		}
	}
	return nil
}

// Delete 使用model删除资源。如果没有主键，为了避免批量删除，会返回ErrMissingWhereClause
// example: repo.Delete(ctx, &User{ID: 1}, &User{ID: 2})
// T必须为指针类型
func (repo *Repository[T]) Delete(ctx context.Context, models ...T) error {
	return repo.GetDB(ctx).Delete(models).Error
}

// DeleteWithBuilder 使用query删除资源
// example: repo.Delete(ctx, db.ID(1))、或repo.Delete(ctx, cnd.Where("name", "tom"))
func (repo *Repository[T]) DeleteWithBuilder(ctx context.Context, query *cnd.QueryBuilder) error {
	orm := repo.GetDB(ctx).Model(repo.modelCreator())
	// Delete()第一个参数必须是model，即repo.modelCreator()，不然无法绑定Where条件，并且不能在Delete之前设置orm.Model(...)
	if err := query.Build(repo.GetDB(ctx)).Delete(repo.modelCreator()).Error; err != nil {
		return err
	}

	return repo.onModelEvent(ctx, orm, nil, event.BatchDeleted, query)
}

// DeletePrimary 通过主键删除资源
// example: repo.DeletePrimary(ctx, 1, 2, 3)
func (repo *Repository[T]) DeletePrimary(ctx context.Context, primary ...any) error {
	orm := repo.GetDB(ctx).Model(repo.modelCreator())
	// Delete() 第一个参数必须是model，并且不能在Delete之前设置orm.Model(...)
	if err := repo.GetDB(ctx).Delete(repo.modelCreator(), primary).Error; err != nil {
		return err
	}

	return repo.onModelEvent(ctx, orm, nil, event.BatchDeleted, cnd.InID(primary))
}

// UpdateColumns 更新资源多个字段
func (repo *Repository[T]) UpdateColumns(ctx context.Context, query *cnd.QueryBuilder, attributes Columns) error {
	orm := repo.GetDB(ctx).Model(repo.modelCreator())
	// 和Delete不同的是，需要在Updates之前设置orm.Model(...)
	if err := query.Build(orm).Updates(attributes).Error; err != nil {
		return errors.Wrapf(err, "repo UpdateColumns method of table \"%s\" failed", repo.modelCreator().TableName())
	}

	return repo.onModelEvent(ctx, orm, nil, event.BatchUpdated, query, attributes)
}

// UpdateColumn 更新资源单个字段
func (repo *Repository[T]) UpdateColumn(ctx context.Context, query *cnd.QueryBuilder, key string, value any) error {
	orm := repo.GetDB(ctx).Model(repo.modelCreator())
	// 和Delete不同的是，需要在Update之前设置orm.Model(...)
	if err := query.Build(orm).Update(key, value).Error; err != nil {
		return errors.Wrapf(err, "repo UpdateColumn method of table \"%s\" failed", repo.modelCreator().TableName())
	}

	return repo.onModelEvent(ctx, orm, nil, event.BatchUpdated, query, Columns{key: value})
}

// Incr 递增某字段
func (repo *Repository[T]) Incr(ctx context.Context, query *cnd.QueryBuilder, field string, val any) error {
	if err := query.Build(repo.GetDB(ctx).Model(repo.modelCreator())).
		Update(
			field,
			db.Expr(fmt.Sprintf("%s + ?", field), val),
		).Error; err != nil {
		return errors.Wrapf(err, "repo Incr method of table \"%s\" failed", repo.modelCreator().TableName())
	}
	return nil
}

// Decr 递减某字段
func (repo *Repository[T]) Decr(ctx context.Context, query *cnd.QueryBuilder, field string, val any) error {
	if err := query.Build(repo.GetDB(ctx).Model(repo.modelCreator())).
		Update(
			field,
			db.Expr(fmt.Sprintf("%s - ?", field), val),
		).Error; err != nil {
		return errors.Wrapf(err, "repo Decr method of table \"%s\" failed", repo.modelCreator().TableName())
	}
	return nil
}
