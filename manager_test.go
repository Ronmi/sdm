package sdm

import (
	"database/sql"
	"log"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type testok struct {
	Perper       int
	nonExportInt int       `sdm:"nint"`
	ExportInt    int       `sdm:"eint"`
	ExportString string    `sdm:"estr"`
	ExportTime   time.Time `sdm:"t"`
}

var db *sql.DB

func init() {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Cannot open sqlite connection: %s", err)
	}
	db = conn

	s := `CREATE TABLE testok (eint int, estr varchar(10), t datetime)`
	if _, err := db.Exec(s); err != nil {
		log.Fatalf("Cannot create table testok: %s", err)
	}

	s = `INSERT INTO testok (eint, estr, t) VALUES (10, "test", "2016-05-03 00:00:00")`
	if _, err := db.Exec(s); err != nil {
		log.Fatalf("Cannot insert preset data into testok: %s", err)
	}
}

func TestScanOK(t *testing.T) {
	m := New()
	var val testok

	rows, err := db.Query(`SELECT * FROM testok`)
	if err != nil {
		t.Fatalf("Cannot query select to testok: %s", err)
	}

	proxy := m.Proxify(rows, val)
	if proxy.Err() != nil {
		t.Fatalf("Cannot proxy the sql.Rows with value: %s", err)
	}
	proxy = m.Proxify(rows, &val)
	if proxy.Err() != nil {
		t.Fatalf("Cannot proxy the sql.Rows with pointer: %s", err)
	}
	defer proxy.Close()
	proxy.Next()
	if err := proxy.Scan(&val); err != nil {
		t.Fatalf("Cannot scan: %s", err)
	}

	if val.ExportInt != 10 {
		t.Errorf("ExportInt != 10: %d", val.ExportInt)
	}

	if val.ExportString != "test" {
		t.Errorf("ExportString != test: %s", val.ExportString)
	}

	if val.ExportTime.Unix() != 1462233600 {
		t.Errorf("ExportTime != 2016-05-03 00:00:00: %s", val.ExportTime.UTC().String())
	}

	if val.Perper != 0 {
		t.Errorf("Perper != 0: %d", val.Perper)
	}

	if val.nonExportInt != 0 {
		t.Errorf("nonExportInt != 0: %d", val.nonExportInt)
	}
}

func TestInsert(t *testing.T) {
	m := New()
	ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-05-04 08:00:00 +0800")
	data := testok{1, 2, 3, "insert", ti}

	if _, err := m.Insert(db, "testok", data); err != nil {
		t.Fatalf("Error inserting data: %s", err)
	}

	var cnt int
	row := db.QueryRow(`SELECT COUNT(eint) FROM testok WHERE eint=3 AND estr="insert" AND strftime("%s", t)="1462320000"`)
	if err := row.Scan(&cnt); err != nil {
		t.Fatalf("Cannot scan COUNT(eint) for insert: %s", err)
	}
	if cnt != 1 {
		t.Errorf("There should be only one result after inserting, but we got %d", cnt)
	}
}

func TestUpdate(t *testing.T) {
	m := New()
	ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-05-04 08:00:00 +0800")
	data := testok{1, 2, 3, "update", ti}

	if _, err := m.Insert(db, "testok", data); err != nil {
		t.Fatalf("Error inserting data for updating: %s", err)
	}

	data.ExportInt = 4
	if _, err := m.Update(db, "testok", data, `eint=? AND estr=? AND strftime("%s", t)="1462320000"`, 3, "update"); err != nil {
		t.Fatalf("Error updating data: %s", err)
	}

	var cnt int
	row := db.QueryRow(`SELECT COUNT(eint) FROM testok WHERE eint=4 AND estr="update" AND strftime("%s", t)="1462320000"`)
	if err := row.Scan(&cnt); err != nil {
		t.Fatalf("Cannot scan COUNT(eint) for update: %s", err)
	}
	if cnt != 1 {
		t.Errorf("There should be only one result after updating, but we got %d", cnt)
	}
}
