package cnd

import (
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/clause"
	"gorm.io/gorm"
	"time"
)

type QueryBuilder struct {
	columns       []string
	and, or       []ParamPair  // 参数
	orders        []OrderByCol // 排序
	paging        *Paging      // 分页
	limit, offset *int
	locker        *clause.Locking
	selector      *ParamPair

	preloads  map[string][]any
	withTrash bool
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		preloads: map[string][]any{},
	}
}

// Columns 只获取这些字段
func (q *QueryBuilder) Columns(columns ...string) *QueryBuilder {
	if len(columns) > 0 {
		q.columns = append(q.columns, columns...)
	}
	return q
}

// Select .
func (q *QueryBuilder) Select(query any, args ...any) *QueryBuilder {
	q.selector = &ParamPair{Query: query, Args: args}
	return q
}

// Where 构建where查询条件
func (q *QueryBuilder) Where(query string, args ...any) *QueryBuilder {
	q.and = append(q.and, ParamPair{Query: query, Args: args})
	return q
}

// Or 构建or查询条件
func (q *QueryBuilder) Or(query string, args ...any) *QueryBuilder {
	q.or = append(q.or, ParamPair{Query: query, Args: args})
	return q
}

// Between Where的Between条件
func (q *QueryBuilder) Between(column string, start, end any) *QueryBuilder {
	q.Where(column+" between ? and ?", start, end)
	return q
}

// Eq Where column = arg
func (q *QueryBuilder) Eq(column string, arg any) *QueryBuilder {
	q.Where(column+" = ?", arg)
	return q
}

// Not .
func (q *QueryBuilder) Not(column string, arg any) *QueryBuilder {
	q.Where(column+" <> ?", arg)
	return q
}

// Gt .
func (q *QueryBuilder) Gt(column string, arg any) *QueryBuilder {
	q.Where(column+" > ?", arg)
	return q
}

// Gte .
func (q *QueryBuilder) Gte(column string, arg any) *QueryBuilder {
	q.Where(column+" >= ?", arg)
	return q
}

// Lt .
func (q *QueryBuilder) Lt(column string, arg any) *QueryBuilder {
	q.Where(column+" < ?", arg)
	return q
}

// Lte .
func (q *QueryBuilder) Lte(column string, arg any) *QueryBuilder {
	q.Where(column+" <= ?", arg)
	return q
}

// In .
func (q *QueryBuilder) In(column string, arg any) *QueryBuilder {
	q.Where(column+" in ? ", arg)
	return q
}

// NotIn .
func (q *QueryBuilder) NotIn(column string, arg any) *QueryBuilder {
	q.Where(column+" not in ? ", arg)
	return q
}

// EqDate .
func (q *QueryBuilder) EqDate(column string, t time.Time) *QueryBuilder {
	q.Where("date("+column+") = ?", t.Format("2006-01-02"))
	return q
}

// GtDate .
func (q *QueryBuilder) GtDate(column string, t time.Time) *QueryBuilder {
	q.Where("date("+column+") > ?", t.Format("2006-01-02"))
	return q
}

// GteDate .
func (q *QueryBuilder) GteDate(column string, t time.Time) *QueryBuilder {
	q.Where("date("+column+") >= ?", t.Format("2006-01-02"))
	return q
}

// LtDate .
func (q *QueryBuilder) LtDate(column string, t time.Time) *QueryBuilder {
	q.Where("date("+column+") < ?", t.Format("2006-01-02"))
	return q
}

// LteDate .
func (q *QueryBuilder) LteDate(column string, t time.Time) *QueryBuilder {
	q.Where("date("+column+") <= ?", t.Format("2006-01-02"))
	return q
}

// NeqDate .
func (q *QueryBuilder) NeqDate(column string, t time.Time) *QueryBuilder {
	q.Where("date("+column+") <> ?", t.Format("2006-01-02"))
	return q
}

// Order .
func (q *QueryBuilder) Order(column string, ascend bool) *QueryBuilder {
	q.orders = append(q.orders, OrderByCol{Column: column, Asc: ascend})
	return q
}

// True .
func (q *QueryBuilder) True(column string) *QueryBuilder {
	return q.Eq(column, true)
}

// False .
func (q *QueryBuilder) False(column string) *QueryBuilder {
	return q.Eq(column, false)
}

// Asc .
func (q *QueryBuilder) Asc(column string) *QueryBuilder {
	return q.Order(column, true)
}

// Desc .
func (q *QueryBuilder) Desc(column string) *QueryBuilder {
	return q.Order(column, false)
}

// Limit .
func (q *QueryBuilder) Limit(limit int) *QueryBuilder {
	q.limit = &limit
	return q
}

// Offset .
func (q *QueryBuilder) Offset(offset int) *QueryBuilder {
	q.offset = &offset
	return q
}

// LockingForUpdate 构建lock for update查询
func (q *QueryBuilder) LockingForUpdate(table ...clause.Table) *QueryBuilder {
	q.locker = &clause.Locking{
		Strength: "UPDATE",
	}
	if len(table) > 0 {
		q.locker.Table = table[0]
	}

	return q
}

// LockingForUpdateWithOption .
func (q *QueryBuilder) LockingForUpdateWithOption(opts string, table ...clause.Table) *QueryBuilder {
	q.locker = &clause.Locking{
		Strength: "UPDATE",
		Options:  opts,
	}
	if len(table) > 0 {
		q.locker.Table = table[0]
	}

	return q
}

// LockingForShare 构建lock for share查询
func (q *QueryBuilder) LockingForShare(table ...clause.Table) *QueryBuilder {
	q.locker = &clause.Locking{
		Strength: "SHARE",
	}
	if len(table) > 0 {
		q.locker.Table = table[0]
	}

	return q
}

// LockingForShareWithOption .
func (q *QueryBuilder) LockingForShareWithOption(opts string, table ...clause.Table) *QueryBuilder {
	q.locker = &clause.Locking{
		Strength: "SHARE",
		Options:  opts,
	}
	if len(table) > 0 {
		q.locker.Table = table[0]
	}

	return q
}

// Paginate .
func (q *QueryBuilder) Paginate(page, limit int) *QueryBuilder {
	if q.paging == nil {
		q.paging = &Paging{Page: page, Limit: limit}
	} else {
		q.paging.Page = page
		q.paging.Limit = limit
	}
	return q
}

// PreloadWithBuilder 带条件的预加载，只支持一个关联
func (q *QueryBuilder) PreloadWithBuilder(preload string, args ...any) *QueryBuilder {
	q.preloads[preload] = args
	return q
}

// Preloads 不带条件的预加载，支持多个关联
func (q *QueryBuilder) Preloads(preloads ...string) *QueryBuilder {
	for _, preload := range preloads {
		q.preloads[preload] = nil
	}
	return q
}

// WithTrash .
func (q *QueryBuilder) WithTrash() *QueryBuilder {
	q.withTrash = true
	return q
}

// Build .
func (q *QueryBuilder) Build(db *gorm.DB) *gorm.DB {
	ret := db

	if q.withTrash {
		ret = ret.Unscoped()
	}

	for preload, args := range q.preloads {
		ret = ret.Preload(preload, args...)
	}

	if len(q.columns) > 0 {
		ret = ret.Select(q.columns)
	}

	if len(q.and) > 0 {
		for _, item := range q.and {
			ret = ret.Where(item.Query, item.Args...)
		}
	}

	if len(q.or) > 0 {
		for _, item := range q.or {
			ret = ret.Or(item.Query, item.Args...)
		}
	}

	if len(q.orders) > 0 {
		for _, item := range q.orders {
			if item.Asc {
				ret = ret.Order("`" + item.Column + "` ASC")
			} else {
				ret = ret.Order("`" + item.Column + "` DESC")
			}
		}
	}

	if q.paging != nil {
		ret = ret.Limit(q.paging.Page).Offset(q.paging.Offset())
	}

	if q.limit != nil {
		ret = ret.Limit(*q.limit)
	}

	if q.offset != nil {
		ret = ret.Offset(*q.offset)
	}

	if q.locker != nil {
		ret = ret.Clauses(*q.locker)
	}

	if q.selector != nil {
		ret = ret.Select(q.selector.Query, q.selector.Args...)
	}

	return ret
}
