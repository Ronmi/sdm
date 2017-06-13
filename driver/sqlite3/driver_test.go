package sqlite3

import (
	"reflect"
	"testing"
	"time"

	"git.ronmi.tw/ronmi/sdm/driver"
)

type testSQLCase struct {
	timeAs string
	t      reflect.Type
	col    []driver.Column
	idx    []driver.Index
	qstr   string
	msg    string
}

func TestColumnSQL(t *testing.T) {
	cases := []testSQLCase{
		{
			timeAs: TimeAsInt,
			t: reflect.TypeOf(struct {
				ID int
				T  time.Time
			}{}),
			col: []driver.Column{
				{ID: 0, AI: false, Name: "id"},
				{ID: 1, AI: false, Name: "t"},
			},
			idx:  []driver.Index{},
			qstr: `'id' INTEGER NOT NULL,'t' INTEGER NOT NULL`,
			msg:  `time as integer (timestamp)`,
		},
		{
			timeAs: TimeAsString,
			t: reflect.TypeOf(struct {
				ID int
				T  time.Time
			}{}),
			col: []driver.Column{
				{ID: 0, AI: false, Name: "id"},
				{ID: 1, AI: false, Name: "t"},
			},
			idx:  []driver.Index{},
			qstr: `'id' INTEGER NOT NULL,'t' TEXT NOT NULL`,
			msg:  `time as string (formatted string)`,
		},
		{
			timeAs: TimeAsTime,
			t: reflect.TypeOf(struct {
				ID    int    `driver:"id,ai,pri_test_pk"`
				Year  int    `driver:"y,uniq_birth"`
				Month int    `driver:"m,uniq_birth"`
				Date  int    `driver:"d,uniq_birth"`
				Name  string `driver:"name"`
			}{}),
			col: []driver.Column{
				{ID: 0, AI: true, Name: "id"},
				{ID: 1, AI: false, Name: "y"},
				{ID: 2, AI: false, Name: "m"},
				{ID: 3, AI: false, Name: "d"},
				{ID: 4, AI: false, Name: "name"},
			},
			idx: []driver.Index{
				{Type: driver.IndexTypePrimary, Name: "test_pk", Cols: []string{"id"}},
				{Type: driver.IndexTypeUnique, Name: "birth", Cols: []string{"y", "m", "d"}},
			},
			qstr: `'id' INTEGER NOT NULL CONSTRAINT 'test_pk' PRIMARY KEY AUTOINCREMENT,'y' INTEGER NOT NULL,'m' INTEGER NOT NULL,'d' INTEGER NOT NULL,'name' TEXT NOT NULL,CONSTRAINT 'birth' UNIQUE ('y','m','d')`,
			msg:  "well-defined struct with auto increment, primary and unique key",
		},
		{
			timeAs: TimeAsTime,
			t: reflect.TypeOf(struct {
				ID    int    `driver:"id,pri_test_pk"`
				Year  int    `driver:"y,uniq_birth"`
				Month int    `driver:"m,uniq_birth"`
				Date  int    `driver:"d,uniq_birth"`
				Name  string `driver:"name"`
			}{}),
			col: []driver.Column{
				{ID: 0, AI: false, Name: "id"},
				{ID: 1, AI: false, Name: "y"},
				{ID: 2, AI: false, Name: "m"},
				{ID: 3, AI: false, Name: "d"},
				{ID: 4, AI: false, Name: "name"},
			},
			idx: []driver.Index{
				{Type: driver.IndexTypePrimary, Name: "test_pk", Cols: []string{"id"}},
				{Type: driver.IndexTypeUnique, Name: "birth", Cols: []string{"y", "m", "d"}},
			},
			qstr: `'id' INTEGER NOT NULL,'y' INTEGER NOT NULL,'m' INTEGER NOT NULL,'d' INTEGER NOT NULL,'name' TEXT NOT NULL,CONSTRAINT 'test_pk' PRIMARY KEY ('id'),CONSTRAINT 'birth' UNIQUE ('y','m','d')`,
			msg:  "well-defined struct with primary and unique key",
		},
		{
			timeAs: TimeAsTime,
			t: reflect.TypeOf(struct {
				ID    int    `driver:"id,ai"`
				Year  int    `driver:"y,uniq_birth"`
				Month int    `driver:"m,uniq_birth"`
				Date  int    `driver:"d,uniq_birth"`
				Name  string `driver:"name"`
			}{}),
			col: []driver.Column{
				{ID: 0, AI: true, Name: "id"},
				{ID: 1, AI: false, Name: "y"},
				{ID: 2, AI: false, Name: "m"},
				{ID: 3, AI: false, Name: "d"},
				{ID: 4, AI: false, Name: "name"},
			},
			idx: []driver.Index{
				{Type: driver.IndexTypeUnique, Name: "birth", Cols: []string{"y", "m", "d"}},
			},
			qstr: `'id' INTEGER NOT NULL CONSTRAINT '_pk' PRIMARY KEY AUTOINCREMENT,'y' INTEGER NOT NULL,'m' INTEGER NOT NULL,'d' INTEGER NOT NULL,'name' TEXT NOT NULL,CONSTRAINT 'birth' UNIQUE ('y','m','d')`,
			msg:  "well-defined struct with auto increment and unique key",
		},
		{
			timeAs: TimeAsTime,
			t: reflect.TypeOf(struct {
				ID   int       `driver:"id"`
				Col1 float64   `driver:"c1"`
				Col2 bool      `driver:"c2"`
				Col3 time.Time `driver:"c3"`
				Col4 string    `driver:"c4"`
				Col5 []byte    `driver:"c5"`
			}{}),
			col: []driver.Column{
				{ID: 0, AI: false, Name: "id"},
				{ID: 1, AI: false, Name: "c1"},
				{ID: 2, AI: false, Name: "c2"},
				{ID: 3, AI: false, Name: "c3"},
				{ID: 4, AI: false, Name: "c4"},
				{ID: 5, AI: false, Name: "c5"},
			},
			idx:  []driver.Index{},
			qstr: `'id' INTEGER NOT NULL,'c1' REAL NOT NULL,'c2' INTEGER NOT NULL,'c3' DATETIME NOT NULL,'c4' TEXT NOT NULL,'c5' BLOB`,
			msg:  "well-defined struct with every type, but no key",
		},
		{
			timeAs: TimeAsTime,
			t: reflect.TypeOf(struct {
				ID    int    `driver:"id,ai"`
				Year  int    `driver:"y,idx_birth"`
				Month int    `driver:"m,idx_birth"`
				Date  int    `driver:"d,idx_birth"`
				Name  string `driver:"name"`
			}{}),
			col: []driver.Column{
				{ID: 0, AI: true, Name: "id"},
				{ID: 1, AI: false, Name: "y"},
				{ID: 2, AI: false, Name: "m"},
				{ID: 3, AI: false, Name: "d"},
				{ID: 4, AI: false, Name: "name"},
			},
			idx: []driver.Index{
				{Type: driver.IndexTypeIndex, Name: "birth", Cols: []string{"y", "m", "d"}},
			},
			qstr: `'id' INTEGER NOT NULL CONSTRAINT '_pk' PRIMARY KEY AUTOINCREMENT,'y' INTEGER NOT NULL,'m' INTEGER NOT NULL,'d' INTEGER NOT NULL,'name' TEXT NOT NULL`,
			msg:  "well-defined struct with auto increment and index key",
		},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			actual := createTableColumnSQL(c.t, c.col, c.idx, c.timeAs)
			if actual != c.qstr {
				t.Errorf("dumping\nexpect: %s\nactual: %s", c.qstr, actual)
			}
		})
	}
}
