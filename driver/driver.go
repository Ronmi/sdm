package driver

import (
	"database/sql"
	"reflect"

	"git.ronmi.tw/ronmi/sdm"
)

// Driver is used to generate vendor-specific SQL syntax
type Driver struct {
	// Functions to create table
	CreateTable         func(db *sql.DB, name string, typ reflect.Type, cols []sdm.ColumnDef, indexes []sdm.IndexDef) (sql.Result, error)
	CreateTableNotExist func(db *sql.DB, name string, typ reflect.Type, cols []sdm.ColumnDef, indexes []sdm.IndexDef) (sql.Result, error)

	// Quote quotes identifiers like table name or column name
	Quote func(name string) string
}
