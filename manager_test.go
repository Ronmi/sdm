package sdm

import (
	"database/sql"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/Ronmi/sdm/driver"
	_ "github.com/Ronmi/sdm/driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type testok struct {
	Perper       int
	nonExportInt int       `sdm:"nint,idx_a"`
	ExportInt    int       `sdm:"eint,idx_a,uniq_b"`
	ExportString string    `sdm:"estr,idx_a"`
	ExportTime   time.Time `sdm:"t,idx_c"`
}

type testai struct {
	ExportInt    int       `sdm:"eint,ai"`
	ExportString string    `sdm:"estr"`
	ExportTime   time.Time `sdm:"t"`
}

func newdb(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Cannot open sqlite connection: %s", err)
	}
	return db
}

func initdb(t *testing.T) (*sql.DB, *Manager) {
	db := newdb(t)
	m := New(db, "sqlite3:time=int")

	m.Reg(testok{}, testai{})
	m.CreateTablesNotExist()

	return db, m
}

func initmysqldb(t *testing.T) (*sql.DB, *Manager) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Fatal("You must provide MYSQL_DSN environment variable to run test")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("Cannot open mysql connection: %s", err)
	}
	m := New(db, "mysql")

	m.Reg(testok{}, testai{})
	if err := m.CreateTablesNotExist(); err != nil {
		t.Fatalf("Cannot create table: %s", err)
	}

	return db, m
}

func teardownsqlite(t *testing.T, db *sql.DB) {
}

func teardownmysql(t *testing.T, db *sql.DB) {
	db.Exec("DROP TABLE IF EXISTS `testok`, `testai`")
}
func ts2sqlite(ts int) string {
	return `"` + strconv.Itoa(ts) + `"`
}
func ts2mysql(ts int) string {
	return `FROM_UNIXTIME(` + strconv.Itoa(ts) + `)`
}

func TestManager(t *testing.T) {
	cases := []struct {
		name     string
		setup    func(*testing.T) (*sql.DB, *Manager)
		teardown func(*testing.T, *sql.DB)
		ts       func(int) string
	}{
		{"sqlite", initdb, teardownsqlite, ts2sqlite},
		{"mysql", initmysqldb, teardownmysql, ts2mysql},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Run("Scan OK", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)
				s := `INSERT INTO testok (eint, estr, t) VALUES (10, "test", ` + c.ts(1462233600) + `)`
				if _, err := db.Exec(s); err != nil {
					t.Fatalf("Cannot insert preset data into testok: %s", err)
					db.Close()
					os.Exit(1)
				}
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
			})

			t.Run("Insert", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)
				ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-05-04 08:00:00 +0800")
				data := testok{1, 2, 3, "insert", ti}

				if _, err := m.Insert(data); err != nil {
					qstr, _ := m.makeInsert(data)
					t.Fatalf("Error inserting data: %s\nSQL: %s", err, qstr)
				}

				var cnt int
				row := db.QueryRow(`SELECT COUNT(eint) FROM testok WHERE eint=3 AND estr="insert" AND t=` + c.ts(1462320000))
				if err := row.Scan(&cnt); err != nil {
					t.Fatalf("Cannot scan COUNT(eint) for insert: %s", err)
				}
				if cnt != 1 {
					t.Errorf("There should be only one result after inserting, but we got %d", cnt)
				}
			})

			t.Run("LoadSimple", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)

				ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-05-04 08:00:00 +0800")
				data := testai{0, "load simple", ti}

				if _, err := m.Insert(data); err != nil {
					t.Fatalf("Error inserting data: %s", err)
				}

				var target testai
				if err := m.LoadSimple(&target, 1); err != nil {
					t.Fatalf("Error loading from simple table: %s", err)
				}
				if target.ExportInt != 1 {
					t.Errorf("Data mispatch: expect eint to be 1, got %d", target.ExportInt)
				}
				if target.ExportString != "load simple" {
					t.Errorf("Data mispatch: expect estr to be 'load simple', got '%s'", target.ExportString)
				}
				if expect, actual := ti.Unix(), target.ExportTime.Unix(); expect != actual {
					t.Errorf("Data mispatch: expect t to be %d, got %d", expect, actual)
				}
			})

			t.Run("Update", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)
				ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-05-04 08:00:00 +0800")
				data := testok{1, 2, 3, "update", ti}

				if _, err := m.Insert(data); err != nil {
					t.Fatalf("Error inserting data for updating: %s", err)
				}

				data.ExportInt = 4
				if _, err := m.Update(data, `eint=? AND estr=? AND t=`+c.ts(1462320000), 3, "update"); err != nil {
					t.Fatalf("Error updating data: %s", err)
				}

				var cnt int
				row := db.QueryRow(`SELECT COUNT(eint) FROM testok WHERE eint=4 AND estr="update" AND t=` + c.ts(1462320000))
				if err := row.Scan(&cnt); err != nil {
					t.Fatalf("Cannot scan COUNT(eint) for update: %s", err)
				}
				if cnt != 1 {
					t.Errorf("There should be only one result after updating, but we got %d", cnt)
				}
			})

			t.Run("Delete", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)

				ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-05-04 08:00:00 +0800")
				data := testok{1, 2, 3, "delete", ti}

				if _, err := m.Insert(data); err != nil {
					t.Fatalf("Error inserting data for deleting: %s", err)
				}

				_, err := m.Delete(data)
				if err != nil {
					t.Fatalf("Error deleting data: %s", err)
				}

				var cnt int
				row := db.QueryRow(`SELECT COUNT(*) FROM testok WHERE eint=3 AND estr="delete" AND t=` + c.ts(1462320000))
				if err := row.Scan(&cnt); err != nil {
					t.Fatalf("Cannot scan COUNT(eint) for delete: %s", err)
				}
				if cnt != 0 {
					t.Errorf("There should be no result after deleting, but we got %d", cnt)
				}
			})

			t.Run("Insert Auto Increment", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)
				ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-05-04 08:00:00 +0800")
				data := testai{ExportString: "insert", ExportTime: ti}

				if _, err := m.Insert(&data); err != nil {
					t.Fatalf("Error inserting ai data: %s", err)
				}

				var cnt int
				row := db.QueryRow(`SELECT COUNT(eint) FROM testai WHERE estr="insert" AND t=` + c.ts(1462320000))
				if err := row.Scan(&cnt); err != nil {
					t.Fatalf("Cannot scan COUNT(eint) for ai insert: %s", err)
				}
				if cnt != 1 {
					t.Errorf("There should be only one result after ai inserting, but we got %d", cnt)
				}

				// test for auto insert id
				if data.ExportInt != 1 {
					t.Errorf("Failed to set last insert id, get %d", data.ExportInt)
				}
			})

			t.Run("Query Error", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)
				data := testok{}
				rows := m.Query(data, `THIS IS INVALID SQL QUERY`)
				err := rows.Err()
				if err == nil {
					t.Fatal("QueryError should return error, got nil")
				}
			})

			t.Run("Build", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)
				ti, _ := time.Parse("2006-01-02 15:04:05 -0700", "2016-10-20 08:00:00 +0800")
				data := testok{2, 3, 4, "build", ti}

				if _, err := m.Build(data, `INSERT INTO testok (%cols%) VALUES (%vals%)`, driver.QInsert); err != nil {
					t.Fatalf("Error inserting ai data: %s", err)
				}

				var cnt int
				row := db.QueryRow(`SELECT COUNT(eint) FROM testok WHERE eint=4 AND estr="build" AND t=` + c.ts(1476921600))
				if err := row.Scan(&cnt); err != nil {
					t.Fatalf("Cannot scan COUNT(eint) for build: %s", err)
				}
				if cnt != 1 {
					t.Errorf("There should be only one result after building, but we got %d", cnt)
				}
			})

			t.Run("Index", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)
				typ := reflect.TypeOf(testok{})
				info, ok := m.info[typ]
				if !ok {
					t.Fatalf("No index data found!")
				}
				idx := info.Indexes
				if l := len(idx); l != 3 {
					t.Errorf("Expected to have 3 indexes, got %d", l)
				}
				find := func(typ, name string, cols []string) {
					sort.Strings(cols)
					for _, v := range idx {
						if v.Name != name {
							continue
						}

						sort.Strings(v.Cols)
						if typ != v.Type {
							t.Errorf("Expected index %s to be a %s index, get %s", name, typ, v.Type)
						}
						if !reflect.DeepEqual(cols, v.Cols) {
							t.Errorf("Expected %s index %s to have %v, got %v", v.Type, name, cols, v.Cols)
						}
						return
					}
					t.Errorf("Expected to have %s index %s, but not found", typ, name)
				}
				find(driver.IndexTypeIndex, "a", []string{"eint", "estr"})
				find(driver.IndexTypeUnique, "b", []string{"eint"})
				find(driver.IndexTypeIndex, "c", []string{"t"})
			})

			t.Run("AppendTo", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)

				// insert data
				m.Insert(testok{ExportInt: 10})
				m.Insert(testok{ExportInt: 11})

				t.Run("Ptr", func(t *testing.T) {
					arr := []*testok{}
					qstr := m.BuildSQL(testok{}, `SELECT %cols% FROM %table%`, driver.QSelect)
					rows := m.Query(testok{}, qstr)
					if err := rows.AppendTo(&arr); err != nil {
						t.Fatalf("unexpected error: %s", err)
					}
				})
				t.Run("NonPtr", func(t *testing.T) {
					arr := []testok{}
					qstr := m.BuildSQL(testok{}, `SELECT %cols% FROM %table%`, driver.QSelect)
					rows := m.Query(testok{}, qstr)
					if err := rows.AppendTo(&arr); err != nil {
						t.Fatalf("unexpected error: %s", err)
					}
				})
			})

			t.Run("SetTo", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)

				// insert data
				m.Insert(testok{ExportInt: 10})
				m.Insert(testok{ExportInt: 11})

				getKey := func(v interface{}) interface{} {
					return v.(*testok).ExportInt
				}

				t.Run("Ptr", func(t *testing.T) {
					arr := map[int]*testok{}
					qstr := m.BuildSQL(testok{}, `SELECT %cols% FROM %table%`, driver.QSelect)
					rows := m.Query(testok{}, qstr)
					if err := rows.SetTo(&arr, getKey); err != nil {
						t.Fatalf("unexpected error: %s", err)
					}
					if l := len(arr); l != 2 {
						t.Errorf("unexpected map size: %d", l)
					}
					if arr[10] == nil {
						t.Error("expected to have record#10")
					}
					if arr[11] == nil {
						t.Error("expected to have record#10")
					}
				})
				t.Run("NonPtr", func(t *testing.T) {
					arr := map[int]testok{}
					qstr := m.BuildSQL(testok{}, `SELECT %cols% FROM %table%`, driver.QSelect)
					rows := m.Query(testok{}, qstr)
					if err := rows.SetTo(&arr, getKey); err != nil {
						t.Fatalf("unexpected error: %s", err)
					}
					if l := len(arr); l != 2 {
						t.Errorf("unexpected map size: %d", l)
					}
					if arr[10].ExportInt == 0 {
						t.Error("expected to have record#10")
					}
					if arr[11].ExportInt == 0 {
						t.Error("expected to have record#10")
					}
				})
			})

			t.Run("SQLIn", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)

				arr := []int{1, 2, 3}
				expect := `IN (?,?,?)`
				if actual := m.SQLIn(arr); actual != expect {
					t.Fatalf("expected [%s], got [%s]", expect, actual)
				}
			})

			t.Run("AsArgs", func(t *testing.T) {
				db, m := c.setup(t)
				defer c.teardown(t, db)

				arr := []int{1, 2, 3}
				res := m.AsArgs(arr)
				for x, v := range res {
					if v != arr[x] {
						t.Errorf("expect %v, got %v at %d", arr[x], v, x)
					}
				}
			})
		})
	}
}
