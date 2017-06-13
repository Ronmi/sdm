package sqlite3

import (
	"database/sql"
	sqlDriver "database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"time"
)

type wrapper interface {
	sql.Scanner
	sqlDriver.Valuer
}

type wrapTimeString struct {
	v        reflect.Value
	nullable bool
}

func (w wrapTimeString) Scan(src interface{}) error {
	if reflect.TypeOf(src).Kind() != reflect.String {
		return errors.New("sdm: driver: sqlite3: this is not a string")
	}

	str := src.(string)
	if str == "" {
		if w.nullable {
			w.v.Set(reflect.Zero(w.v.Type()))
			return nil
		}

		w.v.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	t, err := time.Parse(TimeStringFormat, src.(string))
	if err != nil {
		return nil
	}

	val := w.v
	if w.nullable {
		if !w.v.Elem().IsValid() {
			w.v.Set(reflect.New(w.v.Type().Elem()))
		}
		val = w.v.Elem()
	}
	val.Set(reflect.ValueOf(t))
	return nil
}

func (w wrapTimeString) Value() (ret sqlDriver.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sdm: driver: sqlite: error converting value: %v", r)
		}
	}()

	var t time.Time
	if w.nullable {
		if w.v.IsNil() || !w.v.Elem().IsValid() {
			return
		}

		t = reflect.Indirect(w.v).Interface().(time.Time)
	} else {
		t = w.v.Interface().(time.Time)
	}

	return t.Format(TimeStringFormat), nil
}

type wrapTimeInt struct {
	v        reflect.Value
	nullable bool
}

func (w wrapTimeInt) Scan(src interface{}) error {
	v := reflect.ValueOf(src)

	if src == nil {
		if w.nullable {
			w.v.Set(reflect.Zero(w.v.Type()))
			return nil
		}

		w.v.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	t := time.Unix(v.Int(), 0)

	val := w.v
	if w.nullable {
		if !w.v.Elem().IsValid() {
			w.v.Set(reflect.New(w.v.Type().Elem()))
		}
		val = w.v.Elem()
	}
	val.Set(reflect.ValueOf(t))
	return nil
}

func (w wrapTimeInt) Value() (ret sqlDriver.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sdm: driver: sqlite: error converting value: %v", r)
		}
	}()

	var t time.Time
	if w.nullable {
		if w.v.IsNil() || !w.v.Elem().IsValid() {
			return
		}

		t = reflect.Indirect(w.v).Interface().(time.Time)
	} else {
		t = w.v.Interface().(time.Time)
	}

	return t.Unix(), nil
}
