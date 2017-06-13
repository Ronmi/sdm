package sdm

import (
	sqlDriver "database/sql/driver"
	"testing"
	"time"
)

func TestBulk(t *testing.T) {
	_, m := initdb(t)

	t.Run("Insert", func(t *testing.T) {
		ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-07-07 08:00:00 +0800")
		d1 := testai{1, "bulk insert", ti}
		d2 := testai{2, "bulk insert", ti}

		b, err := m.BulkInsert(d1)
		if err != nil {
			t.Fatalf("Cannot create bulk insert object: %s", err)
		}

		b.Add(d1)
		b.Add(d2)

		expectStr := `INSERT INTO 'testai' ('estr','t') VALUES (?,?),(?,?)`
		expectVal := []interface{}{
			d1.ExportString, d1.ExportTime.Unix(),
			d2.ExportString, d2.ExportTime.Unix(),
		}
		qstr, vals := b.Make()
		if qstr != expectStr {
			t.Errorf("Expect bulk insert generates [%s], got [%s]", expectStr, qstr)
		}
		if exp, act := len(expectVal), len(vals); exp != act {
			t.Fatalf("Expected to get %d vals, get %d", exp, act)
		}

		for idx, exp := range expectVal {
			i := vals[idx]
			if i == exp {
				continue
			}

			act, ok := i.(sqlDriver.Valuer)
			if !ok {
				t.Fatalf("Expected to get %v, got %v", exp, act)
			}

			iface, err := act.Value()
			if err != nil {
				t.Fatalf("Error calling valuer: %s", err)
			}

			if iface != exp {
				t.Fatalf("Expect %v, got %v", exp, iface)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-07-07 08:00:00 +0800")
		d1 := testai{1, "bulk insert", ti}
		d2 := testai{2, "bulk insert", ti}

		b, err := m.BulkDelete(d1)
		if err != nil {
			t.Fatalf("Cannot create bulk delete object: %s", err)
		}

		b.Add(d1)
		b.Add(d2)

		expectStr := `DELETE FROM 'testai' WHERE ('testai'.'eint'=? AND 'testai'.'estr'=? AND 'testai'.'t'=?) OR ('testai'.'eint'=? AND 'testai'.'estr'=? AND 'testai'.'t'=?)`
		expectVal := []interface{}{
			d1.ExportInt, d1.ExportString, d1.ExportTime.Unix(),
			d2.ExportInt, d2.ExportString, d2.ExportTime.Unix(),
		}
		qstr, vals := b.Make()
		if qstr != expectStr {
			t.Errorf("Expect bulk delete generates [%s], got [%s]", expectStr, qstr)
		}

		if exp, act := len(expectVal), len(vals); exp != act {
			t.Fatalf("Expected to get %d vals, get %d", exp, act)
		}

		for idx, exp := range expectVal {
			i := vals[idx]
			if i == exp {
				continue
			}

			act, ok := i.(sqlDriver.Valuer)
			if !ok {
				t.Fatalf("Expected to get %v, got %v", exp, act)
			}

			iface, err := act.Value()
			if err != nil {
				t.Fatalf("Error calling valuer: %s", err)
			}

			if iface != exp {
				t.Fatalf("Expect %v, got %v", exp, iface)
			}
		}
	})
}
