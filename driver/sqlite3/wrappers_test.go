package sqlite3

import (
	"reflect"
	"testing"
	"time"
)

type wrapperTestType struct {
	Time       time.Time
	ValuedTime *time.Time
	NilledTime *time.Time
}

type wrapperTestCase struct {
	title string
	field int
	value interface{}
	msg   string
}

func getWrapperData(i int) (*wrapperTestType, reflect.Value) {
	t := &wrapperTestType{}
	return t, reflect.Indirect(reflect.ValueOf(t)).Field(i)
}

func TestScanIntTime(t *testing.T) {
	cases := []wrapperTestCase{
		{
			title: "NonNullIntTime",
			field: 0,
			value: 1497302303,
			msg:   "cannot scan integer into time.Time",
		},
		{
			title: "ValuedIntTime",
			field: 1,
			value: 1497302303,
			msg:   "cannot scan integer into *time.Time",
		},
		{
			title: "NilledIntTime",
			field: 1,
			value: nil,
			msg:   "cannot scan nil into *time.Time",
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			_, v := getWrapperData(c.field)
			w := wrapTimeInt{
				v:        v,
				nullable: c.field == 1,
			}
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

			v = reflect.Indirect(v)
			if !reflect.DeepEqual(v.Interface(), time.Unix(reflect.ValueOf(c.value).Int(), 0)) {
				t.Fatalf("expect to get %v, got %v", c.value, v.Interface())
			}
		})
	}
}

func TestScanStringTime(t *testing.T) {
	cases := []wrapperTestCase{
		{
			title: "NonNullStringTime",
			field: 0,
			value: "2017-06-13T05:18:23+0800",
			msg:   "cannot scan string into time.Time",
		},
		{
			title: "ValuedStringTime",
			field: 1,
			value: "2017-06-13T05:18:23+0800",
			msg:   "cannot scan string into *time.Time",
		},
		{
			title: "NilledStringTime",
			field: 1,
			value: "",
			msg:   "cannot scan nil (empty string) into *time.Time",
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			_, v := getWrapperData(c.field)
			w := wrapTimeString{
				v:        v,
				nullable: c.field == 1,
			}
			if err := w.Scan(c.value); err != nil {
				t.Fatal(c.msg+":", err)
			}

			// examine value
			if c.value == "" {
				if !v.IsNil() {
					t.Fatalf("expect to fill nil, got %v", v.Interface())
				}
				return
			}

			ti, _ := time.Parse(TimeStringFormat, c.value.(string))
			v = reflect.Indirect(v)
			if !reflect.DeepEqual(v.Interface(), ti) {
				t.Fatalf("expect to get %v, got %v", c.value, v.Interface())
			}
		})
	}
}

func TestValueInt(t *testing.T) {
	m := time.Now()
	target := wrapperTestType{
		Time:       m,
		ValuedTime: &m,
	}

	cases := []wrapperTestCase{
		{
			title: "NormalInt",
			field: 0,
			value: m,
		},
		{
			title: "ValuedInt",
			field: 1,
			value: m,
		},
		{
			title: "NilledInt",
			field: 2,
			value: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			w := wrapTimeInt{
				v:        reflect.ValueOf(target).Field(c.field),
				nullable: c.field != 0,
			}

			v, err := w.Value()
			if err != nil {
				t.Fatalf("cannot get value: %s", err)
			}

			if c.value == nil {
				if v != nil {
					t.Fatalf("expected to be nil, got %v", v)
				}
				return
			}

			if ts := m.Unix(); v != ts {
				t.Fatalf("expect %d, get %v", ts, v)
			}
		})
	}
}

func TestValueString(t *testing.T) {
	m := time.Now()
	target := wrapperTestType{
		Time:       m,
		ValuedTime: &m,
	}

	cases := []wrapperTestCase{
		{
			title: "NormalString",
			field: 0,
			value: m,
		},
		{
			title: "ValuedString",
			field: 1,
			value: m,
		},
		{
			title: "NilledString",
			field: 2,
			value: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			w := wrapTimeString{
				v:        reflect.ValueOf(target).Field(c.field),
				nullable: c.field != 0,
			}

			v, err := w.Value()
			if err != nil {
				t.Fatalf("cannot get value: %s", err)
			}

			if c.value == nil {
				if v != nil {
					t.Fatalf("expected to be nil, got %v", v)
				}
				return
			}

			if ts := m.Format(TimeStringFormat); v != ts {
				t.Fatalf("expect %s, get %v", ts, v)
			}
		})
	}
}
