package sdm

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Ronmi/sdm/driver"
)

func initBenchDB(b *testing.B) (*sql.DB, *Manager) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Cannot open sqlite connection: %s", err)
	}
	m := New(db, "stub")

	return db, m
}

func BenchmarkRegister(b *testing.B) {
	var t testok
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		_, m := initBenchDB(b)
		b.StartTimer()
		m.Register(t, "testok")
	}
}

func init() {
	driver.RegisterStub(func(n string) string {
		return "'" + n + "'"
	})
}

func BenchmarkSQLBuilder(b *testing.B) {
	_, m := initBenchDB(b)
	t := testok{
		Perper:       1,
		nonExportInt: 2,
		ExportInt:    3,
		ExportTime:   time.Now(),
		ExportString: "benchmark",
	}
	m.Register(t, "testok")

	simpleCases := []struct {
		n string
		f func(interface{})
	}{
		{"Col_SELECT", func(data interface{}) { m.Col(data, driver.QSelect) }},
		{"Col_WHERE", func(data interface{}) { m.Col(data, driver.QWhere) }},
		{"Col_INSERT", func(data interface{}) { m.Col(data, driver.QInsert) }},
		{"Col_UPDATE", func(data interface{}) { m.Col(data, driver.QUpdate) }},
		{"ColIns", func(data interface{}) { m.ColIns(data) }},
		{"Val", func(data interface{}) { m.Val(data) }},
		{"ValIns", func(data interface{}) { m.ValIns(data) }},
		{"Holder", func(data interface{}) { m.Holder(data) }},
		{"HolderIns", func(data interface{}) { m.HolderIns(data) }},
		{"MakeInsert", func(data interface{}) { m.makeInsert(data) }},
		{"MakeUpdate", func(data interface{}) { m.makeUpdate(data, "", nil) }},
		{"MakeDelete", func(data interface{}) { m.makeDelete(data) }},
	}

	for _, c := range simpleCases {
		b.Run(c.n, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				c.f(t)
			}
		})
	}
}
