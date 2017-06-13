package sdm

import (
	"database/sql"
	"reflect"
)

// Tx warps Manager in transaction
type Tx struct {
	tx *sql.Tx
	m  *Manager
}

// Query makes SQL query and proxies it
func (tx *Tx) Query(typ interface{}, qstr string, args ...interface{}) *Rows {
	dbrows, err := tx.tx.Query(qstr, args...)
	if err != nil {
		t := reflect.Indirect(reflect.ValueOf(typ)).Type()
		f, _ := tx.m.getMap(t)

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
// It will skip columns with "ai" tag
func (tx *Tx) Insert(data interface{}) (sql.Result, error) {
	qstr, vals, err := tx.m.makeInsert(data)
	if err != nil {
		return nil, err
	}
	return tx.tx.Exec(qstr, vals...)
}

// Update updates data in db.
func (tx *Tx) Update(data interface{}, where string, whereargs ...interface{}) (sql.Result, error) {
	qstr, vals, err := tx.m.makeUpdate(data, where, whereargs)
	if err != nil {
		return nil, err
	}
	return tx.tx.Exec(qstr, vals...)
}

// Delete deletes data in db.
func (tx *Tx) Delete(data interface{}) (sql.Result, error) {
	qstr, vals, err := tx.m.makeDelete(data)
	if err != nil {
		return nil, err
	}
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

// RunBulk executes a bulk operation
func (tx *Tx) RunBulk(b Bulk) (sql.Result, error) {
	if b.Len() < 1 {
		return nil, nil
	}
	qstr, vals := b.Make()
	return tx.Tx().Exec(qstr, vals...)
}
