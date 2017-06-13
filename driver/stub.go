package driver

import (
	"database/sql"
	sqlDriver "database/sql/driver"
	"reflect"
)

type wrappable interface {
	sql.Scanner
	sqlDriver.Valuer
}

// Stub implements most of methods of a Driver.
//
// By providing a quote function, most features a driver should implement
// is prepared for you. Only CreateTables and CreateblesNotExist are
// required for you to implement.
type Stub struct {
	QuoteFunc func(name string) string
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
