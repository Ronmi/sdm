package driver

import (
	"database/sql"
	"errors"
	"reflect"
)

// Driver is used to generate vendor-specific SQL syntax
type Driver struct {
	// Functions to create table
	CreateTable         func(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error)
	CreateTableNotExist func(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error)

	// Quote quotes identifiers like table name or column name
	Quote func(name string) string
}

func defaultDriverCreateTable(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error) {
	return nil, errors.New("sdm: driver: CreateTable is not supported in default driver")
}
func defaultDriverQuote(name string) string {
	return name
}

// ValidateDriver checks if any method not implemented, and fill with default implementation
func ValidateDriver(d *Driver) {
	if d.CreateTable == nil {
		d.CreateTable = defaultDriverCreateTable
	}
	if d.CreateTableNotExist == nil {
		d.CreateTableNotExist = defaultDriverCreateTable
	}
	if d.Quote == nil {
		d.Quote = defaultDriverQuote
	}
}
