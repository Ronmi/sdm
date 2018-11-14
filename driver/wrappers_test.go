package driver

import (
	"reflect"
	"testing"
	"time"
)

type wrapperTestType struct {
	NullableBool  *bool
	NullableInt   *int
	NullableFloat *float64
	NullableTime  *time.Time
}

type simpleTestCase struct {
	title string
	field int
	value interface{}
	msg   string
}

func getWrapperData(i int) (*wrapperTestType, reflect.Value) {
	t := &wrapperTestType{}
	return t, reflect.Indirect(reflect.ValueOf(t)).Field(i)
}

func TestScanSimple(t *testing.T) {
	cases := []simpleTestCase{
		{
			title: "ValuedBool",
			field: 0,
			value: true,
			msg:   "cannot scan value into bool",
		},
		{
			title: "NilledBool",
			field: 0,
			value: nil,
			msg:   "cannot scan nil into bool",
		},
		{
			title: "ValuedInt",
			field: 1,
			value: 1,
			msg:   "cannot scan value into int",
		},
		{
			title: "NilledInt",
			field: 1,
			value: nil,
			msg:   "cannot scan nil into int",
		},
		{
			title: "ValuedFloat",
			field: 2,
			value: 1.0,
			msg:   "cannot scan value into float",
		},
		{
			title: "NilledFloat",
			field: 2,
			value: nil,
			msg:   "cannot scan nil into float",
		},
		{
			title: "ValuedTime",
			field: 3,
			value: time.Now(),
			msg:   "cannot scan value into time",
		},
		{
			title: "NilledTime",
			field: 3,
			value: nil,
			msg:   "cannot scan nil into time",
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			_, v := getWrapperData(c.field)
			w := wrapper(v)
			if err := w.Scan(c.value); err != nil {
				t.Fatal(c.msg+":", err)
			}

			// examine value
			if c.value == nil {
				if !v.IsNil() {
					t.Fatalf("expect to fill nil, got %v", v.Interface())
				}
				return
			}

			v = v.Elem()
			if !reflect.DeepEqual(v.Interface(), c.value) {
				t.Fatalf("expect to get %v, got %v", c.value, v.Interface())
			}
		})
	}
}

func TestValueSimple(t *testing.T) {
	var (
		b bool      = true
		i int       = 1
		f float64   = 1.1
		m time.Time = time.Now()
	)
	target := wrapperTestType{&b, &i, &f, &m}
	cases := []simpleTestCase{
		{
			title: "Bool",
			field: 0,
			value: b,
		},
		{
			title: "Int",
			field: 1,
			value: int64(i),
		},
		{
			title: "Float",
			field: 2,
			value: f,
		},
		{
			title: "Time",
			field: 3,
			value: m,
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			w := wrapper(reflect.ValueOf(target).Field(c.field))

			v, err := w.Value()
			if err != nil {
				t.Fatalf("cannot get value: %s", err)
			}

			if v != c.value {
				t.Fatalf(
					"expected %v (%t), got %#v (%t)",
					c.value,
					c.value,
					v,
					v,
				)
			}
		})
	}
}

func TestValueNil(t *testing.T) {
	target := wrapperTestType{}
	cases := []simpleTestCase{
		{
			title: "Bool",
			field: 0,
		},
		{
			title: "Int",
			field: 1,
		},
		{
			title: "Float",
			field: 2,
		},
		{
			title: "Time",
			field: 3,
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			w := wrapper(reflect.ValueOf(target).Field(c.field))

			v, err := w.Value()
			if err != nil {
				t.Fatalf("cannot get value: %s", err)
			}

			if v != nil {
				t.Fatalf("expected to be nil, got %v", v)
			}
		})
	}
}
