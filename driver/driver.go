package driver

import (
	"database/sql"
	sqlDriver "database/sql/driver"
	"errors"
	"reflect"
	"strings"
)

// Driver is used to generate vendor-specific SQL syntax
type Driver interface {
	// Functions to create table
	//
	// Pointer type should be created as nullable column if supported
	CreateTable(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error)
	CreateTableNotExist(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error)

	// Quote quotes identifiers like table name or column name
	Quote(name string) string

	// Formats for general SQL statement
	Col(table, col string, kind QuotingType) string // generate column name

	// Get unnamed placeholder for a type, used in INSERT, UPDATE and WHERE clause, mostly "?"
	GetPlaceholder(typ reflect.Type) string

	// Handling custom type conversion here.
	//
	// For internal supported types and types support Scanner/Valuer, implementation
	// SHOULD set ok = false to skip custom type conversion.
	GetScanner(field reflect.Value) (ret sql.Scanner, ok bool)
	GetValuer(field reflect.Value) (ret sqlDriver.Valuer, ok bool)
}

// DriverFactory represents a function to create driver.
type DriverFactory func(params map[string]string) Driver

var registeredDrivers = map[string]DriverFactory{}

// RegisterDriver registers a driver factory for use in sdm.
//
// Note: later one with the same name will be discarded.
func RegisterDriver(name string, f DriverFactory) {
	if _, ok := registeredDrivers[name]; ok {
		return
	}

	registeredDrivers[name] = f
}

// GetDriver creates a driver from config string
func GetDriver(drvStr string) Driver {
	idx := strings.Index(drvStr, ":")
	name := drvStr
	if idx != -1 {
		name = drvStr[:idx]
	}

	factory, ok := registeredDrivers[name]
	if !ok {
		panic(errors.New("sdm: driver " + name + " not found"))
	}

	params := map[string]string{}

	if idx != -1 {
		// parse params
		strs := strings.Split(drvStr[idx+1:], ";")
		for _, paramStr := range strs {
			idx := strings.Index(paramStr, "=")
			val := ""
			if idx > 0 {
				val = paramStr[idx+1:]
			}
			params[paramStr[:idx]] = val
		}
	}

	return factory(params)
}
