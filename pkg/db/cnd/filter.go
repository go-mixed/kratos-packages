package cnd

import (
	"fmt"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
)

type Operators map[string]Operator

type Operator string

const (
	OperatorEq         Operator = "="
	OperatorNe         Operator = "<>"
	OperatorGt         Operator = ">"
	OperatorGe         Operator = ">="
	OperatorLt         Operator = "<"
	OperatorLe         Operator = "<="
	OperatorLike       Operator = "like"
	OperatorIn         Operator = "in"
	OperatorNotIn      Operator = "not in"
	OperatorBetween    Operator = "between"
	OperatorNotBetween Operator = "not between"
	OperatorIsNull     Operator = "is null"
	OperatorIsNotNull  Operator = "is not null"
	OperatorPage       Operator = "page"
	OperatorPageSize   Operator = "page_size"
)

// ParseQueryBuilder 解析protobuf请求参数，构建查询条件
func ParseQueryBuilder(query *QueryBuilder, request utils.IProtobuf, columns Operators) (*QueryBuilder, *db.Pagination) {

	var page, pageSize int64
	// 通过反射获取request的字段和值，插入到requestKVs中
	requestKVs := utils.ProtobufToMap(request, false)

	for colName, operator := range columns {
		val, ok := requestKVs[colName]
		if !ok || val == nil || val == "" { // 字段不存在、或者值为nil、或者值为""
			continue
		}

		switch operator {
		case OperatorEq, OperatorNe, OperatorGt, OperatorGe, OperatorLt, OperatorLe, OperatorIn, OperatorNotIn:
			query.Where(colName+" "+string(operator)+" ?", val)
		case OperatorLike:
			query.Where(colName+" "+string(operator)+" ?", fmt.Sprintf("%%%s%%", val))
		case OperatorBetween, OperatorNotBetween:

		case OperatorIsNull, OperatorIsNotNull:
			query.Where(colName + " " + string(operator))
		case OperatorPage:
			page = utils.ToInt64(val)
		case OperatorPageSize:
			pageSize = utils.ToInt64(val)
		}
	}
	return query, db.NewPagination(int(page), int(pageSize))
}
