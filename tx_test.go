package sdm

import (
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestTxCommit(t *testing.T) {
	db, m := initdb(t)
	ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-07-05 08:00:00 +0800")
	m.Insert(testok{2, 3, 4, "predefined", ti})
	data := testok{1, 2, 3, "insert", ti}

	tx, err := m.Begin()
	if err != nil {
		t.Fatalf("Cannot create transaction for commit: %s", err)
	}

	if _, err = tx.Insert(data); err != nil {
		t.Fatalf("Error inserting data for commit: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Error commit: %s", err)
	}

	var cnt int
	row := db.QueryRow(`SELECT COUNT(eint) FROM testok WHERE eint=3 AND estr="insert" AND t=1467676800`)
	if err := row.Scan(&cnt); err != nil {
		t.Fatalf("Cannot scan COUNT(eint) for commit: %s", err)
	}
	if cnt != 1 {
		t.Errorf("There should be only one result after commit, but we got %d", cnt)
	}
}

func TestTxRollback(t *testing.T) {
	db, m := initdb(t)
	ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-07-05 08:00:00 +0800")
	m.Insert(testok{2, 3, 4, "predefined", ti})
	data := testok{1, 2, 3, "update", ti}

	tx, err := m.Begin()
	if err != nil {
		t.Fatalf("Cannot create transaction for rollback: %s", err)
	}

	if _, err := tx.Insert(data); err != nil {
		t.Fatalf("Error inserting data for rollback: %s", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Error rollback: %s", err)
	}

	var cnt int
	row := db.QueryRow(`SELECT COUNT(eint) FROM testok WHERE eint=4 AND estr="update" AND t=1467676800`)
	if err := row.Scan(&cnt); err != nil {
		t.Fatalf("Cannot scan COUNT(eint) for rollback: %s", err)
	}
	if cnt != 0 {
		t.Errorf("There should be no result after rollback, but we got %d", cnt)
	}
}
