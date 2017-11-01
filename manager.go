package sdm

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/Ronmi/sdm/driver"
)

// return -1 if not found
func findIndexByName(i *[]driver.Index, name, typ string) int {
	arr := *i
	for idx, _ := range arr {
		if arr[idx].Name == name {
			return idx
		}
	}

	idx := len(arr)
	*i = append(arr, driver.Index{
		Type: typ,
		Name: name,
	})
	return idx
}

type tableInfo struct {
	Table   string
	Indexes []driver.Index
	Defs    map[string]driver.Column
	Fields  []driver.Column
	PKIndex int // < 0 if not exists
}

// Manager is just manager. any question?
//
// Manager IS NOT ZERO VALUE SAFE. Always create with New()
//
// Most methods panic when target type is not registered.
// You can set AutoReg to prevent panic, but it's not recommended.
type Manager struct {
	// Automatic register new type with Reg()
	// Use it with care, since Reg() panics if type has no SDM tag.
	AutoReg bool

	info map[reflect.Type]*tableInfo
	lock sync.RWMutex
	db   *sql.DB
	drv  driver.Driver
}

// New create sdm manager
//
// Much like database.sql DSN, the driverStr is a plain string specifies which
// driver to use (and parameters to be passed).
//
// Typical driverStr is "driverName" or "driverName:param1=value1;param2=value2",
// basiclly same as DSN format.
func New(db *sql.DB, driverStr string) *Manager {
	sdmDriver := driver.GetDriver(driverStr)

	return &Manager{
		false,
		map[reflect.Type]*tableInfo{},
		sync.RWMutex{},
		db,
		sdmDriver,
	}
}

// Driver returns the driver we're using
func (m *Manager) Driver() driver.Driver {
	return m.drv
}

func (m *Manager) has(t reflect.Type) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if _, ok := m.info[t]; ok {
		return true
	}
	return false
}

// Reg calls Register for you, just for short. It panics at first error
//
// It will use struct name (convert to lower case) as table name.
func (m *Manager) Reg(data ...interface{}) {
	for _, i := range data {
		t := reflect.Indirect(reflect.ValueOf(i)).Type()
		m.register(t, strings.ToLower(t.Name()))
	}
}

// Register parses and caches a type into SDM. It panics at first error
func (m *Manager) Register(i interface{}, tableName string) {
	t := reflect.Indirect(reflect.ValueOf(i)).Type()
	m.register(t, tableName)
}

func (m *Manager) register(t reflect.Type, tableName string) {
	if m.has(t) {
		return
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	if t.Kind() != reflect.Struct {
		panic(errors.New("sdm: " + t.String() + " is not a struct type"))
	}

	mps := make([]driver.Column, 0, t.NumField())
	idx := make(map[string]driver.Column)
	indexes := []driver.Index{}

	// some flags
	havePK := false
	var lastAIField *driver.Column
	aiFieldCnt := 0

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

		fdef := driver.Column{ID: i, Name: col}
		for _, tag := range tags {

			if tag == "ai" {
				fdef.AI = true
				aiFieldCnt++
				lastAIField = &fdef
				continue
			}

			for _, t := range []string{driver.IndexTypeIndex, driver.IndexTypePrimary, driver.IndexTypeUnique} {
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
				if t == driver.IndexTypePrimary {
					havePK = true
				}
				break
			}
		}

		mps = append(mps, fdef)
		idx[col] = fdef
	}

	if !havePK && aiFieldCnt == 1 {
		// Use AI field as primary key
		indexes = append(indexes, driver.Index{
			Type: driver.IndexTypePrimary,
			Name: tableName + "_" + "pk",
			Cols: []string{lastAIField.Name},
		})
		havePK = true
	}

	pk := -1
	if havePK {
		for x, i := range indexes {
			if i.Type == driver.IndexTypePrimary {
				pk = x
				break
			}
		}
	}

	m.info[t] = &tableInfo{
		Table:   tableName,
		Indexes: indexes,
		Defs:    idx,
		Fields:  mps,
		PKIndex: pk,
	}
}

func (m *Manager) getInfo(t reflect.Type) (ret *tableInfo) {
	m.lock.RLock()
	ret, ok := m.info[t]
	m.lock.RUnlock()

	if !ok {
		if !m.AutoReg {
			panic("info of type " + t.String() + " not found")
		}

		m.register(t, strings.ToLower(t.Name()))
		return m.getInfo(t)
	}
	return
}

func (m *Manager) getPK(t reflect.Type) (ret driver.Index, ok bool) {
	info := m.getInfo(t)

	if info.PKIndex < 0 {
		return
	}

	return info.Indexes[info.PKIndex], true
}

// GetTable returns table name of specified type.
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) GetTable(t reflect.Type) (ret string) {
	info := m.getInfo(t)
	return info.Table
}

// Col returns a list of columns in sql format, including AUTO INCREMENT columns.
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) Col(data interface{}, qType driver.QuotingType) (ret []string) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	info := m.getInfo(t)
	fdef := info.Fields
	ret = make([]string, 0, len(fdef))

	for _, f := range fdef {
		c := m.drv.Col(info.Table, f.Name, qType)
		ret = append(ret, c)
	}

	return
}

// ColSel returns a list of columns in sql format, suitable for SELECT query
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) ColSel(data interface{}) (ret []string) {
	v := reflect.Indirect(reflect.ValueOf(data))
	t := v.Type()
	info := m.getInfo(t)
	fdef := info.Fields
	ret = make([]string, 0, len(fdef))

	for _, f := range fdef {
		c := m.drv.Col(info.Table, f.Name, driver.QSelect)
		ret = append(ret, c)
	}

	return
}

// ColIns returns a list of columns in sql format, excluding AUTO INCREMENT columns
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) ColIns(data interface{}) (ret []string) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	info := m.getInfo(t)
	fdef := info.Fields
	ret = make([]string, 0, len(fdef))

	for _, f := range fdef {
		if f.AI {
			continue
		}
		c := m.drv.Col(info.Table, f.Name, driver.QInsert)
		ret = append(ret, c)
	}

	return
}

// Val converts struct to value array
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) Val(data interface{}) []interface{} {
	var ret []interface{}
	v := reflect.Indirect(reflect.ValueOf(data))
	t := v.Type()
	fdef := m.getInfo(t).Fields

	ret = make([]interface{}, len(fdef))
	for k, f := range fdef {
		vfield := v.Field(f.ID)
		if vsql, ok := m.drv.GetValuer(vfield); ok {
			ret[k] = vsql
		} else {
			ret[k] = vfield.Interface()
		}
	}
	return ret
}

// ValIns converts struct to value array, skipping auto increment fields
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) ValIns(data interface{}) []interface{} {
	var ret []interface{}
	v := reflect.Indirect(reflect.ValueOf(data))
	t := v.Type()
	fdef := m.getInfo(t).Fields

	ret = make([]interface{}, 0, len(fdef))
	for _, f := range fdef {
		if f.AI {
			continue
		}

		vfield := v.Field(f.ID)
		var res interface{}
		if vsql, ok := m.drv.GetValuer(vfield); ok {
			res = vsql
		} else {
			res = vfield.Interface()
		}
		ret = append(ret, res)
	}
	return ret
}

// Holder converts struct to SQL unnamed placeholders
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) Holder(data interface{}) []string {
	v := reflect.Indirect(reflect.ValueOf(data))
	t := v.Type()
	fdef := m.getInfo(t).Fields

	ret := make([]string, len(fdef))
	for k, f := range fdef {
		vfield := v.Field(f.ID)
		ret[k] = m.drv.GetPlaceholder(vfield.Type())
	}
	return ret
}

// HolderIns converts struct to SQL unnamed placeholders and skip auto increment fields
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) HolderIns(data interface{}) []string {
	v := reflect.Indirect(reflect.ValueOf(data))
	t := v.Type()
	fdef := m.getInfo(t).Fields

	ret := make([]string, 0, len(fdef))
	for _, f := range fdef {
		if f.AI {
			continue
		}

		vfield := v.Field(f.ID)
		ret = append(ret, m.drv.GetPlaceholder(vfield.Type()))
	}

	return ret
}

// Connection returns stored *sql.DB
func (m *Manager) Connection() *sql.DB {
	return m.db
}

// Prepare wraps sql.DB.Prepare
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) Prepare(data interface{}, qstr string) (*Stmt, error) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	f := m.getInfo(t).Defs

	stmt, e := m.Connection().Prepare(qstr)
	return &Stmt{
		stmt: stmt,
		def:  f,
		t:    t,
		drv:  m.drv,
	}, e
}

// PrepareSQL builds sql query with BuildSQL(), then prepare it
//
// It is faster than Prepare(data, BuildSQL()).
func (m *Manager) PrepareSQL(data interface{}, tmpl string, qType driver.QuotingType) (*Stmt, error) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	info := m.getInfo(t)
	qstr := m.BuildSQL(data, tmpl, qType)

	stmt, e := m.Connection().Prepare(qstr)
	return &Stmt{
		stmt:    stmt,
		def:     info.Defs,
		t:       t,
		drv:     m.drv,
		columns: m.Col(data, qType),
	}, e
}

// Proxify proxies needed methods of sql.Rows
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) Proxify(r *sql.Rows, data interface{}) *Rows {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	f := m.getInfo(t).Defs
	c, e := r.Columns()

	return &Rows{
		r,
		f,
		c,
		e,
		t,
		m.drv,
	}
}

func (m *Manager) createErrorRow(typ reflect.Type, err error) *Rows {
	f := m.getInfo(typ).Defs

	return &Rows{nil, f, []string{}, err, typ, m.drv}
}

// Query makes SQL query and proxies it.
// It panics if type is not registered and auto register is not enabled.
//
// You can use "%table%" as placeholder for table name, %cols% for column names
func (m *Manager) Query(typ interface{}, qstr string, args ...interface{}) *Rows {
	t := reflect.Indirect(reflect.ValueOf(typ)).Type()
	if strings.Index(qstr, "%table%") != -1 {
		table := m.GetTable(t)
		qstr = strings.Replace(qstr, "%table%", m.drv.Quote(table), -1)
	}

	if strings.Index(qstr, "%cols%") != -1 {
		cols := m.ColSel(typ)
		qstr = strings.Replace(qstr, "%cols%", strings.Join(cols, ","), 1)
	}

	dbrows, err := m.db.Query(qstr, args...)
	if err != nil {
		return m.createErrorRow(t, err)
	}

	return m.Proxify(dbrows, typ)
}

// QueryRow makes SQL query, fill first row into data and discard the rest
// It panics if type is not registered and auto register is not enabled.
//
// QueryRow is a wrapper for Query, see Query() for detail.
func (m *Manager) QueryRow(data interface{}, qstr string, args ...interface{}) error {
	rows := m.Query(data, qstr, args)
	defer rows.Close()

	rows.Next()
	rows.Scan(data)
	return rows.Err()
}

// LoadSimple loads simple table
//
// Simple table is a table with single-column primary key
func (m *Manager) LoadSimple(data, pkVal interface{}) error {
	qstr := `SELECT %cols% FROM %table% WHERE `
	v := reflect.Indirect(reflect.ValueOf(data))
	if !v.CanSet() {
		return errors.New("Need pointer type to scan value")
	}
	t := v.Type()
	pk, ok := m.getPK(t)
	if !ok {
		return errors.New(t.Name() + " does not have a primary key")
	}

	if len(pk.Cols) != 1 {
		return errors.New(t.Name() + " has more than 1 column")
	}

	cols := m.getInfo(t).Defs
	f := v.Field(cols[pk.Cols[0]].ID)
	qstr += m.drv.Col(m.GetTable(t), pk.Cols[0], driver.QWhere) + `=` + m.drv.GetPlaceholder(f.Type())
	rows := m.Query(data, qstr, pkVal)
	for rows.Next() {
		rows.Scan(data)
	}

	return rows.Err()
}

// Exec wraps sql.DB.Exec
func (m *Manager) Exec(qstr string, args ...interface{}) (sql.Result, error) {
	return m.db.Exec(qstr, args...)
}

// BuildSQL constructs sql query
// It panics if type is not registered and auto register is not enabled.
//
// You can use "%table%" as placeholder for table name.
// There are 3 special place holders to use in template, each for exactly one time most:
//
//   %cols%           col, col, col       (must use with %vals%)
//   %vals%           ?, ?, ?             (must use with %cols%)
//   %combined%       col=?, col=?, col=? (must not use with other two)
//
// Rules above is not validated, YOU MUST TAKE CARE OF IT YOURSELF.
//
// Custom parameters are not supported, use Exec instead.
//
// Order of columns is not guaranteed, use Val/ValIns to generate it. For example:
//
//     qstr := m.BuildSQL(myStruct, `REPLACE INTO %table% (%cols%) VALUES (%vals%)`, driver.QInsert)
//     m.Exec(qstr, m.ValIns(myStruct))
func (m *Manager) BuildSQL(data interface{}, tmpl string, qType driver.QuotingType) (qstr string) {
	cols := m.Col(data, qType)
	sz := len(cols)

	hd := m.Holder(data)
	com := make([]string, sz)
	for k, v := range hd {
		com[k] = cols[k] + "=" + v
	}

	if strings.Index(tmpl, "%table%") != -1 {
		table := m.GetTable(reflect.Indirect(reflect.ValueOf(data)).Type())
		tmpl = strings.Replace(tmpl, "%table%", m.drv.Quote(table), -1)
	}

	tmpl = strings.Replace(tmpl, "%cols%", strings.Join(cols, ","), 1)
	tmpl = strings.Replace(tmpl, "%vals%", strings.Join(hd, ","), 1)
	tmpl = strings.Replace(tmpl, "%combined%", strings.Join(com, ","), 1)
	return tmpl
}

// Build is shortcut of building custom query and executes it
//
// It is identical to following codes
//
//     qstr := m.BuildSQL(myStruct, `REPLACE INTO %table% (%cols%) VALUES (%vals%)`, driver.QInsert)
//     return m.Exec(qstr, m.Val(myStruct)...)
func (m *Manager) Build(data interface{}, tmpl string, qType driver.QuotingType) (sql.Result, error) {
	return m.Exec(
		m.BuildSQL(data, tmpl, qType),
		m.Val(data)...,
	)
}

func (m *Manager) makeInsert(data interface{}) (qstr string, vals []interface{}) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	table := m.GetTable(t)
	vals = m.ValIns(data)
	cols := m.ColIns(data)
	hd := m.HolderIns(data)
	qstr = `INSERT INTO ` +
		m.drv.Quote(table) +
		`(` + strings.Join(cols, ",") + `) VALUES ` +
		`(` + strings.Join(hd, ",") + `)`
	return
}

func (m *Manager) tryFillPK(data interface{}, res sql.Result) {
	v := reflect.Indirect(reflect.ValueOf(data))
	if !v.CanSet() {
		return
	}

	t := v.Type()
	info := m.getInfo(t)
	indexes := info.Indexes
	cols := info.Defs
	for _, idx := range indexes {
		if idx.Type != driver.IndexTypePrimary {
			continue
		}

		if len(idx.Cols) != 1 {
			// multiple column key, skip
			return
		}

		col := cols[idx.Cols[0]]
		if !col.AI {
			// skip if primary key is not auto increment
			return
		}

		vf := v.Field(col.ID)
		id, err := res.LastInsertId()
		if err != nil || id < 1 {
			return
		}

		vf.SetInt(id)
	}
}

// Insert inserts data into table.
// It panics if type is not registered and auto register is not enabled.
//
// It will skip columns with "ai" tag.
//
// If following restrstions are fulfilled, primary key is filled back to data:
//
//   - data is settable (pointer type)
//   - DB and DB driver supports sql.Result.LastInsertId()
//   - Table has exactly one primary key.
//   - Prmary key contains exactly ne column.
//   - The column is AUTO INCREMENT enabled.
func (m *Manager) Insert(data interface{}) (sql.Result, error) {
	qstr, vals := m.makeInsert(data)
	res, err := m.db.Exec(qstr, vals...)
	if err == nil {
		m.tryFillPK(data, res)
	}
	return res, err
}

func (m *Manager) makeUpdate(data interface{}, where string, whereargs []interface{}) (qstr string, vals []interface{}) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	table := m.GetTable(t)
	vals = m.ValIns(data)
	cols := m.ColIns(data)
	hd := m.HolderIns(data)
	com := make([]string, len(hd))
	for k, v := range hd {
		com[k] = cols[k] + "=" + v
	}

	qstr = `UPDATE ` + m.drv.Quote(table) +
		` SET ` + strings.Join(com, ",") +
		` WHERE ` + where
	if len(whereargs) > 0 {
		vals = append(vals, whereargs...)
	}
	return
}

// Update updates data in db.
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) Update(data interface{}, where string, whereargs ...interface{}) (sql.Result, error) {
	qstr, vals := m.makeUpdate(data, where, whereargs)
	return m.db.Exec(qstr, vals...)
}

func (m *Manager) makeDelete(data interface{}) (qstr string, vals []interface{}) {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()
	table := m.GetTable(t)
	vals = m.Val(data)
	cols := m.Col(data, driver.QWhere)
	hd := m.Holder(data)
	com := make([]string, len(hd))
	for k, v := range hd {
		com[k] = cols[k] + "=" + v
	}

	qstr = `DELETE FROM ` + m.drv.Quote(table) +
		` WHERE ` + strings.Join(com, " AND ")

	return
}

// Delete deletes data in db.
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) Delete(data interface{}) (sql.Result, error) {
	qstr, vals := m.makeDelete(data)
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
func (m *Manager) BulkInsert(typ interface{}) Bulk {
	t := reflect.Indirect(reflect.ValueOf(typ)).Type()
	table := m.GetTable(t)

	return &bulkInsert{
		newBulkInfo(table, t, m),
	}
}

// BulkDelete creates a generator to generate long statement which deletes many data at once
// It panics if type is not registered and auto register is not enabled.
func (m *Manager) BulkDelete(typ interface{}) Bulk {
	t := reflect.Indirect(reflect.ValueOf(typ)).Type()
	table := m.GetTable(t)

	return &bulkDelete{
		newBulkInfo(table, t, m),
	}
}

// RunBulk executes bulk operations in transaction. It does not return a result when
// success and panics if bulk implementation goes wrong.
func (m *Manager) RunBulk(b Bulk) (sql.Result, error) {
	if b.Len() < 1 {
		return nil, nil
	}

	tx, err := m.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ret, err := tx.RunBulk(b)
	if err == nil {
		err = tx.Commit()
	}

	return ret, err
}
