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
// unix timestamp in seconds (Time.Unix()).
//
// SJSONFactory is neither thread-safe nor reentrant, initialize it multiple
// times or after wrapper created might run into race condition.
type SJSONFactory struct {
	typ  reflect.Type
	f2c  map[int]string
	keys map[int][]byte
}

func (f *SJSONFactory) Init(typ reflect.Type, mapFieldToColumn map[int]string) {
	f.typ = typ
	f.f2c = mapFieldToColumn
	f.keys = make(map[int][]byte)

	for x, c := range f.f2c {
		f.keys[x], _ = json.Marshal(c)
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
		buf.WriteString(strconv.FormatInt(v.Interface().(time.Time).Unix(), 10))
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
