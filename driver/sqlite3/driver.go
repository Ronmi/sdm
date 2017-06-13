package sqlite3

import (
	"database/sql"
	sqlDriver "database/sql/driver"
	"fmt"
	"reflect"
	"strings"

	"github.com/Ronmi/sdm/driver"
)

// possible time format, pass in driver param like time=string
const (
	TimeAsTime   = "time"   // raw time.Time
	TimeAsInt    = "int"    // integer, unix timestamp
	TimeAsString = "string" // string format
)

// with time=string, store this format of time in db
const TimeStringFormat = "2006-01-02T15:04:05-0700"

func getType(typ reflect.Type, timeAs string) string {
	t := driver.ElementType(typ)

	// check []byte first
	switch typ.Kind() {
	case reflect.Array, reflect.Slice:
		if t.Kind() == reflect.Uint8 {
			return "BLOB"
		}
	}

	postfix := ""
	if typ.Kind() != reflect.Ptr {
		postfix = " NOT NULL"
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Bool:
		return "INTEGER" + postfix
	case reflect.Float32, reflect.Float64:
		return "REAL" + postfix
	case reflect.String:
		return "TEXT" + postfix
	default:
		if driver.IsTime(t) {
			ret := "DATETIME"
			switch timeAs {
			case TimeAsString:
				ret = "TEXT"
			case TimeAsInt:
				ret = "INTEGER"
			}
			return ret + postfix
		}
	}

	panic("sdm: driver: sqlite3: unsupported type " + t.String())
}

func quote(name string) string {
	return `"` + strings.Replace(name, `"`, "\\\"", -1) + `"`
}

func createTableColumnSQL(typ reflect.Type, cols []driver.Column, indexes []driver.Index, timeAs string) string {
	ret := make([]string, 0, len(cols)+len(indexes))

	hasAI := false

	for _, c := range cols {
		def := quote(c.Name) + ` ` + getType(typ.Field(c.ID).Type, timeAs)
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
			continue
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
	timeAs string
	driver.Stub
}

func (d drv) CreateTable(db *sql.DB, name string, typ reflect.Type, cols []driver.Column, indexes []driver.Index) (sql.Result, error) {
	qstr := fmt.Sprintf(
		"CREATE TABLE '%s' (%s)",
		name,
		createTableColumnSQL(typ, cols, indexes, d.timeAs),
	)

	return db.Exec(qstr)
}

func (d drv) CreateTableNotExist(db *sql.DB, name string, typ reflect.Type, cols []driver.Column, indexes []driver.Index) (sql.Result, error) {
	qstr := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS '%s' (%s)",
		name,
		createTableColumnSQL(typ, cols, indexes, d.timeAs),
	)

	return db.Exec(qstr)
}

func (d drv) Col(table, col string, kind driver.QuotingType) string {
	if kind == driver.QWhere {
		return quote(table) + "." + quote(col)
	}

	return quote(col)
}

func (d drv) GetScanner(field reflect.Value) (ret sql.Scanner, ok bool) {
	ret, ok = d.getWrapper(field)
	if !ok {
		ret, ok = d.Stub.GetScanner(field)
	}

	return
}

func (d drv) GetValuer(field reflect.Value) (ret sqlDriver.Valuer, ok bool) {
	ret, ok = d.getWrapper(field)
	if !ok {
		ret, ok = d.Stub.GetValuer(field)
	}

	return
}

func (d drv) getWrapper(v reflect.Value) (ret wrapper, ok bool) {
	typ := v.Type()

	if d.timeAs == TimeAsTime {
		// use stub
		return
	}

	if !driver.IsTime(typ) {
		// not time, use stub
		return
	}

	// process Ptr = *time.Time, Struct = time.Time
	k := typ.Kind()
	if d.timeAs == TimeAsString {
		return wrapTimeString{v: v, nullable: k == reflect.Ptr}, true
	}

	return wrapTimeInt{v: v, nullable: k == reflect.Ptr}, true
}

func init() {
	driver.RegisterDriver("sqlite3", func(p map[string]string) driver.Driver {
		var timeAs = TimeAsTime
		if t, ok := p["time"]; ok {
			switch t {
			case TimeAsString, TimeAsInt:
				timeAs = t
			}
		}
		return drv{
			timeAs: timeAs,
			Stub:   driver.Stub{QuoteFunc: quote},
		}
	})
}
