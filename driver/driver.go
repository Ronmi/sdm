package driver

import (
	"database/sql"
	"errors"
	"reflect"
)

// Driver is used to generate vendor-specific SQL syntax
type Driver interface {
	// Functions to create table
	CreateTable(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error)
	CreateTableNotExist(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error)

	// Quote quotes identifiers like table name or column name
	Quote(name string) string
}

type defaultDriver struct {
}

func (d defaultDriver) CreateTable(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error) {
	return nil, errors.New("sdm: driver: CreateTable is not supported in default driver")
}

func (d defaultDriver) CreateTableNotExist(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error) {
	return nil, errors.New("sdm: driver: CreateTable is not supported in default driver")
}

func (d defaultDriver) Quote(name string) string {
	return name
}

// ValidateDriver checks if any method not implemented, and fill with default implementation
func Default() Driver {
	return defaultDriver{}
}
