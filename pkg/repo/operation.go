package repo

import (
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/clause"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/cnd"
)

type repositoryOperations[T db.Tabler] struct {
	clauses   []clause.Expression
	selectors []db.ParamPair
	omits     []string
	preloads  map[string][]any
}

// Clauses 添加条件：冲突解决、读写分离等
//   - 比如：Clauses(clause.Write).Find(ctx, cnd.ID(1))，表示强制使用主库查询
func (repo *Repository[T]) Clauses(cnds ...clause.Expression) IOrm[T] {
	_repo := repo.Clone()
	_repo.operations.clauses = append(_repo.operations.clauses, cnds...)
	return _repo
}

// Select 指定查询/更新/创建字段。比如：Select("name", "age").Find(ctx, cnd.Eq("id", 1))，表示只查询name和age字段
//
//	FieldName：如果后面传递是Struct，使用Struct的字段名；如果后面传递是map，使用map的key
func (repo *Repository[T]) Select(query any, args ...any) IOrm[T] {
	_repo := repo.Clone()
	_repo.operations.selectors = append(repo.operations.selectors, db.ParamPair{Query: query, Args: args})
	return _repo
}

// Only 同Select函数，但是参数是字段名
func (repo *Repository[T]) Only(columns ...string) IOrm[T] {
	if len(columns) == 0 {
		return repo
	}
	return repo.Select(columns[0], lo.Map(columns[1:], func(v string, _ int) any { return v })...)
}

// Omit 排除字段。比如：Omit("name", "age").Find(ctx, cnd.Eq("id", 1))，表示排除name和age字段
//
//	FieldName：如果后面传递是Struct，使用Struct的字段名；如果后面传递是map，使用map的key
func (repo *Repository[T]) Omit(columns ...string) IOrm[T] {
	_repo := repo.Clone()
	_repo.operations.omits = append(_repo.operations.omits, columns...)
	return _repo
}

// PreloadWithBuilder 带条件的预加载，一次只能设置一个关联
func (repo *Repository[T]) PreloadWithBuilder(preload string, args ...any) IOrm[T] {
	_repo := repo.Clone()
	_repo.operations.preloads[preload] = args
	return _repo
}

// Preloads 不带条件的预加载关联，可以一次设置多个关联名
func (repo *Repository[T]) Preloads(preloads ...string) IOrm[T] {
	_repo := repo.Clone()
	for _, preload := range preloads {
		_repo.operations.preloads[preload] = nil
	}
	return _repo
}

// 构造ORM
func (repo *Repository[T]) buildOrm(db *db.DB, bindModel db.Tabler, query *cnd.QueryBuilder) *db.DB {
	orm := db
	// GORM的参数顺序为：
	//
	// 创建
	// db.Session().Clauses().Select().Omit().Create(&input)
	// db.Session().Clauses().Model(&blankModel).Select().Omit().Create(map{})
	//
	// 批量创建
	// db.Session().Clauses().Select().Omit().Create(&[]input | []*input)，或CreateInBatches([]input, 100)
	// db.Session().Clauses().Model(&blankModel).Select().Omit().Create([]map{})
	//
	// 更新，WHERE条件是：modelWithPk.Pk + Where
	// db.Session().Clauses().Model(&blankModel | &modelWithPk).Where().Select().Omit().Updates(&input | map{})
	//
	// 更新单个字段，WHERE条件是：modelWithPk.Pk + Where
	// db.Session().Clauses().Model(&blankModel | &modelWithPk).Where().Update("field", value)，或UpdateColumn("field", value)
	//
	// 删除
	// db.Session().Clauses().Model(&blankModel).Where().Delete(&[]return)
	//
	// 按主键删除
	// db.Session().Clauses().Model(&blankModel).Delete(&[]return, primaryKeys...)
	//
	// 查找多条记录
	// db.Session().Clauses().Model(&blankModel).Where().Select().Omit().Preload().Find(&[]return)
	//
	// 查找单条记录
	// db.Session().Clauses().Model(&blankModel).Where().Select().Omit().Preload().First(&return)

	// 先添加Clauses
	if len(repo.operations.clauses) > 0 {
		orm = orm.Clauses(repo.operations.clauses...)
	}

	// 绑定空白模型
	if bindModel != nil {
		orm = orm.Model(bindModel)
	}

	// 绑定Where、Order、Limit、Offset等条件
	if query != nil {
		orm = query.Build(orm)
	}

	for _, selector := range repo.operations.selectors {
		orm = orm.Select(selector.Query, selector.Args...)
	}
	for _, omit := range repo.operations.omits {
		orm = orm.Omit(omit)
	}

	for preload, args := range repo.operations.preloads {
		orm = orm.Preload(preload, args...)
	}

	return orm
}
