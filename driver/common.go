package driver

import (
	"reflect"
	"time"
)

var timeType reflect.Type = reflect.TypeOf(time.Time{})

// ElementType returns type of element for ptr, array or slice, or unchanged otherwise.
func ElementType(t reflect.Type) reflect.Type {
	switch t.Kind() {
	case reflect.Array, reflect.Slice, reflect.Ptr:
		return t.Elem()
	}

	return t
}

// IsTime determins if a type is time.Time or *time.Time
func IsTime(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = ElementType(t)
	}
	return t == timeType
}

// IsString determins if a type is string/[]byte/[]rune or their pointer type
func IsString(t reflect.Type) bool {
	k := t.Kind()
	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}

	if k == reflect.String {
		return true
	}

	if k != reflect.Slice {
		return false
	}

	// slice of something
	k = t.Elem().Kind()
	if k == reflect.Uint8 {
		// []uint8 == []byte
		return true
	}

	if k == reflect.Int32 {
		// []int32 == []rune
		return true
	}

	return false
}

// IsInteger determins is a type is integer type or their pointer type
//
// Integer types includes int8/int16/int32/int64/int and their unsigned version
func IsInteger(t reflect.Type) bool {
	k := t.Kind()
	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}

	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:

		return true
	}

	return false
}

// IsUinteger determins is a type is unsigned integer type or their pointer type
func IsUinteger(t reflect.Type) bool {
	k := t.Kind()
	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}

	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:

		return true
	}

	return false
}

// IsFloat determins is a type is float type or their pointer type
func IsFloat(t reflect.Type) bool {
	k := t.Kind()
	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}

	switch k {
	case reflect.Float32, reflect.Float64:
		return true
	}

	return false
}

// QuotingType specified which kind of statement will this column be used
type QuotingType int

// When generating column name, specify where to use
const (
	QSelect = QuotingType(iota) // column list in SELECT statement
	QWhere                      // in WHERE clause
	QInsert                     // column list of INSERT statement
	QUpdate                     // column list of UPDATE statement
)
