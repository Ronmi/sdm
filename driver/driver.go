package driver

import (
	"database/sql"
	"reflect"
)

// Driver is used to generate vendor-specific SQL syntax
type Driver interface {
	// Functions to create table
	CreateTable(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error)
	CreateTableNotExist(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error)

	// Quote quotes identifiers like table name or column name
	Quote(name string) string

	// Formats for general SQL statement
	ColIns(table, col string) string // insert
	Col(table, col string) string    // others
}

