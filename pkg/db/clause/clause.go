package clause

import (
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

type (
	Locking    = clause.Locking
	Expression = clause.Expression
	Table      = clause.Table
	Returning  = clause.Returning
	Column     = clause.Column
	Join       = clause.Join
)

const (
	CurrentTable = clause.CurrentTable
	Read         = dbresolver.Read
	Write        = dbresolver.Write
)
