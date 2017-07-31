package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/Ronmi/sdm/driver"
)

var typeMap = map[reflect.Kind]string{
	reflect.Int:     "BIGINT",
	reflect.Int8:    "TINYINT",
	reflect.Int16:   "SMALLINT",
	reflect.Int32:   "INT",
	reflect.Int64:   "BIGINT",
	reflect.Uint:    "BIGINT UNSIGNED",
	reflect.Uint8:   "TINYINT UNSIGNED",
	reflect.Uint16:  "SMALLINT UNSIGNED",
	reflect.Uint32:  "INT UNSIGNED",
	reflect.Uint64:  "BIGINT UNSIGNED",
	reflect.Float32: "FLOAT",
	reflect.Float64: "DOUBLE",
	reflect.Bool:    "BIT(1)",
	reflect.String:  "TEXT",
}

func getType(typ reflect.Type, charset, collate string) string {
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

	if def, ok := typeMap[t.Kind()]; ok {
		if t.Kind() == reflect.String {
			def += " CHARACTER SET " + charset + " COLLATE " + collate
		}
		return def + postfix
	}

	if driver.IsTime(t) {
		return "TIMESTAMP" + postfix
	}

	panic("sdm: driver: mysql: unsupported type " + t.String())
}

func quote(name string) string {
	return "`" + name + "`"
}

func createTableColumnSQL(typ reflect.Type, cols []driver.Column, indexes []driver.Index, charset, collate string) string {
	ret := make([]string, 0, len(cols)+len(indexes))

	hasAI := false

	for _, c := range cols {
		def := quote(c.Name) + ` ` + getType(typ.Field(c.ID).Type, charset, collate)
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
			def += " CONSTRAINT " + quote(name) + " PRIMARY KEY AUTO_INCREMENT"
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
			def = fmt.Sprintf(
				"INDEX %s (%s)",
				quote(i.Name),
				strings.Join(quoted, ","),
			)
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
				"CONSTRAINT %s UNIQUE KEY (%s)",
				quote(i.Name),
				strings.Join(quoted, ","),
			)
		}

		ret = append(ret, def)
	}

	return strings.Join(ret, ",") + " DEFAULT CHARACTER SET " + charset + " DEFAULT COLLATE " + collate
}

type drv struct {
	charset string
	collate string
	driver.Stub
}

func (d drv) CreateTable(db *sql.DB, name string, typ reflect.Type, cols []driver.Column, indexes []driver.Index) (sql.Result, error) {
	qstr := fmt.Sprintf(
		"CREATE TABLE '%s' (%s)",
		name,
		createTableColumnSQL(typ, cols, indexes, d.charset, d.collate),
	)

	return db.Exec(qstr)
}

func (d drv) CreateTableNotExist(db *sql.DB, name string, typ reflect.Type, cols []driver.Column, indexes []driver.Index) (sql.Result, error) {
	qstr := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS '%s' (%s)",
		name,
		createTableColumnSQL(typ, cols, indexes, d.charset, d.collate),
	)

	return db.Exec(qstr)
}

func (d drv) Col(table, col string, kind driver.QuotingType) string {
	return quote(table) + "." + quote(col)
}

func init() {
	driver.RegisterDriver("mysql", func(p map[string]string) driver.Driver {
		charset := "utf8"
		collate := "utf8_general_ci"

		if c, ok := p["charset"]; ok {
			charset = c
		}
		if c, ok := p["collate"]; ok {
			collate = c
		}

		return drv{
			charset: charset,
			collate: collate,
			Stub:    driver.Stub{QuoteFunc: quote},
		}
	})
}
