package db

import "database/sql"

type ParamPair struct {
	Query any   // 查询
	Args  []any // 参数
}

type OrderByCol struct {
	Column string
	Asc    bool
}

type CursorResult struct {
	Results any    `json:"results"`
	Cursor  string `json:"cursor"`
}

func SqlNullString(value string) sql.NullString {
	return sql.NullString{
		String: value,
		Valid:  len(value) > 0,
	}
}
