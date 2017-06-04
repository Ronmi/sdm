package sqlite3

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"git.ronmi.tw/ronmi/sdm/driver"
)

var timeType reflect.Type

func init() {
	timeType = reflect.TypeOf(time.Time{})
}

func getType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Bool:
		return "INT"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.String:
		return "TEXT"
	case reflect.Array, reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "BLOB"
		}
	case reflect.Struct:
		if t == timeType {
			return "DATETIME"
		}
	}

	panic("sdm: driver: sqlite3: unsupported type " + t.String())
}

func quote(name string) string {
	return `'` + strings.Replace(name, `'`, "\\'", -1) + `'`
}

func createTableColumnSQL(typ reflect.Type, cols []driver.Column, indexes []driver.Index) string {
	ret := make([]string, 0, len(cols)+len(indexes))

	hasAI := false

	for _, c := range cols {
		def := quote(c.Name) + ` ` + getType(typ.Field(c.ID).Type)
		if c.AI {
			hasAI = true
			// in sqlite, auto increment must pair with primary key
			name := typ.Name() + "_pk"
			for _, i := range indexes {
				if i.Type == driver.IndexTypePrimary {
					name = i.Name
					break
				}
			}
			def += " CONSTRAINT " + quote(name) + " PRIMARY KEY AUTOINCREMENT"
		}
		ret = append(ret, def)
	}

	for _, i := range indexes {
		var def string
		quoted := make([]string, len(i.Cols))
		for k, v := range i.Cols {
			quoted[k] = quote(v)
		}

		switch i.Type {
		case driver.IndexTypeIndex:
		case driver.IndexTypePrimary:
			if hasAI {
				continue
			}
			def = fmt.Sprintf(
				"CONSTRAINT %s PRIMARY KEY (%s)",
				quote(i.Name),
				strings.Join(quoted, ","),
			)
		case driver.IndexTypeUnique:
			def = fmt.Sprintf(
				"CONSTRAINT %s UNIQUE (%s)",
				quote(i.Name),
				strings.Join(quoted, ","),
			)
		}

		ret = append(ret, def)
	}

	return strings.Join(ret, ",")
}

type drv struct {
}

func (d drv) CreateTable(db *sql.DB, name string, typ reflect.Type, cols []driver.Column, indexes []driver.Index) (sql.Result, error) {
	qstr := fmt.Sprintf(
		"CREATE TABLE '%s' (%s)",
		name,
		createTableColumnSQL(typ, cols, indexes),
	)

	return db.Exec(qstr)
}

func (d drv) CreateTableNotExist(db *sql.DB, name string, typ reflect.Type, cols []driver.Column, indexes []driver.Index) (sql.Result, error) {
	qstr := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS '%s' (%s)",
		name,
		createTableColumnSQL(typ, cols, indexes),
	)

	return db.Exec(qstr)
}

func (d drv) Quote(name string) string {
	return quote(name)
}

// New creates a driver instance
func New() driver.Driver {
	return drv{}
}
