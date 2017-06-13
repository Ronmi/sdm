package sdm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Ronmi/sdm/driver"
)

// Bulk represents batch operations
type Bulk interface {
	// Add some elements to be bulk operated
	Add(data ...interface{}) error
	// current data length, Manager and Tx will not do execution if Len returns 0
	Len() int

	// Make sql statement.
	Make() (qstr []string, vals [][]interface{})
}

type bulkinfo struct {
	table string
	typ   reflect.Type
	data  []interface{}
	m     *Manager
}

func newBulkInfo(table string, typ reflect.Type, m *Manager) *bulkinfo {
	return &bulkinfo{
		table,
		typ,
		[]interface{}{},
		m,
	}
}

func (b *bulkinfo) Add(data ...interface{}) error {
	for _, v := range data {
		if t := reflect.Indirect(reflect.ValueOf(v)).Type(); t != b.typ {
			return fmt.Errorf("sdm: bulk: type error: expecting %s, got %s", b.typ, t)
		}
	}

	b.data = append(b.data, data...)
	return nil
}

func (b *bulkinfo) Len() int {
	return len(b.data)
}

type bulkInsert struct {
	*bulkinfo
}

func (b *bulkInsert) Make() ([]string, [][]interface{}) {
	if len(b.data) < 1 {
		return []string{}, [][]interface{}{}
	}

	data := b.data[0]
	cols := b.m.ColIns(data)
	l := len(cols)
	placeholders := make([]string, 0, l)
	hds := b.m.HolderIns(data)
	vals := make([]interface{}, 0, len(b.data)*l)

	paramstr := "(" + strings.Join(hds, ",") + ")"

	for _, v := range b.data {
		placeholders = append(placeholders, paramstr)
		i := b.m.ValIns(v)
		vals = append(vals, i...)
	}

	qstr := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		b.m.drv.Quote(b.table),
		strings.Join(cols, ","),
		strings.Join(placeholders, ","),
	)

	return []string{qstr}, [][]interface{}{vals}
}

type bulkDelete struct {
	*bulkinfo
}

func (b *bulkDelete) Make() ([]string, [][]interface{}) {
	if len(b.data) < 1 {
		return []string{""}, [][]interface{}{}
	}

	data := b.data[0]
	cols := b.m.Col(data, driver.QWhere)
	l := len(cols)
	hds := b.m.Holder(data)
	placeholders := make([]string, 0, len(b.data))
	com := make([]string, l)
	vals := make([]interface{}, 0, len(b.data)*l)

	for k, v := range cols {
		com[k] = v + "=" + hds[k]
	}
	paramstr := "(" + strings.Join(com, " AND ") + ")"

	for _, v := range b.data {
		placeholders = append(placeholders, paramstr)
		i := b.m.Val(v)
		vals = append(vals, i...)
	}

	qstr := fmt.Sprintf(
		`DELETE FROM %s WHERE %s`,
		b.m.drv.Quote(b.table),
		strings.Join(placeholders, " OR "),
	)

	return []string{qstr}, [][]interface{}{vals}
}
