package driver

import (
	"database/sql"
	sqlDriver "database/sql/driver"
	"reflect"
	"strings"
)

type wrappable interface {
	sql.Scanner
	sqlDriver.Valuer
}

// Stub implements most of methods of a Driver.
//
// By providing a quote function, most features a driver should implement
// are prepared for you. Only CreateTables and CreateblesNotExist are
// not supported, and they always panic.
//
// Using Stub, you should map nullable columns to fields declared as pointer type.
// The only exception is string type, in which NULL is always mapped to "".
type Stub struct {
	QuoteFunc func(name string) string
}

// CreateTable is not implemented, panic always.
func (s Stub) CreateTable(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error) {
	panic("sdm: driver: Default stub driver does not support table creation!")
}

// CreateTableNotExist is not implemented, panic always.
func (s Stub) CreateTableNotExist(db *sql.DB, name string, typ reflect.Type, cols []Column, indexes []Index) (sql.Result, error) {
	panic("sdm: driver: Default stub driver does not support table creation!")
}

func (s Stub) ParseColumnName(c string) string {
	arr := strings.Split(c, ".")
	return arr[len(arr)-1]
}

func (s Stub) Quote(name string) string {
	return s.QuoteFunc(name)
}

func (s Stub) Col(table, col string, kind QuotingType) string {
	return s.QuoteFunc(col)
}

func (s Stub) GetPlaceholder(typ reflect.Type) string {
	return "?"
}

func (s Stub) GetScanner(field reflect.Value) (ret sql.Scanner, ok bool) {
	return s.getWrapper(field)
}

func (s Stub) GetValuer(field reflect.Value) (ret sqlDriver.Valuer, ok bool) {
	return s.getWrapper(field)
}

func (s Stub) getWrapper(v reflect.Value) (ret wrappable, ok bool) {
	typ := v.Type()
	k := typ.Kind()

	// skip types not Ptr
	if k != reflect.Ptr {
		return
	}

	// handle *time.Time
	if IsTime(typ) {
		return wrapper(v), true
	}

	switch reflect.Indirect(v).Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Bool:
		fallthrough
	case reflect.Float32, reflect.Float64:
		return wrapper(v), true
	}

	return
}

// RegisterStub helps you to use stub as driver.
//
// After registering the stub driver, you can create SDM manager like this
//
//     m := sdm.New(conn, "stub")
func RegisterStub(quoteFunc func(columnName string) (quotedName string)) {
	RegisterStubAs("stub", quoteFunc)
}

// RegisterStub is just like RegisterStub, with custom driver name
func RegisterStubAs(registerAs string, quoteFunc func(columnName string) (quotedName string)) {
	RegisterDriver(registerAs, func(p map[string]string) Driver {
		return Stub{
			QuoteFunc: quoteFunc,
		}
	})
}

// AnsiStub registers a single-quote (ANSI SQL compitable) based SDM stub driver as "ansistub"
func AnsiStub() {
	RegisterStubAs("ansistub", func(n string) string {
		return `'` + n + `'`
	})
}

// MySQLStub registers a back-quote (MySQL compitable) based SDM stub driver as "mysqlstub"
func MySQLStub() {
	RegisterStubAs("mysqlstub", func(n string) string {
		return "`" + n + "`"
	})
}
