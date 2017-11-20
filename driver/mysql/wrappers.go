package mysql

import (
	sqlDriver "database/sql/driver"
	"errors"
	"reflect"
	"time"
)

type timeWrapper reflect.Value

func scanPointer(vs, v reflect.Value, src interface{}) error {
	if vs.Type().Kind() == reflect.Ptr && vs.IsNil() {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	return scanValue(vs, v.Elem(), src)
}

func scanValue(vs, v reflect.Value, src interface{}) error {
	// some mysql driver provides dsn param to parse time format, we can detect it here
	if t, ok := src.(time.Time); ok {
		v.Set(reflect.ValueOf(t))
		return nil
	}

	// normal string representation
	arr, ok := src.([]byte)
	if !ok {
		return errors.New("sdm: driver: mysql: invalid time source: " + vs.Type().Name())
	}
	str := string(arr)
	if t, err := time.Parse(`2006-01-02 15:04:05.000000`, str); err == nil {
		// DATETIME/TIMESTAMP with fraction
		v.Set(reflect.ValueOf(t))
		return nil
	}
	if t, err := time.Parse(`2006-01-02 15:04:05`, str); err == nil {
		// DATETIME/TIMESTAMP
		v.Set(reflect.ValueOf(t))
		return nil
	}
	if t, err := time.Parse(`2006-01-02`, str); err == nil {
		// DATE
		v.Set(reflect.ValueOf(t))
		return nil
	}
	if t, err := time.Parse(`2006`, str); err == nil {
		// YEAR
		v.Set(reflect.ValueOf(t))
		return nil
	}

	return errors.New("sdm: driver: mysql: incorrect time string: " + str)
}

func (w timeWrapper) Scan(src interface{}) error {
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

func (w timeWrapper) Value() (ret sqlDriver.Value, err error) {
	v := reflect.Value(w)
	if v.Type().Kind() == reflect.Ptr && (v.IsNil() || !v.Elem().IsValid()) {
		return nil, nil
	}

	return v.Interface(), nil
}
