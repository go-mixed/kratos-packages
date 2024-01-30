package cnd

import (
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db/clause"
	"time"
)

// ID 构造字段为id条件的便捷操作
func ID(value any) *QueryBuilder {
	return NewQueryBuilder().Eq("id", value)
}

// Columns 构造select columns便捷操作
func Columns(columns ...string) *QueryBuilder {
	return NewQueryBuilder().Columns(columns...)
}

// Select 构造select的便捷操作
func Select(query any, args ...any) *QueryBuilder {
	return NewQueryBuilder().Select(query, args...)
}

// Where 构造where条件的便捷操作
func Where(query string, args ...any) *QueryBuilder {
	return NewQueryBuilder().Where(query, args...)
}

// Or 构造or条件的便捷操作
func Or(query string, arg any) *QueryBuilder {
	return NewQueryBuilder().Or(query, arg)
}

// Eq 构造where eq条件的便捷操作
func Eq(query string, arg any) *QueryBuilder {
	return NewQueryBuilder().Eq(query, arg)
}

// Not 构造not条件的便捷操作
func Not(query string, arg any) *QueryBuilder {
	return NewQueryBuilder().Not(query, arg)
}

// NotIn 构造not in条件的便捷操作
func NotIn(query string, arg any) *QueryBuilder {
	return NewQueryBuilder().NotIn(query, arg)
}

// NotInID 构造not in条件的便捷操作
func NotInID(args any) *QueryBuilder {
	return NotIn("id", args)
}

// Gt .
func Gt(column string, arg any) *QueryBuilder {
	return NewQueryBuilder().Gt(column, arg)
}

// Gte .
func Gte(column string, arg any) *QueryBuilder {
	return NewQueryBuilder().Gte(column, arg)
}

// Lt .
func Lt(column string, arg any) *QueryBuilder {
	return NewQueryBuilder().Lt(column, arg)
}

// Lte .
func Lte(column string, arg any) *QueryBuilder {
	return NewQueryBuilder().Lte(column, arg)
}

// In .
func In(column string, args any) *QueryBuilder {
	return NewQueryBuilder().In(column, args)
}

// InID .
func InID(args any) *QueryBuilder {
	return NewQueryBuilder().In("id", args)
}

// EqDate .
func EqDate(column string, t time.Time) *QueryBuilder {
	return NewQueryBuilder().EqDate(column, t)
}

// GtDate .
func GtDate(column string, t time.Time) *QueryBuilder {
	return NewQueryBuilder().GtDate(column, t)
}

// GteDate .
func GteDate(column string, t time.Time) *QueryBuilder {
	return NewQueryBuilder().GteDate(column, t)
}

// LtDate .
func LtDate(column string, t time.Time) *QueryBuilder {
	return NewQueryBuilder().LtDate(column, t)
}

// LteDate .
func LteDate(column string, t time.Time) *QueryBuilder {
	return NewQueryBuilder().LteDate(column, t)
}

// NeqDate .
func NeqDate(column string, t time.Time) *QueryBuilder {
	return NewQueryBuilder().NeqDate(column, t)
}

// Order .
func Order(column string, ascend bool) *QueryBuilder {
	return NewQueryBuilder().Order(column, ascend)
}

// True .
func True(column string) *QueryBuilder {
	return NewQueryBuilder().True(column)
}

// False .
func False(column string) *QueryBuilder {
	return NewQueryBuilder().False(column)
}

// Asc 构造order asc的便捷操作
func Asc(value string) *QueryBuilder {
	return NewQueryBuilder().Asc(value)
}

// Desc .
func Desc(column string) *QueryBuilder {
	return NewQueryBuilder().Desc(column)
}

// Limit .
func Limit(limit int) *QueryBuilder {
	return NewQueryBuilder().Limit(limit)
}

// Offset .
func Offset(offset int) *QueryBuilder {
	return NewQueryBuilder().Offset(offset)
}

// LockingForUpdate .
func LockingForUpdate(table ...clause.Table) *QueryBuilder {
	return NewQueryBuilder().LockingForUpdate(table...)
}

// LockingForUpdateWithOption .
func LockingForUpdateWithOption(opts string, table ...clause.Table) *QueryBuilder {
	return NewQueryBuilder().LockingForUpdateWithOption(opts, table...)
}

// LockingForShare .
func LockingForShare(table ...clause.Table) *QueryBuilder {
	return NewQueryBuilder().LockingForShare(table...)
}

// LockingForShareWithOption .
func LockingForShareWithOption(opts string, table ...clause.Table) *QueryBuilder {
	return NewQueryBuilder().LockingForShareWithOption(opts, table...)
}

// Paginate .
func Paginate(page, limit int) *QueryBuilder {
	return NewQueryBuilder().Paginate(page, limit)
}

// WithTrash 软删除的数据也会被查询出来
func WithTrash() *QueryBuilder {
	return NewQueryBuilder().WithTrash()
}

// PreloadWithBuilder 带条件的预加载，只支持一个关联
func PreloadWithBuilder(preload string, args ...any) *QueryBuilder {
	return NewQueryBuilder().PreloadWithBuilder(preload, args...)
}

// Preloads 多个不带条件的预加载，支持多个关联
func Preloads(preloads ...string) *QueryBuilder {
	return NewQueryBuilder().Preloads(preloads...)
}
