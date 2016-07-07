package sdm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type fielddef struct {
	id   int    // field id
	isAI bool   // auto increment
	name string // column name
}

// Manager is just manager. any question?
type Manager struct {
	columns map[reflect.Type]map[string]*fielddef
	fields  map[reflect.Type][]*fielddef
	lock    sync.RWMutex
	db      *sql.DB
}

// New create sdm manager
func New(db *sql.DB) *Manager {
	return &Manager{
		map[reflect.Type]map[string]*fielddef{},
		map[reflect.Type][]*fielddef{},
		sync.RWMutex{},
		db,
	}
}

func (m *Manager) register(t reflect.Type) (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.fields[t]; ok {
		return
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("sdm: %s is not a struct type", t.String())
	}

	mps := make([]*fielddef, 0, t.NumField())
	idx := make(map[string]*fielddef)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
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

		fdef := &fielddef{id: i, name: col}
		for _, tag := range tags {
			switch tag {
			case "ai":
				fdef.isAI = true
			}
		}

		mps = append(mps, fdef)
		idx[col] = fdef
	}

	m.columns[t] = idx
	m.fields[t] = mps
	return
}

func (m *Manager) getDef(t reflect.Type) (ret []*fielddef, err error) {
	if err = m.register(t); err != nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	ret = m.fields[t]
	return
}

func (m *Manager) getMap(t reflect.Type) (ret map[string]*fielddef, err error) {
	if err = m.register(t); err != nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	ret = m.columns[t]
	return
}

// Col returns a list of columns in sql format, including AUTO INCREMENT columns
func (m *Manager) Col(data interface{}, table string) (ret []string, err error) {
	fdef, err := m.getDef(reflect.Indirect(reflect.ValueOf(data)).Type())
	if err != nil {
		return
	}
	ret = make([]string, 0, len(fdef))

	for _, f := range fdef {
		c := f.name
		if table != "" {
			c = table + "." + c
		}
		ret = append(ret, c)
	}

	return
}

// ColIns returns a list of columns in sql format, excluding AUTO INCREMENT columns
func (m *Manager) ColIns(data interface{}, table string) (ret []string, err error) {
	fdef, err := m.getDef(reflect.Indirect(reflect.ValueOf(data)).Type())
	if err != nil {
		return
	}
	ret = make([]string, 0, len(fdef))

	for _, f := range fdef {
		if f.isAI {
			continue
		}
		c := f.name
		if table != "" {
			c = table + "." + c
		}
		ret = append(ret, c)
	}

	return
}

// Val converts struct to value array
func (m *Manager) Val(data interface{}) ([]interface{}, error) {
	var ret []interface{}
	v := reflect.ValueOf(data)
	t := reflect.Indirect(v).Type()
	fdef, err := m.getDef(t)
	if err != nil {
		return nil, err
	}

	ret = make([]interface{}, len(fdef))
	for k, f := range fdef {
		ret[k] = v.Field(f.id).Interface()
	}
	return ret, nil
}

// Connection returns stored *sql.DB
func (m *Manager) Connection() *sql.DB {
	return m.db
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

// Query makes SQL query and proxies it
func (m *Manager) Query(typ interface{}, qstr string, args ...interface{}) *Rows {
	dbrows, err := m.db.Query(qstr, args...)
	if err != nil {
		t := reflect.Indirect(reflect.ValueOf(typ)).Type()
		f, _ := m.getMap(t)

		return &Rows{
			nil,
			f,
			[]string{},
			err,
			t,
		}
	}

	return m.Proxify(dbrows, typ)
}

func (m *Manager) makeInsert(table string, data interface{}) (qstr string, vals []interface{}, err error) {
	if vals, err = m.Val(data); err != nil {
		return
	}

	var cols []string
	if cols, err = m.ColIns(data, ""); err != nil {
		return
	}

	holders := "?" + strings.Repeat(",?", len(cols)-1)
	qstr = fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		table,
		strings.Join(cols, ","),
		holders,
	)
	return
}

// Insert inserts data into table.
// It will skip columns with "ai" tag
func (m *Manager) Insert(table string, data interface{}) (sql.Result, error) {
	qstr, vals, err := m.makeInsert(table, data)
	if err != nil {
		return nil, err
	}
	return m.db.Exec(qstr, vals...)
}

func (m *Manager) makeUpdate(table string, data interface{}, where string, whereargs []interface{}) (qstr string, vals []interface{}, err error) {
	if vals, err = m.Val(data); err != nil {
		return
	}

	var cols []string
	if cols, err = m.ColIns(data, ""); err != nil {
		return
	}

	qstr = fmt.Sprintf(
		`UPDATE %s SET %s WHERE %s`,
		table,
		strings.Join(cols, "=?,")+"=?",
		where,
	)
	if len(whereargs) > 0 {
		vals = append(vals, whereargs...)
	}
	return
}

// Update updates data in db.
func (m *Manager) Update(table string, data interface{}, where string, whereargs ...interface{}) (sql.Result, error) {
	qstr, vals, err := m.makeUpdate(table, data, where, whereargs)
	if err != nil {
		return nil, err
	}
	return m.db.Exec(qstr, vals...)
}

func (m *Manager) makeDelete(table string, data interface{}) (qstr string, vals []interface{}, err error) {
	if vals, err = m.Val(data); err != nil {
		return
	}

	var cols []string
	if cols, err = m.ColIns(data, ""); err != nil {
		return
	}

	qstr = fmt.Sprintf(
		`DELETE FROM %s WHERE %s`,
		table,
		strings.Join(cols, "=? AND ")+"=?",
	)

	return
}

// Delete deletes data in db.
func (m *Manager) Delete(table string, data interface{}) (sql.Result, error) {
	qstr, vals, err := m.makeDelete(table, data)
	if err != nil {
		return nil, err
	}
	return m.db.Exec(qstr, vals...)
}

// Begin creates a transaction
func (m *Manager) Begin() (*Tx, error) {
	tx, err := m.db.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx, m}, nil
}
