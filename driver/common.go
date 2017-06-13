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

// QuotingType specified which kind of statement will this column be used
type QuotingType int

// When generating column name, specify where to use
const (
	QSelect = QuotingType(iota) // column list in SELECT statement
	QWhere                      // in WHERE clause
	QInsert                     // column list of INSERT statement
	QUpdate                     // column list of UPDATE statement
)
