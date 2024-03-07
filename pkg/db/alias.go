package db

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	Open              = gorm.Open
	ErrRecordNotFound = gorm.ErrRecordNotFound
	Expr              = gorm.Expr
)

type (
	DB     = gorm.DB
	Config = gorm.Config
	JSON   = datatypes.JSON
	Tabler = schema.Tabler
)

type ModelCollection []Tabler
type SeederCollection []ISeeder
