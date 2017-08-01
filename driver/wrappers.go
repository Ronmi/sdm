package driver

import (
	sqlDriver "database/sql/driver"
	"reflect"
)

type wrapper reflect.Value

func scanPointer(vs, v reflect.Value, src interface{}) error {
	if vs.Type().Kind() == reflect.Ptr && vs.IsNil() {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	return scanValue(vs, v.Elem(), src)
}

func scanValue(vs, v reflect.Value, src interface{}) error {
	v.Set(vs)
	return nil
}

func (w wrapper) Scan(src interface{}) error {
	vs := reflect.ValueOf(src)
	v := reflect.Value(w)

	// allocate space if needed
	if !v.IsValid() {
		v.Set(reflect.New(v.Type()))
	}

	if v.Type().Kind() == reflect.Ptr {
		return scanPointer(vs, v, src)
	}

	return scanValue(vs, v, src)
}

func (w wrapper) Value() (ret sqlDriver.Value, err error) {
	v := reflect.Value(w)
	if v.Type().Kind() == reflect.Ptr && (v.IsNil() || !v.Elem().IsValid()) {
		return nil, nil
	}

	return v.Interface(), nil
}
