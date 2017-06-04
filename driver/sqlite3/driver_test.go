package sqlite3

import (
	"reflect"
	"testing"
	"time"

	"git.ronmi.tw/ronmi/sdm"
)

func TestColumnSQL(t *testing.T) {
	cases := []struct {
		t    reflect.Type
		col  []sdm.ColumnDef
		idx  []sdm.IndexDef
		qstr string
		msg  string
	}{
		{
			t: reflect.TypeOf(struct {
				ID    int    `sdm:"id,ai,pri_test_pk"`
				Year  int    `sdm:"y,uniq_birth"`
				Month int    `sdm:"m,uniq_birth"`
				Date  int    `sdm:"d,uniq_birth"`
				Name  string `sdm:"name"`
			}{}),
			col: []sdm.ColumnDef{
				{ID: 0, AI: true, Name: "id"},
				{ID: 1, AI: false, Name: "y"},
				{ID: 2, AI: false, Name: "m"},
				{ID: 3, AI: false, Name: "d"},
				{ID: 4, AI: false, Name: "name"},
			},
			idx: []sdm.IndexDef{
				{Type: sdm.IndexTypePrimary, Name: "test_pk", Cols: []string{"id"}},
				{Type: sdm.IndexTypeUnique, Name: "birth", Cols: []string{"y", "m", "d"}},
			},
			qstr: `'id' INT CONSTRAINT 'test_pk' PRIMARY KEY AUTOINCREMENT,'y' INT,'m' INT,'d' INT,'name' TEXT,CONSTRAINT 'birth' UNIQUE ('y','m','d')`,
			msg:  "well-defined struct with auto increment, primary and unique key",
		},
		{
			t: reflect.TypeOf(struct {
				ID    int    `sdm:"id,pri_test_pk"`
				Year  int    `sdm:"y,uniq_birth"`
				Month int    `sdm:"m,uniq_birth"`
				Date  int    `sdm:"d,uniq_birth"`
				Name  string `sdm:"name"`
			}{}),
			col: []sdm.ColumnDef{
				{ID: 0, AI: false, Name: "id"},
				{ID: 1, AI: false, Name: "y"},
				{ID: 2, AI: false, Name: "m"},
				{ID: 3, AI: false, Name: "d"},
				{ID: 4, AI: false, Name: "name"},
			},
			idx: []sdm.IndexDef{
				{Type: sdm.IndexTypePrimary, Name: "test_pk", Cols: []string{"id"}},
				{Type: sdm.IndexTypeUnique, Name: "birth", Cols: []string{"y", "m", "d"}},
			},
			qstr: `'id' INT,'y' INT,'m' INT,'d' INT,'name' TEXT,CONSTRAINT 'test_pk' PRIMARY KEY ('id'),CONSTRAINT 'birth' UNIQUE ('y','m','d')`,
			msg:  "well-defined struct with primary and unique key",
		},
		{
			t: reflect.TypeOf(struct {
				ID    int    `sdm:"id,ai"`
				Year  int    `sdm:"y,uniq_birth"`
				Month int    `sdm:"m,uniq_birth"`
				Date  int    `sdm:"d,uniq_birth"`
				Name  string `sdm:"name"`
			}{}),
			col: []sdm.ColumnDef{
				{ID: 0, AI: true, Name: "id"},
				{ID: 1, AI: false, Name: "y"},
				{ID: 2, AI: false, Name: "m"},
				{ID: 3, AI: false, Name: "d"},
				{ID: 4, AI: false, Name: "name"},
			},
			idx: []sdm.IndexDef{
				{Type: sdm.IndexTypeUnique, Name: "birth", Cols: []string{"y", "m", "d"}},
			},
			qstr: `'id' INT CONSTRAINT '_pk' PRIMARY KEY AUTOINCREMENT,'y' INT,'m' INT,'d' INT,'name' TEXT,CONSTRAINT 'birth' UNIQUE ('y','m','d')`,
			msg:  "well-defined struct with auto increment and unique key",
		},
		{
			t: reflect.TypeOf(struct {
				ID   int       `sdm:"id"`
				Col1 float64   `sdm:"c1"`
				Col2 bool      `sdm:"c2"`
				Col3 time.Time `sdm:"c3"`
				Col4 string    `sdm:"c4"`
				Col5 []byte    `sdm:"c5"`
			}{}),
			col: []sdm.ColumnDef{
				{ID: 0, AI: false, Name: "id"},
				{ID: 1, AI: false, Name: "c1"},
				{ID: 2, AI: false, Name: "c2"},
				{ID: 3, AI: false, Name: "c3"},
				{ID: 4, AI: false, Name: "c4"},
				{ID: 5, AI: false, Name: "c5"},
			},
			idx:  []sdm.IndexDef{},
			qstr: `'id' INT,'c1' REAL,'c2' INT,'c3' DATETIME,'c4' TEXT,'c5' BLOB`,
			msg:  "well-defined struct with every type, but no key",
		},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			actual := createTableColumnSQL(c.t, c.col, c.idx)
			if actual != c.qstr {
				t.Errorf("dumping\nexpect: %s\nactual: %s", c.qstr, actual)
			}
		})
	}
}
