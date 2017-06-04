package sdm

import (
	"fmt"
	"reflect"
	"strings"

	"git.ronmi.tw/ronmi/sdm/driver"
)

type Bulk interface {
	// Add some elements to be bulk operated
	Add(data ...interface{}) error
	// current data length, Manager and Tx will not do execution if Len returns 0
	Len() int

	// Make sql statement
	Make() (qstr string, vals []interface{})
}

type bulkinfo struct {
	table string
	typ   reflect.Type
	def   []driver.Column
	data  []interface{}
}

func newBulkInfo(table string, typ reflect.Type, def []driver.Column) *bulkinfo {
	return &bulkinfo{
		table,
		typ,
		def,
		[]interface{}{},
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

func (b *bulkInsert) Make() (string, []interface{}) {
	ids := make([]int, 0, len(b.def))
	cols := make([]string, 0, len(b.def))
	placeholders := make([]string, 0, len(b.data))
	vals := make([]interface{}, 0, len(b.data)*len(cols))

	paramarr := make([]string, 0, len(b.data))
	for _, v := range b.def {
		if v.AI {
			continue
		}

		ids = append(ids, v.ID)
		cols = append(cols, v.Name)
		paramarr = append(paramarr, "?")
	}
	paramstr := "(" + strings.Join(paramarr, ",") + ")"

	for _, v := range b.data {
		placeholders = append(placeholders, paramstr)
		val := reflect.ValueOf(v)

		for _, fid := range ids {
			vals = append(vals, val.Field(fid).Interface())
		}
	}

	qstr := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		b.table,
		strings.Join(cols, ","),
		strings.Join(placeholders, ","),
	)

	return qstr, vals
}

type bulkDelete struct {
	*bulkinfo
}

func (b *bulkDelete) Make() (string, []interface{}) {
	ids := make([]int, 0, len(b.def))
	cols := make([]string, 0, len(b.def))
	placeholders := make([]string, 0, len(b.data))
	vals := make([]interface{}, 0, len(b.data)*len(cols))

	paramarr := make([]string, 0, len(b.data))
	for _, v := range b.def {
		ids = append(ids, v.ID)
		cols = append(cols, v.Name)
		paramarr = append(paramarr, v.Name+"=?")
	}
	paramstr := "(" + strings.Join(paramarr, " AND ") + ")"

	for _, v := range b.data {
		placeholders = append(placeholders, paramstr)
		val := reflect.ValueOf(v)

		for _, fid := range ids {
			vals = append(vals, val.Field(fid).Interface())
		}
	}

	qstr := fmt.Sprintf(
		`DELETE FROM %s WHERE %s`,
		b.table,
		strings.Join(placeholders, " OR "),
	)

	return qstr, vals
}
