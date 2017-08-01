package mysql

import (
	"reflect"
	"testing"
	"time"

	"github.com/Ronmi/sdm/driver"
)

type testSQLCase struct {
	t    reflect.Type
	col  []driver.Column
	idx  []driver.Index
	qstr string
	msg  string
}

func TestColumnSQL(t *testing.T) {
	cases := []testSQLCase{
		{
			t: reflect.TypeOf(struct {
				ID int
				T  time.Time
			}{}),
			col: []driver.Column{
				{ID: 0, AI: false, Name: "id"},
				{ID: 1, AI: false, Name: "t"},
			},
			idx:  []driver.Index{},
			qstr: "`id` BIGINT NOT NULL,`t` TIMESTAMP NOT NULL",
			msg:  "simple table",
		},
		{
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
			qstr: "`id` BIGINT NOT NULL AUTO_INCREMENT,`y` BIGINT NOT NULL,`m` BIGINT NOT NULL,`d` BIGINT NOT NULL,`name` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,CONSTRAINT `test_pk` PRIMARY KEY (`id`),CONSTRAINT `birth` UNIQUE KEY (`y`,`m`,`d`)",
			msg:  "well-defined struct with auto increment, primary and unique key",
		},
		{
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
			qstr: "`id` BIGINT NOT NULL,`y` BIGINT NOT NULL,`m` BIGINT NOT NULL,`d` BIGINT NOT NULL,`name` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,CONSTRAINT `test_pk` PRIMARY KEY (`id`),CONSTRAINT `birth` UNIQUE KEY (`y`,`m`,`d`)",
			msg:  "well-defined struct with primary and unique key",
		},
		{
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
			qstr: "`id` BIGINT NOT NULL AUTO_INCREMENT,`y` BIGINT NOT NULL,`m` BIGINT NOT NULL,`d` BIGINT NOT NULL,`name` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,CONSTRAINT `_pk` PRIMARY KEY (`id`),CONSTRAINT `birth` UNIQUE KEY (`y`,`m`,`d`)",
			msg:  "well-defined struct with auto increment and unique key",
		},
		{
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
			qstr: "`id` BIGINT NOT NULL,`c1` DOUBLE NOT NULL,`c2` BIT(1) NOT NULL,`c3` TIMESTAMP NOT NULL,`c4` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,`c5` BLOB",
			msg:  "well-defined struct with every type, but no key",
		},
		{
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
			qstr: "`id` BIGINT NOT NULL AUTO_INCREMENT,`y` BIGINT NOT NULL,`m` BIGINT NOT NULL,`d` BIGINT NOT NULL,`name` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,CONSTRAINT `_pk` PRIMARY KEY (`id`),INDEX `birth` (`y`,`m`,`d`)",
			msg:  "well-defined struct with auto increment and index key",
		},
		{
			t: reflect.TypeOf(struct {
				ID    int    `driver:"id,ai"`
				Year  int    `driver:"y"`
				Month int    `driver:"m"`
				Date  int    `driver:"d"`
				Name  string `driver:"name,idx_myname"`
			}{}),
			col: []driver.Column{
				{ID: 0, AI: true, Name: "id"},
				{ID: 1, AI: false, Name: "y"},
				{ID: 2, AI: false, Name: "m"},
				{ID: 3, AI: false, Name: "d"},
				{ID: 4, AI: false, Name: "name"},
			},
			idx: []driver.Index{
				{Type: driver.IndexTypeIndex, Name: "myname", Cols: []string{"name"}},
			},
			qstr: "`id` BIGINT NOT NULL AUTO_INCREMENT,`y` BIGINT NOT NULL,`m` BIGINT NOT NULL,`d` BIGINT NOT NULL,`name` VARCHAR(256) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,CONSTRAINT `_pk` PRIMARY KEY (`id`),INDEX `myname` (`name`)",
			msg:  "well-defined struct with auto increment and index key on string column",
		},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			d := &drv{
				stringKeySize: "256",
				blobKeySize:   "2048",
				charset:       "utf8",
				collate:       "utf8_general_ci",
				Stub:          driver.Stub{QuoteFunc: quote},
			}
			actual := d.createTableColumnSQL(c.t, c.col, c.idx)
			expect := c.qstr
			if actual != expect {
				t.Errorf("dumping\nexpect: %s\nactual: %s", expect, actual)
			}
		})
	}
}
