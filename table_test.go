package sdm

import "testing"

type testTable1 struct {
	A int `sdm:"a"`
}

type testTable2 struct {
	A int `sdm:"a"`
}

type testTable3 struct {
	A int `sdm:"a"`
}

func TestCreateTables(t *testing.T) {
	db := newdb()
	m := New(db, "sqlite3")
	m.Reg(testTable1{}, testTable2{}, testTable3{})

	if err := m.CreateTables(); err != nil {
		t.Fatalf("error occured when trying to create all table: %s", err)
	}

	if _, err := m.Insert(testTable1{0}); err != nil {
		t.Errorf("failed to insert into table 1: %s", err)
	}
	if _, err := m.Insert(testTable2{0}); err != nil {
		t.Errorf("failed to insert into table 2: %s", err)
	}
	if _, err := m.Insert(testTable3{0}); err != nil {
		t.Errorf("failed to insert into table 3: %s", err)
	}
}

func TestCreateTablesNotExist(t *testing.T) {
	db := newdb()
	m := New(db, "sqlite3")
	m.Register(testTable1{}, "t1")
	m.Register(testTable2{}, "t2")

	if err := m.CreateTables(); err != nil {
		t.Fatalf("error occured when trying to create all table: %s", err)
	}

	if _, err := m.Insert(testTable1{0}); err != nil {
		t.Fatalf("failed to insert into table 1: %s", err)
	}
	if _, err := m.Insert(testTable2{0}); err != nil {
		t.Fatalf("failed to insert into table 2: %s", err)
	}

	m.Register(testTable3{}, "t3")
	if err := m.CreateTablesNotExist(); err != nil {
		t.Fatalf("error occured when trying to create all table: %s", err)
	}
	if _, err := m.Insert(testTable3{0}); err != nil {
		t.Fatalf("failed to insert into table 3: %s", err)
	}
}
