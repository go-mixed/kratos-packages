package clause

import "gorm.io/gorm/clause"

type (
	Locking    = clause.Locking
	Expression = clause.Expression
	Table      = clause.Table
)

const (
	CurrentTable = clause.CurrentTable
)
