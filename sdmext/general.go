package sdmext

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Ronmi/sdm"
)

// ExtGeneral is a general purpose extension primarily for http forms.
//
// It supports integers, float points, string and boolean types. time.Time is also
// supported via customizable parser.
//
// For boolean fields, empty string/false/f/0 are treated as false, others are true.
//
// Only fields matched are filled (excepts boolean fields), others are leave
// unchanged.
type ExtGeneral struct {
	ParseTime func(string) (time.Time, error)
	typ       reflect.Type
	columns   map[int]string
}

func (g *ExtGeneral) Init(t reflect.Type, c map[int]string) {
	g.typ = t
	g.columns = c
}

func (g *ExtGeneral) ReadTo(data interface{}, retriver func(string) string) *sdm.ErrExtension {
	if g.typ == nil || g.columns == nil {
		return &sdm.ErrExtension{
			ExtName: "ExtGeneral",
			Reason:  errors.New("extension must be initialized before using"),
		}
	}

	vstruct := reflect.ValueOf(data)
	if vstruct.Type().Kind() != reflect.Ptr {
		return &sdm.ErrExtension{
			ExtName:    "ExtGeneral",
			StructName: g.typ.String(),
			Reason:     errors.New("need pointer type to set value"),
		}
	}
	vstruct = vstruct.Elem()
	if t := vstruct.Type(); t != g.typ {
		return &sdm.ErrExtension{
			ExtName:    "ExtGeneral",
			StructName: g.typ.String(),
			Reason:     fmt.Errorf("type mismatch, need %s but got %s", g.typ.String(), t.String()),
		}
	}

	for x := 0; x < vstruct.NumField(); x++ {
		f := vstruct.Field(x)
		k := f.Kind()
		c := g.columns[x]

		data := retriver(c)
		if data == "" && k != reflect.Bool {
			continue
		}

		switch k {
		case reflect.Bool:
			data = strings.ToLower(data)
			switch data {
			case "", "false", "f", "0", "0.0":
				f.SetBool(false)
			default:
				f.SetBool(true)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(data, 10, 64)
			if err != nil {
				return &sdm.ErrExtension{
					ExtName:    "ExtGeneral",
					StructName: g.typ.String(),
					FieldName:  f.Type().String(),
					Reason:     err,
				}
			}
			f.SetInt(i)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			i, err := strconv.ParseUint(data, 10, 64)
			if err != nil {
				return &sdm.ErrExtension{
					ExtName:    "ExtGeneral",
					StructName: g.typ.String(),
					FieldName:  f.Type().String(),
					Reason:     err,
				}
			}
			f.SetUint(i)
		case reflect.Float32, reflect.Float64:
			i, err := strconv.ParseFloat(data, 64)
			if err != nil {
				return &sdm.ErrExtension{
					ExtName:    "ExtGeneral",
					StructName: g.typ.String(),
					FieldName:  f.Type().String(),
					Reason:     err,
				}
			}
			f.SetFloat(i)
		case reflect.String:
			f.SetString(data)
		case reflect.Slice:
			switch f.Elem().Kind() {
			case reflect.Uint8, reflect.Int32:
				// []byte or []rune
				f.SetString(data)
			}
		case reflect.Struct:
			if g.ParseTime == nil {
				continue
			}
			if f.Type().String() != "time.Time" {
				continue
			}

			t, err := g.ParseTime(data)
			if err != nil {
				return &sdm.ErrExtension{
					ExtName:    "ExtGeneral",
					StructName: g.typ.String(),
					FieldName:  f.Type().String(),
					Reason:     err,
				}
			}
			f.Set(reflect.ValueOf(t))
		}
	}

	return nil
}

// FromPost creates function to be used in ReadTo, retrives data from post form
func FromPost(r *http.Request) func(string) string {
	r.ParseForm()
	return func(key string) string {
		return r.PostForm.Get(key)
	}
}

// FromGet creates function to be used in ReadTo, retrives data from url queries
func FromGet(r *http.Request) func(string) string {
	q := r.URL.Query()
	return func(key string) string {
		return q.Get(key)
	}
}

// FromForms creates function to be used in ReadTo, retrives data from form values
func FromForms(r *http.Request) func(string) string {
	r.ParseForm()
	return func(key string) string {
		return r.FormValue(key)
	}
}
