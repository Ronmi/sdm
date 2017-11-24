package sdmext

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"

	"github.com/Ronmi/sdm/driver"
)

// SJSONFactory creates wrapper to (un)marshal struct to/from JSON
//
// It supports integers/floats/string/boolean/time.Time and their pointers.
//
// To prevent timezone related problems, time.Time is exported/imported as
// unix timestamp in milliseconds (Time.UnixNano()/1000000).
//
// SJSONFactory is neither thread-safe nor reentrant, initialize it multiple
// times or after wrapper created might run into race condition.
type SJSONFactory struct {
	typ  reflect.Type
	keys map[int][]byte
	cols map[string]int
}

func (f *SJSONFactory) Init(typ reflect.Type, mapFieldToColumn map[int]string) {
	f.typ = typ
	f.keys = make(map[int][]byte)
	f.cols = make(map[string]int)

	for x, c := range mapFieldToColumn {
		data, _ := json.Marshal(c)
		f.keys[x] = data
		f.cols[string(data)] = x
	}
}

func (f *SJSONFactory) encode(buf *bytes.Buffer, v reflect.Value) {
	// test pointer and null
	if v.Type().Kind() == reflect.Ptr {
		if v.IsNil() {
			buf.WriteString("null")
			return
		}

		v = reflect.Indirect(v)
	}

	t := v.Type()
	k := t.Kind()

	switch {
	case k == reflect.Bool:
		data := []byte("false")
		if v.Bool() {
			data = []byte("true")
		}
		buf.Write(data)
		return
	case driver.IsString(t):
		var s string
		data, _ := json.Marshal(v.Convert(reflect.TypeOf(s)).String())
		buf.Write(data)
		return
	case driver.IsUinteger(t):
		buf.WriteString(strconv.FormatUint(v.Uint(), 10))
		return
	case driver.IsInteger(t):
		buf.WriteString(strconv.FormatInt(v.Int(), 10))
		return
	case driver.IsFloat(t):
		buf.WriteString(strconv.FormatFloat(v.Float(), 'f', -1, 64))
		return
	case driver.IsTime(t):
		buf.WriteString(strconv.FormatInt(
			v.Interface().(time.Time).UnixNano()/1000000, 10))
		return
	}
	panic(errors.New("SJSONFactory does not support [" + t.PkgPath() + "." + t.Name() + "]"))
}

func (f *SJSONFactory) m(data interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte('{')
	v := reflect.ValueOf(data)
	if v.Type().Kind() == reflect.Ptr {
		if v.IsNil() {
			return []byte("null"), nil
		}
		v = reflect.Indirect(v)
	}

	for x, c := range f.keys {
		buf.Write(c)
		buf.WriteByte(':')

		f.encode(buf, v.Field(x))

		buf.WriteByte(',')
	}

	ret := buf.Bytes()
	ret[len(ret)-1] = '}'
	return ret, nil
}

func (f *SJSONFactory) decode(data []byte, target reflect.Value) error {
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}

		target = target.Elem()
	}

	if driver.IsTime(target.Type()) {
		n := json.Number(string(data))
		i, err := n.Int64()
		if err != nil {
			f, err := n.Float64()
			if err != nil {
				return err
			}
			i = int64(f)
		}
		t := time.Unix(i/1000, i%1000)
		target.Set(reflect.ValueOf(t))

		return nil
	}

	return json.Unmarshal(data, target.Interface())
}

func (f *SJSONFactory) u(data interface{}, buf []byte) error {
	return f.decode(buf, reflect.ValueOf(data))
}

// JSONWrapper wraps registered struct to convert to/from json
type JSONWrapper interface {
	json.Marshaler
	json.Unmarshaler
}

// Wrap wraps registered struct.
//
// Pass incompatible type may cause unexpected issue.
//
// For unmarshaling, do not forget to pass a settable interface.
func (f *SJSONFactory) Wrap(data interface{}) JSONWrapper {
	return &jsonWrapper{
		m:    f.m,
		u:    f.u,
		data: data,
	}
}

type jsonWrapper struct {
	m    func(interface{}) ([]byte, error)
	u    func(interface{}, []byte) error
	data interface{}
}

func (w *jsonWrapper) MarshalJSON() ([]byte, error) {
	return w.m(w.data)
}

func (w *jsonWrapper) UnmarshalJSON(data []byte) error {
	return w.u(w.data, data)
}
