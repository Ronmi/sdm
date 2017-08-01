package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
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

func quote(name string) string {
	return "`" + name + "`"
}

type drv struct {
	charset       string
	collate       string
	stringKeySize string
	blobKeySize   string
	driver.Stub
}

func (d *drv) getType(typ reflect.Type, name string, indexes []driver.Index) string {
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
			has := false
			for _, i := range indexes {
				if i.HasCol(name) {
					has = true
					break
				}
			}
			if has {
				def = "VARCHAR(" + d.stringKeySize + ")"
			}
			def += " CHARACTER SET " + d.charset + " COLLATE " + d.collate
		}
		return def + postfix
	}

	if driver.IsTime(t) {
		return "TIMESTAMP" + postfix
	}

	panic("sdm: driver: mysql: unsupported type " + t.String())
}

func (d *drv) createTableColumnSQL(typ reflect.Type, cols []driver.Column, indexes []driver.Index) string {
	ret := make([]string, 0, len(cols)+len(indexes))

	hasAI := false

	for _, c := range cols {
		def := quote(c.Name) + ` ` + d.getType(typ.Field(c.ID).Type, c.Name, indexes)
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
			for _, c := range cols {
				if c.Name != v {
					continue
				}

				ki := typ.Field(c.ID).Type.Kind()
				kie := ki
				if ki == reflect.Array || ki == reflect.Slice || ki == reflect.Ptr {
					kie = typ.Field(c.ID).Type.Elem().Kind()
				}
				if ki == reflect.Array || ki == reflect.Slice {
					if kie == reflect.Uint8 {
						quoted[k] += "(" + d.blobKeySize + ")"
					}
				}
				break
			}
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

	return strings.Join(ret, ",")
}

func (d drv) CreateTable(db *sql.DB, name string, typ reflect.Type, cols []driver.Column, indexes []driver.Index) (sql.Result, error) {
	qstr := fmt.Sprintf(
		"CREATE TABLE %s (%s) DEFAULT CHARACTER SET %s,DEFAULT COLLATE %s",
		quote(name),
		d.createTableColumnSQL(typ, cols, indexes),
		d.charset,
		d.collate,
	)

	return db.Exec(qstr)
}

func (d drv) CreateTableNotExist(db *sql.DB, name string, typ reflect.Type, cols []driver.Column, indexes []driver.Index) (sql.Result, error) {
	qstr := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s) DEFAULT CHARACTER SET %s,DEFAULT COLLATE %s",
		quote(name),
		d.createTableColumnSQL(typ, cols, indexes),
		d.charset,
		d.collate,
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
		sSize := 256
		bSize := 2048

		if c, ok := p["charset"]; ok {
			charset = c
		}
		if c, ok := p["collate"]; ok {
			collate = c
		}
		if c, ok := p["stringSize"]; ok {
			if sz, err := strconv.Atoi(c); err == nil && sz <= sSize && sz > 0 {
				sSize = sz
			}
		}
		if c, ok := p["blobSize"]; ok {
			if sz, err := strconv.Atoi(c); err == nil {
				bSize = sz
			}
		}

		return drv{
			charset:       charset,
			collate:       collate,
			stringKeySize: strconv.Itoa(sSize),
			blobKeySize:   strconv.Itoa(bSize),
			Stub:          driver.Stub{QuoteFunc: quote},
		}
	})
}
