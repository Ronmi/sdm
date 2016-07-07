package sdm

import (
	"reflect"
	"testing"
	"time"
)

func TestBulkInsert(t *testing.T) {
	ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-07-07 08:00:00 +0800")
	d1 := testai{1, "bulk insert", ti}
	d2 := testai{2, "bulk insert", ti}

	b, err := m.BulkInsert("testai", d1)
	if err != nil {
		t.Fatalf("Cannot create bulk insert object: %s", err)
	}

	b.Add(d1)
	b.Add(d2)

	expectStr := `INSERT INTO testai (estr,t) VALUES (?,?),(?,?)`
	expectVal := []interface{}{
		d1.ExportString, d1.ExportTime,
		d2.ExportString, d2.ExportTime,
	}
	qstr, vals := b.Make()
	if qstr != expectStr {
		t.Errorf("Expect bulk insert generates [%s], got [%s]", expectStr, qstr)
	}
	if !reflect.DeepEqual(vals, expectVal) {
		t.Errorf("Bulk insert fills in wrong order: %#v", vals)
	}
}

func TestBulkDelete(t *testing.T) {
	ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-07-07 08:00:00 +0800")
	d1 := testai{1, "bulk insert", ti}
	d2 := testai{2, "bulk insert", ti}

	b, err := m.BulkDelete("testai", d1)
	if err != nil {
		t.Fatalf("Cannot create bulk delete object: %s", err)
	}

	b.Add(d1)
	b.Add(d2)

	expectStr := `DELETE FROM testai WHERE (eint=? AND estr=? AND t=?) OR (eint=? AND estr=? AND t=?)`
	expectVal := []interface{}{
		d1.ExportInt, d1.ExportString, d1.ExportTime,
		d2.ExportInt, d2.ExportString, d2.ExportTime,
	}
	qstr, vals := b.Make()
	if qstr != expectStr {
		t.Errorf("Expect bulk delete generates [%s], got [%s]", expectStr, qstr)
	}
	if !reflect.DeepEqual(vals, expectVal) {
		t.Errorf("Bulk delete fills in wrong order: %#v", vals)
	}
}
