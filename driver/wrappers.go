package driver

import (
	sqlDriver "database/sql/driver"
	"reflect"
)

type wrapper reflect.Value

func (w wrapper) Scan(src interface{}) error {
	vs := reflect.ValueOf(src)
	v := reflect.Value(w)
	if src == nil {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	if !v.Elem().IsValid() {
		// allocate if not initial alocated
		v.Set(reflect.New(v.Type().Elem()))
	}
	v.Elem().Set(vs)
	return nil
}

func (w wrapper) Value() (ret sqlDriver.Value, err error) {
	v := reflect.Value(w)
	if v.IsNil() || !v.Elem().IsValid() {
		return nil, nil
	}

	return v.Interface(), nil
}
