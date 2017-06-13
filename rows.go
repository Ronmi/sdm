package sdm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/Ronmi/sdm/driver"
)

// Rows proxies all needed methods of sql.Rows
type Rows struct {
	rows    *sql.Rows
	def     map[string]driver.Column
	columns []string
	e       error
	t       reflect.Type
	drv     driver.Driver
}

func (r *Rows) err(msg string) error {
	r.e = errors.New(msg)
	return r.e
}

func (r *Rows) errf(msg string, args ...interface{}) error {
	r.e = fmt.Errorf(msg, args...)
	return r.e
}

// Scan reads columns into fields
func (r *Rows) Scan(data interface{}) (err error) {
	if err = r.e; err != nil {
		return
	}

	vstruct := reflect.ValueOf(data)
	if k := vstruct.Kind(); k != reflect.Ptr && k != reflect.Interface {
		return r.err("sdm: need reference to change data")
	}
	vstruct = vstruct.Elem()
	if t := vstruct.Type(); t != r.t {
		return r.errf("sdm: type mismatch, need %s but got %s", r.t.String(), t.String())
	}

	for _, col := range r.columns {
		if _, ok := r.def[col]; !ok {
			return r.errf("sdm: column %s not in struct", col)
		}
	}

	holders := make([]interface{}, len(r.columns))
	for idx, col := range r.columns {
		vf := vstruct.Field(r.def[col].ID)

		if val, ok := r.drv.GetScanner(vf); ok {
			holders[idx] = val
		} else {
			holders[idx] = vf.Addr().Interface()
		}
	}

	r.e = r.rows.Scan(holders...)
	return r.e
}

// Err proxies sql.Rows.Close
func (r *Rows) Err() error {
	if r.e == nil {
		r.e = r.rows.Err()
	}
	return r.e
}

// Next proxies sql.Rows.Next
func (r *Rows) Next() bool {
	if r.e != nil {
		return false
	}

	return r.rows.Next()
}

// Close proxies sql.Rows.Close
func (r *Rows) Close() error {
	if r.e != nil {
		return r.e
	}

	if r.rows == nil {
		return r.e
	}

	r.e = r.rows.Close()
	return r.e
}

// Columns proxies sql.Rows.Columns
func (r *Rows) Columns() ([]string, error) {
	return r.columns, r.e
}
