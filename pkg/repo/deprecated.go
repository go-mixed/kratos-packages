package repo

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/cnd"
	"sync"

	"gorm.io/gorm/schema"

	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
)

// Deprecated: 请使用IRepository[T]接口
// BaseRepository 基础repo
type BaseRepository interface {
	// Transaction 开启事务
	// example:
	// repo.Transaction(ctx, func(ctx context.Context) error {
	//     return repo.First(ctx, db.ID(1).LockingForUpdate())
	//}, func(ctx context.Context) error {
	//	   return repo.Create(ctx, &model)
	//}...)
	Transaction(ctx context.Context, steps ...func(ctx context.Context) error) error

	// GetDB 获取db 从上下文中取出db
	GetDB(ctx context.Context) *db.DB

	// Create 创建资源
	Create(ctx context.Context, value interface{}) error
	// Delete 删除资源
	// example:
	// 使用model删除 repo.Delete(ctx, &User{})
	// 使用query删除 repo.Delete(ctx, db.ID(1))、repo.Delete(ctx, db.Where("name", "urionz"))
	Delete(ctx context.Context, queryOrModel ...interface{}) error

	// DeleteByPrimary 通过主键删除资源
	// example:
	// repo.DeleteByPrimary(ctx, 1)
	DeleteByPrimary(ctx context.Context, primary interface{}) error

	// UpdateColumns 更新资源多个字段
	UpdateColumns(ctx context.Context, query *cnd.QueryBuilder, attributes map[string]interface{}, withDeleted ...bool) error
	// UpdateColumn 更新资源单个字段
	UpdateColumn(ctx context.Context, query *cnd.QueryBuilder, key string, value interface{}, withDeleted ...bool) error

	// Count 查询资源数量。于2023-06-14 dxg修改，原来的Count没返回error，会导致无法分辨是报错还是真的为0
	// query 查询条件
	// withDeleted 是否查询软删除资源
	Count(ctx context.Context, query *cnd.QueryBuilder, withDeleted ...bool) (int64, error)

	// First 查询第一个资源，如果没有找到【不会】gorm.ErrRecordNotFound，但是scanner仍然可能被空值实例化
	// query 查询条件
	// scanner 结构指针
	// withDeleted 是否查询软删除资源
	First(ctx context.Context, query *cnd.QueryBuilder, scanner interface{}, withDeleted ...bool) error

	// Latest 查询第一个最新资源，如果没有找到【不会】gorm.ErrRecordNotFound，但是scanner仍然可能被空值实例化
	// query 查询条件
	// scanner 结构指针
	// withDeleted 是否查询软删除资源
	Latest(ctx context.Context, query *cnd.QueryBuilder, scanner interface{}, withDeleted ...bool) error

	// Oldest 查询第一个最旧资源，如果没有找到【不会】gorm.ErrRecordNotFound，但是scanner仍然可能被空值实例化
	// query 查询条件
	// scanner 结构指针
	// withDeleted 是否查询软删除资源
	Oldest(ctx context.Context, query *cnd.QueryBuilder, scanner interface{}, withDeleted ...bool) error

	// FirstByPrimary 通过主键查询第一个资源，如果没有找到【不会】gorm.ErrRecordNotFound，但是scanner仍然可能被空值实例化
	// primary 主键值
	// scanner 结构指针
	// withDeleted 是否查询软删除资源
	FirstByPrimary(ctx context.Context, primary interface{}, scanner interface{}, withDeleted ...bool) error

	// Get 查询获取资源集合，如果没有找到【不会】gorm.ErrRecordNotFound，但是scanner仍然可能被空值实例化
	// query 查询条件
	// scanner 结构指针
	// withDeleted 是否查询软删除资源
	Get(ctx context.Context, query *cnd.QueryBuilder, scanner interface{}, withDeleted ...bool) error

	// Pluck 获取资源单个字段集合，如果没有找到【不会】gorm.ErrRecordNotFound
	Pluck(ctx context.Context, queries *cnd.QueryBuilder, field string, scanner interface{}, withDeleted ...bool) error

	// Paginate 资源分页
	Paginate(ctx context.Context, query *cnd.QueryBuilder, pagination *db.Pagination, scanner interface{}, withDeleted ...bool) error

	// IncrBy 递增某字段
	IncrBy(ctx context.Context, query *cnd.QueryBuilder, field string, steps ...int) error

	// DecrBy 递减某字段
	DecrBy(ctx context.Context, query *cnd.QueryBuilder, field string, steps ...int) error
}

var _ BaseRepository = (*BaseRepo)(nil)

type BaseRepo struct {
	mu          *sync.RWMutex
	db          *db.DB
	createModel func() schema.Tabler
	logger      *log.Helper
}

// Deprecated: 请使用NewRepository[T]函数
func NewRepo(db *db.DB, modelCreator func() schema.Tabler, logger log.Logger) *BaseRepo {
	return &BaseRepo{
		mu:          &sync.RWMutex{},
		db:          db,
		createModel: modelCreator,
		logger:      log.NewModuleHelper(logger, "repo/base"),
	}
}

type transactionKey struct{}

// newTxContext 构造事务context
func newTxContext(ctx context.Context, value *db.DB) context.Context {
	return context.WithValue(ctx, &transactionKey{}, value)
}

// GetDB 获取db，如果是事务，并将ctx附加到gorm中
func (repo *BaseRepo) GetDB(ctx context.Context) *db.DB {
	if ctx != nil {
		if tx := ctx.Value(&transactionKey{}); tx != nil {
			return tx.(*db.DB).WithContext(ctx)
		}
	}
	return repo.db.WithContext(ctx)
}

// Transaction 开启事务
func (repo *BaseRepo) Transaction(ctx context.Context, steps ...func(ctx context.Context) error) error {
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

// Create 创建资源
func (repo *BaseRepo) Create(ctx context.Context, value interface{}) error {
	return repo.GetDB(ctx).Model(repo.createModel()).Create(value).Error
}

// Delete 删除资源
func (repo *BaseRepo) Delete(ctx context.Context, queryOrModel ...interface{}) error {
	var err error
	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(queryOrModel) == 0 {
		return errors.New("请至少提供一种删除方式")
	}
	for _, method := range queryOrModel {
		switch t := method.(type) {
		case schema.Tabler:
			err = orm.Delete(orm).Error
		case *cnd.QueryBuilder:
			err = t.Build(orm).Delete(repo.createModel()).Error
		default:
			return errors.New("暂不支持此类型的删除方式")
		}
	}

	if err != nil {
		err = errors.Wrap(err, "repo Delete method failed")
	}

	return err
}

// DeleteByPrimary 通过主键删除资源
func (repo *BaseRepo) DeleteByPrimary(ctx context.Context, primary interface{}) error {
	if err := repo.GetDB(ctx).Delete(repo.createModel(), primary).Error; err != nil {
		return errors.Wrap(err, "repo DeleteByPrimary method failed")
	}
	return nil
}

// UpdateColumns 更新资源多个字段
func (repo *BaseRepo) UpdateColumns(ctx context.Context, query *cnd.QueryBuilder, attributes map[string]interface{}, withDeleted ...bool) error {
	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(withDeleted) > 0 {
		orm = orm.Unscoped()
	}
	if err := query.Build(orm).Updates(attributes).Error; err != nil {
		return errors.Wrap(err, "repo UpdateColumns method failed")
	}
	return nil
}

// UpdateColumn 更新资源单个字段
func (repo *BaseRepo) UpdateColumn(ctx context.Context, query *cnd.QueryBuilder, key string, value interface{}, withDeleted ...bool) error {
	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(withDeleted) > 0 {
		orm = orm.Unscoped()
	}
	if err := query.Build(orm).Update(key, value).Error; err != nil {
		return errors.Wrap(err, "repo UpdateColumn method failed")
	}
	return nil
}

// Count 查询资源数量，但是数据库错误了也返回0，则会让程序
func (repo *BaseRepo) Count(ctx context.Context, query *cnd.QueryBuilder, withDeleted ...bool) (count int64, err error) {
	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(withDeleted) > 0 {
		orm = orm.Unscoped()
	}
	err = query.Build(orm).Count(&count).Error
	return count, err
}

// Latest 查询第一个最新资源，如果没有找到【不会】返回ErrRecordNotFound
func (repo *BaseRepo) Latest(ctx context.Context, query *cnd.QueryBuilder, scanner interface{}, withDeleted ...bool) error {
	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(withDeleted) > 0 {
		orm = orm.Unscoped()
	}
	if err := query.Build(orm).Order("created_at desc").First(scanner).Error; err != nil && !errors.Is(err, db.ErrRecordNotFound) {
		return errors.Wrap(err, "repo Latest method failed")
	}
	return nil
}

// Oldest 查询第一个最旧资源，如果没有找到【不会】返回ErrRecordNotFound
func (repo *BaseRepo) Oldest(ctx context.Context, query *cnd.QueryBuilder, scanner interface{}, withDeleted ...bool) error {
	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(withDeleted) > 0 {
		orm = orm.Unscoped()
	}
	if err := query.Build(orm).Order("created_at asc").First(scanner).Error; err != nil && !errors.Is(err, db.ErrRecordNotFound) {
		return errors.Wrap(err, "repo Oldest method failed")
	}
	return nil
}

// First 查询第一个资源，如果没有找到【不会】返回ErrRecordNotFound
func (repo *BaseRepo) First(ctx context.Context, query *cnd.QueryBuilder, scanner interface{}, withDeleted ...bool) error {
	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(withDeleted) > 0 {
		orm = orm.Unscoped()
	}
	if err := query.Build(orm).First(scanner).Error; err != nil && !errors.Is(err, db.ErrRecordNotFound) {
		return errors.Wrap(err, "repo First method failed")
	}
	return nil
}

// FirstByPrimary 通过主键查询第一个资源，如果没有找到【不会】返回ErrRecordNotFound
func (repo *BaseRepo) FirstByPrimary(ctx context.Context, primary interface{}, scanner interface{}, withDeleted ...bool) error {
	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(withDeleted) > 0 {
		orm = orm.Unscoped()
	}
	if err := orm.First(scanner, primary).Error; err != nil && !errors.Is(err, db.ErrRecordNotFound) {
		return errors.Wrap(err, "repo FirstByPrimary method failed")
	}
	return nil
}

// Get 查询获取资源集合，如果没有找到【不会】返回ErrRecordNotFound
func (repo *BaseRepo) Get(ctx context.Context, query *cnd.QueryBuilder, scanner interface{}, withDeleted ...bool) error {
	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(withDeleted) > 0 {
		orm = orm.Unscoped()
	}
	if err := query.Build(orm).Find(scanner).Error; err != nil && !errors.Is(err, db.ErrRecordNotFound) {
		return errors.Wrap(err, "repo Get method failed")
	}
	return nil
}

// Pluck 获取资源单个字段集合，如果没有找到【不会】返回ErrRecordNotFound
func (repo *BaseRepo) Pluck(ctx context.Context, queries *cnd.QueryBuilder, field string, scanner interface{}, withDeleted ...bool) error {
	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(withDeleted) > 0 {
		orm = orm.Unscoped()
	}
	if err := queries.Build(orm).Pluck(field, scanner).Error; err != nil && !errors.Is(err, db.ErrRecordNotFound) {
		return errors.Wrap(err, "repo Pluck method failed")
	}

	return nil
}

// Paginate 资源分页，如果没有找到【不会】返回ErrRecordNotFound
func (repo *BaseRepo) Paginate(ctx context.Context, query *cnd.QueryBuilder, pagination *db.Pagination, scanner interface{}, withDeleted ...bool) error {
	var total int64
	var err error

	orm := repo.GetDB(ctx).Model(repo.createModel())
	if len(withDeleted) > 0 {
		orm = orm.Unscoped()
	}

	cond := query.Build(orm)

	if err = cond.Count(&total).Error; err != nil && !errors.Is(err, db.ErrRecordNotFound) {
		return errors.Wrap(err, "repo Paginate method get count failed")
	}

	pagination.Total = total

	if err = cond.Limit(pagination.Limit).Offset(pagination.GetOffset()).Find(scanner).Error; err != nil && !errors.Is(err, db.ErrRecordNotFound) {
		return errors.Wrap(err, "repo Paginate method failed")
	}

	return nil
}

// IncrBy 递增某字段
func (repo *BaseRepo) IncrBy(ctx context.Context, query *cnd.QueryBuilder, field string, steps ...int) error {
	step := 1
	if len(steps) > 0 {
		step = steps[0]
	}
	if err := query.Build(repo.GetDB(ctx).Model(repo.createModel())).
		Update(
			field,
			db.Expr(fmt.Sprintf("%s + %d", field, step)),
		).Error; err != nil {
		return errors.Wrap(err, "repo Incr method failed")
	}
	return nil
}

// DecrBy 递减某字段
func (repo *BaseRepo) DecrBy(ctx context.Context, query *cnd.QueryBuilder, field string, steps ...int) error {
	step := 1
	if len(steps) > 0 {
		step = steps[0]
	}
	if err := query.Build(repo.GetDB(ctx).Model(repo.createModel())).
		Update(
			field,
			db.Expr(fmt.Sprintf("%s - %d", field, step)),
		).Error; err != nil {
		return errors.Wrap(err, "repo Decr method failed")
	}
	return nil
}
