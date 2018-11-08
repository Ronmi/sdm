package driver

import (
	sqlDriver "database/sql/driver"
	"reflect"
)

type wrapper reflect.Value

func scanPointer(vs, v reflect.Value) error {
	if vs.Type().Kind() == reflect.Ptr && vs.IsNil() {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	return scanValue(vs, v.Elem())
}

func scanValue(vs, v reflect.Value) error {
	if vs.Type().Kind() == reflect.Ptr && vs.IsNil() {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	v.Set(vs.Convert(v.Type()))
	return nil
}

func (w wrapper) Scan(src interface{}) error {
	v := reflect.Value(w)
	if src == nil {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	vs := reflect.ValueOf(src)

	// allocate space if needed
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		if !v.Elem().IsValid() {
			v.Set(reflect.New(t.Elem()))
		}
		return scanPointer(vs, v)
	}

	return scanValue(vs, v)
}

func (w wrapper) Value() (ret sqlDriver.Value, err error) {
	v := reflect.Value(w)
	if v.Type().Kind() == reflect.Ptr && (v.IsNil() || !v.Elem().IsValid()) {
		return nil, nil
	}

	return v.Interface(), nil
}
