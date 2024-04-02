package driver

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"strings"
)

func Open(driver string, dsn string) gorm.Dialector {
	switch strings.ToLower(driver) {
	case "mysql":
		return mysql.Open(dsn)
	case "pg", "postgresql", "pgsql", "postgres":
		return postgres.Open(dsn)
	case "sqlite":
		return sqlite.Open(dsn)
	case "sqlserver", "mssql":
		return sqlserver.Open(dsn)
	default:
		panic("invalid database driver： " + driver)
	}
}
