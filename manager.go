package sdm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type fielddef struct {
	id   int
	isAI bool // auto increment
}

// Rows proxies all needed methods of sql.Rows
type Rows struct {
	*sql.Rows
	fields  map[string]*fielddef
	columns []string
	e       error
	t       reflect.Type
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
		if _, ok := r.fields[col]; !ok {
			return r.errf("sdm: column %s not in struct", col)
		}
	}

	holders := make([]interface{}, len(r.columns))
	for idx, col := range r.columns {
		vf := vstruct.Field(r.fields[col].id)
		vfa := vf.Addr()
		holders[idx] = vfa.Interface()
	}

	r.e = r.Rows.Scan(holders...)
	return r.e
}

// Err proxies sql.Rows.Close
func (r *Rows) Err() error {
	if r.e == nil {
		r.e = r.Rows.Err()
	}
	return r.e
}

// Columns proxies sql.Rows.Columns
func (r *Rows) Columns() ([]string, error) {
	return r.columns, r.e
}

// Manager is just manager. any question?
type Manager struct {
	mappings map[reflect.Type]map[string]*fielddef
	lock     sync.Mutex
}

// New create sdm manager
func New() *Manager {
	return &Manager{
		map[reflect.Type]map[string]*fielddef{},
		sync.Mutex{},
	}
}

func (m *Manager) getMap(t reflect.Type) (ret map[string]*fielddef, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	ret, ok := m.mappings[t]
	if ok {
		return
	}

	ret = make(map[string]*fielddef)

	if t.Kind() != reflect.Struct {
		return ret, fmt.Errorf("sdm: %s is not a struct type", t.String())
	}

	for idx := 0; idx < t.NumField(); idx++ {
		f := t.Field(idx)
		tag := f.Tag.Get("sdm")
		if tag == "" {
			// not decorated, skip
			continue
		}

		if f.Name[0] < 'A' || f.Name[0] > 'Z' {
			// not exported, skip
			continue
		}

		tags := strings.Split(tag, ",")
		col := tags[0]
		tags = tags[1:]
		ret[col] = &fielddef{}
		ret[col].id = idx
		for _, tag := range tags {
			switch tag {
			case "ai":
				ret[col].isAI = true
			}
		}
	}

	return
}

// Proxify proxies needed methods of sql.Rows
func (m *Manager) Proxify(r *sql.Rows, data interface{}) *Rows {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	f, e := m.getMap(t)
	c, err := r.Columns()
	if e == nil {
		e = err
	}

	return &Rows{
		r,
		f,
		c,
		e,
		t,
	}
}

// Col returns a list of columns in sql format
func (m *Manager) Col(data interface{}, table string) (ret []string, err error) {
	f, err := m.getMap(reflect.Indirect(reflect.ValueOf(data)).Type())
	if err != nil {
		return
	}
	ret = make([]string, 0, len(f))

	for col := range f {
		c := col
		if table != "" {
			c = table + "." + c
		}
		ret = append(ret, c)
	}

	return
}

// Insert inserts data into table.
// It will skip columns with "ai" tag
func (m *Manager) Insert(db *sql.DB, table string, data interface{}) (res sql.Result, err error) {
	val := reflect.Indirect(reflect.ValueOf(data))
	def, err := m.getMap(val.Type())
	if err != nil {
		return nil, err
	}

	cols := make([]string, 0, len(def))
	vals := make([]interface{}, 0, len(def))
	for col, fdef := range def {
		if fdef.isAI {
			// skip auto increment columns
			continue
		}

		cols = append(cols, col)
		vals = append(vals, val.Field(fdef.id).Interface())
	}
	holders := "?" + strings.Repeat(",?", len(cols)-1)
	qstr := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		table,
		strings.Join(cols, ","),
		holders,
	)

	return db.Exec(qstr, vals...)
}

// Update updates data in db.
func (m *Manager) Update(db *sql.DB, table string, data interface{}, where string, whereargs ...interface{}) (sql.Result, error) {
	val := reflect.Indirect(reflect.ValueOf(data))
	def, err := m.getMap(val.Type())
	if err != nil {
		return nil, err
	}

	cols := make([]string, 0, len(def))
	vals := make([]interface{}, 0, len(def)+len(whereargs))
	for col, fdef := range def {
		cols = append(cols, col+"=?")
		vals = append(vals, val.Field(fdef.id).Interface())
	}
	qstr := fmt.Sprintf(
		`UPDATE %s SET %s WHERE %s`,
		table,
		strings.Join(cols, ","),
		where,
	)
	if len(whereargs) > 0 {
		vals = append(vals, whereargs...)
	}

	return db.Exec(qstr, vals...)
}

// Delete deletes data in db.
func (m *Manager) Delete(db *sql.DB, table string, data interface{}) (sql.Result, error) {
	val := reflect.Indirect(reflect.ValueOf(data))
	def, err := m.getMap(val.Type())
	if err != nil {
		return nil, err
	}

	cols := make([]string, 0, len(def))
	vals := make([]interface{}, 0, len(def))
	for col, fdef := range def {
		cols = append(cols, col+"=?")
		vals = append(vals, val.Field(fdef.id).Interface())
	}
	qstr := fmt.Sprintf(
		`DELETE FROM %s WHERE %s`,
		table,
		strings.Join(cols, " AND "),
	)

	return db.Exec(qstr, vals...)
}
