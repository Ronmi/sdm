package sdm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	IndexTypeIndex   = "idx"
	IndexTypeUnique  = "uniq"
	IndexTypePrimary = "pri"
)

// IndexDef represents defination of an index
type IndexDef struct {
	Type string
	Name string
	Cols []string
}

// return -1 if not found
func findIndexByName(i *[]IndexDef, name, typ string) int {
	arr := *i
	for idx, _ := range arr {
		if arr[idx].Name == name {
			return idx
		}
	}

	idx := len(arr)
	*i = append(arr, IndexDef{
		Type: typ,
		Name: name,
	})
	return idx
}

// ColumnDef represents defination of a column, for internal use only
type ColumnDef struct {
	ID   int    // field id
	AI   bool   // auto increment
	Name string // column name
}

// Manager is just manager. any question?
type Manager struct {
	indexes map[reflect.Type][]IndexDef
	columns map[reflect.Type]map[string]*ColumnDef
	fields  map[reflect.Type][]*ColumnDef
	table   map[reflect.Type]string
	lock    sync.RWMutex
	db      *sql.DB
}

// New create sdm manager
func New(db *sql.DB) *Manager {
	return &Manager{
		map[reflect.Type][]IndexDef{},
		map[reflect.Type]map[string]*ColumnDef{},
		map[reflect.Type][]*ColumnDef{},
		map[reflect.Type]string{},
		sync.RWMutex{},
		db,
	}
}

func (m *Manager) has(t reflect.Type) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if _, ok := m.fields[t]; ok {
		return true
	}
	return false
}

// Register parses and caches a type into SDM
func (m *Manager) Register(i interface{}, tableName string) (err error) {
	t := reflect.Indirect(reflect.ValueOf(i)).Type()
	if m.has(t) {
		return
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("sdm: %s is not a struct type", t.String())
	}

	mps := make([]*ColumnDef, 0, t.NumField())
	idx := make(map[string]*ColumnDef)
	indexes := []IndexDef{}

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

		fdef := &ColumnDef{ID: i, Name: col}
		for _, tag := range tags {

			if tag == "ai" {
				fdef.AI = true
				continue
			}

			for _, t := range []string{IndexTypeIndex, IndexTypePrimary, IndexTypeUnique} {
				l := len(t) + 1
				if len(tag) <= l {
					continue
				}

				if !strings.HasPrefix(tag, t+"_") {
					continue
				}

				name := tag[l:]
				pos := findIndexByName(&indexes, name, t)
				indexes[pos].Cols = append(indexes[pos].Cols, col)
				break
			}
		}

		mps = append(mps, fdef)
		idx[col] = fdef
	}

	m.indexes[t] = indexes
	m.columns[t] = idx
	m.fields[t] = mps
	m.table[t] = tableName
	return
}

func (m *Manager) getDef(t reflect.Type) (ret []*ColumnDef, err error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	ret, ok := m.fields[t]
	if !ok {
		err = errors.New("info of type " + t.String() + " not found")
	}
	return
}

func (m *Manager) getMap(t reflect.Type) (ret map[string]*ColumnDef, err error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	ret, ok := m.columns[t]
	if !ok {
		err = errors.New("info of type " + t.String() + " not found")
	}
	return
}

func (m *Manager) getTable(t reflect.Type) (ret string, err error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	ret, ok := m.table[t]
	if !ok {
		err = errors.New("info of type " + t.String() + " not found")
	}
	return
}

// Col returns a list of columns in sql format, including AUTO INCREMENT columns
func (m *Manager) Col(data interface{}) (ret []string, err error) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	fdef, err := m.getDef(t)
	if err != nil {
		return
	}
	ret = make([]string, 0, len(fdef))

	for _, f := range fdef {
		c := f.Name
		ret = append(ret, c)
	}

	return
}

// ColIns returns a list of columns in sql format, excluding AUTO INCREMENT columns
func (m *Manager) ColIns(data interface{}) (ret []string, err error) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	fdef, err := m.getDef(t)
	if err != nil {
		return
	}
	ret = make([]string, 0, len(fdef))

	for _, f := range fdef {
		if f.AI {
			continue
		}
		c := f.Name
		ret = append(ret, c)
	}

	return
}

// Val converts struct to value array
func (m *Manager) Val(data interface{}) ([]interface{}, error) {
	var ret []interface{}
	v := reflect.Indirect(reflect.ValueOf(data))
	t := v.Type()
	fdef, err := m.getDef(t)
	if err != nil {
		return nil, err
	}

	ret = make([]interface{}, len(fdef))
	for k, f := range fdef {
		ret[k] = v.Field(f.ID).Interface()
	}
	return ret, nil
}

// ValIns converts struct to value array, skipping auto increment fields
func (m *Manager) ValIns(data interface{}) ([]interface{}, error) {
	var ret []interface{}
	v := reflect.Indirect(reflect.ValueOf(data))
	t := v.Type()
	fdef, err := m.getDef(t)
	if err != nil {
		return nil, err
	}

	ret = make([]interface{}, 0, len(fdef))
	for _, f := range fdef {
		if f.AI {
			continue
		}
		ret = append(ret, v.Field(f.ID).Interface())
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

// Exec warps sql.DB.Exec
func (m *Manager) Exec(qstr string, args ...interface{}) (sql.Result, error) {
	return m.db.Exec(qstr, args...)
}

// Build constructs sql query, and executes it with Exec
//
// There are 3 special place holders to use in template, each for exactly one time most:
//
//   %cols%           col, col, col       (must use with %vals%)
//   %vals%           ?, ?, ?             (must use with %cols%)
//   %combined%       col=?, col=?, col=? (must not use with other two)
//
// Rules above is not validated, YOU MUST TAKE CARE OF IT YOURSELF.
//
// Custom parameters are not supported, use Exec instead.
func (m *Manager) Build(data interface{}, tmpl string) (sql.Result, error) {
	cols, err := m.Col(data)
	if err != nil {
		return nil, err
	}
	vals, err := m.Val(data)
	if err != nil {
		return nil, err
	}
	sz := len(vals)
	if sz < 1 {
		sz = 1
	}

	tmpl = strings.Replace(tmpl, "%cols%", strings.Join(cols, ","), 1)
	tmpl = strings.Replace(tmpl, "%vals%", "?"+strings.Repeat(",?", sz-1), 1)
	tmpl = strings.Replace(tmpl, "%combined%", strings.Join(cols, "=?,")+"=?", 1)
	return m.Exec(tmpl, vals...)
}

func (m *Manager) makeInsert(data interface{}) (qstr string, vals []interface{}, err error) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	table, err := m.getTable(t)
	if err != nil {
		return
	}
	if vals, err = m.ValIns(data); err != nil {
		return
	}

	var cols []string
	if cols, err = m.ColIns(data); err != nil {
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
func (m *Manager) Insert(data interface{}) (sql.Result, error) {
	qstr, vals, err := m.makeInsert(data)
	if err != nil {
		return nil, err
	}
	return m.db.Exec(qstr, vals...)
}

func (m *Manager) makeUpdate(data interface{}, where string, whereargs []interface{}) (qstr string, vals []interface{}, err error) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	table, err := m.getTable(t)
	if err != nil {
		return
	}

	if vals, err = m.Val(data); err != nil {
		return
	}

	var cols []string
	if cols, err = m.ColIns(data); err != nil {
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
func (m *Manager) Update(data interface{}, where string, whereargs ...interface{}) (sql.Result, error) {
	qstr, vals, err := m.makeUpdate(data, where, whereargs)
	if err != nil {
		return nil, err
	}
	return m.db.Exec(qstr, vals...)
}

func (m *Manager) makeDelete(data interface{}) (qstr string, vals []interface{}, err error) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	table, err := m.getTable(t)
	if err != nil {
		return
	}

	if vals, err = m.Val(data); err != nil {
		return
	}

	var cols []string
	if cols, err = m.ColIns(data); err != nil {
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
func (m *Manager) Delete(data interface{}) (sql.Result, error) {
	qstr, vals, err := m.makeDelete(data)
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

// BulkInsert creates a generator to generate long statement which inserts many data at once
func (m *Manager) BulkInsert(typ interface{}) (Bulk, error) {
	t := reflect.Indirect(reflect.ValueOf(typ)).Type()
	table, err := m.getTable(t)
	if err != nil {
		return nil, err
	}
	def, err := m.getDef(t)
	if err != nil {
		return nil, err
	}

	return &bulkInsert{
		newBulkInfo(table, t, def),
	}, nil
}

// BulkDelete creates a generator to generate long statement which deletes many data at once
func (m *Manager) BulkDelete(typ interface{}) (Bulk, error) {
	t := reflect.Indirect(reflect.ValueOf(typ)).Type()
	table, err := m.getTable(t)
	if err != nil {
		return nil, err
	}
	def, err := m.getDef(t)
	if err != nil {
		return nil, err
	}

	return &bulkDelete{
		newBulkInfo(table, t, def),
	}, nil
}

// RunBulk executes a bulk operation
func (m *Manager) RunBulk(b Bulk) (sql.Result, error) {
	if b.Len() < 1 {
		return nil, nil
	}
	qstr, vals := b.Make()
	return m.Connection().Exec(qstr, vals...)
}
