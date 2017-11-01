package sdm

import (
	"database/sql"
	"fmt"
	"reflect"
)

// Tx warps Manager in transaction
type Tx struct {
	tx *sql.Tx
	m  *Manager
}

// Query makes SQL query and proxies it
// It panics if type is not registered and auto register is not enabled.
func (tx *Tx) Query(typ interface{}, qstr string, args ...interface{}) *Rows {
	dbrows, err := tx.tx.Query(qstr, args...)
	if err != nil {
		t := reflect.Indirect(reflect.ValueOf(typ)).Type()
		f := tx.m.getInfo(t).Defs

		return &Rows{
			nil,
			f,
			[]string{},
			err,
			t,
			tx.m.drv,
		}
	}

	return tx.m.Proxify(dbrows, typ)
}

// Insert inserts data into table.
// It panics if type is not registered and auto register is not enabled.
//
// It will skip columns with "ai" tag
func (tx *Tx) Insert(data interface{}) (sql.Result, error) {
	qstr, vals := tx.m.makeInsert(data)
	res, err := tx.tx.Exec(qstr, vals...)
	tx.m.tryFillPK(data, res)
	return res, err
}

// Update updates data in db.
// It panics if type is not registered and auto register is not enabled.
func (tx *Tx) Update(data interface{}, where string, whereargs ...interface{}) (sql.Result, error) {
	qstr, vals := tx.m.makeUpdate(data, where, whereargs)
	return tx.tx.Exec(qstr, vals...)
}

// Delete deletes data in db.
// It panics if type is not registered and auto register is not enabled.
func (tx *Tx) Delete(data interface{}) (sql.Result, error) {
	qstr, vals := tx.m.makeDelete(data)
	return tx.tx.Exec(qstr, vals...)
}

// Rollback is just same as sql.Tx.Rollback
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

// Commit is just same as sql.Tx.Commit
func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

// Tx returns internal *sql.Tx
func (tx *Tx) Tx() *sql.Tx {
	return tx.tx
}

// RunBulk executes a bulk operation. It does not return a result when
// success and panics if bulk implementation goes wrong.
func (tx *Tx) RunBulk(b Bulk) (sql.Result, error) {
	if b.Len() < 1 {
		return nil, nil
	}
	qstr, vals := b.Make()
	if x, y := len(qstr), len(vals); x != y {
		panic(fmt.Sprintf("Bulk implementation goes wrong: number of query string (%d) and parameters (%d) does not match", x, y))
	}

	for idx, q := range qstr {
		v := vals[idx]
		if res, err := tx.Tx().Exec(q, v...); err != nil {
			return res, err
		}
	}

	return nil, nil
}
