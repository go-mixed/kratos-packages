package cnd

import (
	"fmt"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"reflect"
	"strings"
)

type Operators map[string]Operator

type Operator func(query *QueryBuilder, col string, val any) error

func newStdOperator(operator string) Operator {
	return func(query *QueryBuilder, col string, val any) error {
		query.Where(fmt.Sprintf("%s %s ?", col, operator), val)
		return nil
	}
}

func newLikeOperator(operator string) Operator {
	return func(query *QueryBuilder, col string, val any) error {
		query.Where(fmt.Sprintf("%s %s ?", col, operator), fmt.Sprintf("%%%s%%", val))
		return nil
	}
}

func newBetweenOperator(operator string) Operator {
	return func(query *QueryBuilder, col string, val any) error {
		refVal := reflect.ValueOf(val)
		if refVal.Kind() != reflect.Slice || refVal.Len() < 2 {
			return fmt.Errorf("between operator need two values")
		}
		query.Where(fmt.Sprintf("%s %s ? AND ?", col, operator), refVal.Index(0).Interface(), refVal.Index(1).Interface())
		return nil
	}
}

// OperatorEq WHERE column = value
func OperatorEq(query *QueryBuilder, col string, val any) error {
	return newStdOperator("=")(query, col, val)
}

// OperatorNe WHERE column <> value
func OperatorNe(query *QueryBuilder, col string, val any) error {
	return newStdOperator("<>")(query, col, val)
}

// OperatorGt WHERE column > value
func OperatorGt(query *QueryBuilder, col string, val any) error {
	return newStdOperator(">")(query, col, val)
}

// OperatorGte WHERE column >= value
func OperatorGte(query *QueryBuilder, col string, val any) error {
	return newStdOperator(">=")(query, col, val)
}

// OperatorLt WHERE column < value
func OperatorLt(query *QueryBuilder, col string, val any) error {
	return newStdOperator("<")(query, col, val)
}

// OperatorLte WHERE column <= value
func OperatorLte(query *QueryBuilder, col string, val any) error {
	return newStdOperator("<=")(query, col, val)
}

// OperatorIn WHERE column IN (value1, value2, ...)
func OperatorIn(query *QueryBuilder, col string, val any) error {
	return newStdOperator("in")(query, col, val)
}

// OperatorNotIn WHERE column NOT IN (value1, value2, ...)
func OperatorNotIn(query *QueryBuilder, col string, val any) error {
	return newStdOperator("not in")(query, col, val)
}

// OperatorLike WHERE column LIKE value
func OperatorLike(query *QueryBuilder, col string, val any) error {
	return newLikeOperator("like")(query, col, val)
}

// OperatorNotLike WHERE column NOT LIKE value
func OperatorNotLike(query *QueryBuilder, col string, val any) error {
	return newLikeOperator("not like")(query, col, val)
}

// OperatorAnyLike WHERE (column1 LIKE value OR column2 LIKE value OR ...)
func OperatorAnyLike(columns ...string) Operator {
	sb := strings.Builder{}
	sb.WriteString("(")
	for i, col := range columns {
		if i != 0 {
			sb.WriteString(" OR ")
		}
		sb.WriteString(col + " like ?")
	}
	sb.WriteString(")")
	return func(query *QueryBuilder, col string, val any) error {
		_val := fmt.Sprintf("%%%s%%", val)
		args := lo.RepeatBy(len(columns), func(_ int) any {
			return _val
		})

		query.Where(sb.String(), args...)
		return nil
	}
}

// OperatorAnyEq WHERE (column1 = value OR column2 = value OR ...)
func OperatorAnyEq(columns ...string) Operator {
	sb := strings.Builder{}
	sb.WriteString("(")
	for i, col := range columns {
		if i != 0 {
			sb.WriteString(" OR ")
		}
		sb.WriteString(col + " = ?")
	}
	sb.WriteString(")")
	return func(query *QueryBuilder, col string, val any) error {
		args := lo.RepeatBy(len(columns), func(_ int) any {
			return val
		})
		query.Where(sb.String(), args...)
		return nil
	}
}

// OperatorBetween WHERE column BETWEEN value0 AND value1
func OperatorBetween(query *QueryBuilder, col string, val any) error {
	return newBetweenOperator("between")(query, col, val)
}

// OperatorNotBetween WHERE column NOT BETWEEN value0 AND value1
func OperatorNotBetween(query *QueryBuilder, col string, val any) error {
	return newBetweenOperator("not between")(query, col, val)
}

// OperatorBetweenDate WHERE column >= value0 AND column <= value1
func OperatorBetweenDate(query *QueryBuilder, col string, val any) error {
	dateRange, ok := val.([]string)
	if !ok || len(dateRange) < 2 {
		return fmt.Errorf("between date operator need two values")
	}
	if dateRange[0] != "" && utils.IsDate(dateRange[0]) {
		query.Where(col+" >= ?", dateRange[0])
	}
	if dateRange[1] != "" && utils.IsDate(dateRange[1]) {
		query.Where(col+" <= ?", dateRange[1])
	}
	return nil
}

// OperatorBetweenDateTime WHERE column >= time0 AND column <= time1
func OperatorBetweenDateTime(query *QueryBuilder, col string, val any) error {
	dateRange, ok := val.([]string)
	if !ok || len(dateRange) < 2 {
		return fmt.Errorf("between date time operator need two values")
	}
	if dateRange[0] != "" && utils.IsDateTimeOrDate(dateRange[0]) {
		// only date, add time
		if utils.IsDate(dateRange[0]) {
			dateRange[0] += " 00:00:00"
		}
		query.Where(col+" >= ?", dateRange[0])
	}
	if dateRange[1] != "" && utils.IsDateTimeOrDate(dateRange[1]) {
		// only date, add time
		if !utils.IsDate(dateRange[1]) {
			dateRange[1] += " 23:59:59"
		}
		query.Where(col+" <= ?", dateRange[1])
	}
	return nil
}

// OperatorBetweenTime WHERE column >= datetime0 AND column <= datetime1
// will add 00:00:00 to the date 0 and 23:59:59 to the date 1 if only date is provided
func OperatorBetweenTime(query *QueryBuilder, col string, val any) error {
	timeRange, ok := val.([]string)
	if !ok || len(timeRange) < 2 {
		return fmt.Errorf("between time operator need two values")
	}
	if timeRange[0] != "" && utils.IsTime(timeRange[0]) {
		query.Where(col+" >= ?", timeRange[0])
	}
	if timeRange[1] != "" && utils.IsTime(timeRange[1]) {
		query.Where(col+" <= ?", timeRange[1])
	}
	return nil
}

// OperatorIsNull WHERE column IS NULL
func OperatorIsNull(query *QueryBuilder, col string, _ any) error {
	query.Where(col + " is null")
	return nil
}

// OperatorIsNotNull WHERE column IS NOT NULL
func OperatorIsNotNull(query *QueryBuilder, col string, _ any) error {
	query.Where(col + " is not null")
	return nil
}

// ParseQueryBuilder 解析protobuf请求参数，构建查询条件。注意：使用的proto文件的字段名
func ParseQueryBuilder(query *QueryBuilder, request utils.IProtobuf, columnOperators Operators) (*QueryBuilder, error) {
	// 通过反射获取request的字段和值，插入到requestKVs中
	requestKVs := utils.ProtobufToMap(request, false)

	for colName, operator := range columnOperators {
		val, ok := requestKVs[colName]
		if !ok || val == nil || val == "" { // 字段不存在、或者值为nil、或者值为""
			continue
		}

		if err := operator(query, colName, val); err != nil {
			return nil, err
		}
	}
	return query, nil
}

// ParsePagination 解析protobuf请求参数，构建分页参数。注意：使用的proto文件的字段名
func ParsePagination(request utils.IProtobuf, pageName, pageSizeName string) *db.Pagination {
	if request == nil {
		return nil
	}

	var page, pageSize int64
	vOf := reflect.ValueOf(request).Elem()
	tOf := vOf.Type()

	for i := 0; i < tOf.NumField(); i++ {
		field := tOf.Field(i)
		// 如果是私有字段、匿名字段，则跳过
		if !field.IsExported() || field.Anonymous {
			continue
		}

		name := field.Name
		tagName, ok := field.Tag.Lookup("json")
		if ok && tagName != "-" && tagName != "_" {
			segments := strings.Split(tagName, ",")
			name = segments[0]
		}

		if name == pageName || name == pageSizeName {
			valOf := vOf.Field(i)
			// 兼容 int 或 *int (optional int32)
			if valOf.Kind() == reflect.Ptr {
				if valOf.IsNil() { // 空指针，则跳过
					continue
				}
				// 获取指针指向的值
				valOf = valOf.Elem()
			}

			if name == pageName {
				page = utils.ToInt64(valOf.Interface())
			} else if name == pageSizeName {
				pageSize = utils.ToInt64(valOf.Interface())
			}
		}
	}

	return db.NewPagination(int(page), int(pageSize))
}
