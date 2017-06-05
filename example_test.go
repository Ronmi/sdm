package sdm

import (
	"database/sql"
	"fmt"
	"log"

	"git.ronmi.tw/ronmi/sdm/driver/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

// Member represents a member in a group
//
// Inline comments are *DEMO*, sqlite3 driver will generate those for you
// when you call CreateTables
type Member struct {
	// CONSTRAINT member_pk PRIMARY KEY (id) AUTOINCREMENT
	ID int `sdm:"id,ai,pri_member_pk"` // id INTEGER

	// CONSTRAINT position UNIQUE (group_id,battle_position)
	GroupID  int    `sdm:"group_id,uniq_position"`        // group_id INTEGER
	Position string `sdm:"battle_position,uniq_position"` // battle_position TEXT

	// sqlite doesn't support INDEX constraint, so sad
	ColdDown int `sdm:"cd,idx_cd"` // cd INTEGER
}

// Group represents a group of people
//
// Inline comments are *DEMO*, sqlite3 driver will generate those for you
// when you call CreateTables
type Group struct {
	// CONSTRAINT group_pk PRIMARY KEY (id) AUTOINCREMENT
	ID int `sdm:"id,ai"` // id INTEGER

	// CONSTRAINT group_name UNIQUE (name)
	Name string `sdm:"name,uniq_group_name"` // name TEXT
}

func Example() {
	// prepare db connection
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create sdm
	m := New(db, sqlite3.New())

	// register types
	m.Reg(Member{}, Group{})

	// create table in db
	m.CreateTables()

	// insert record into table
	if _, err := m.Insert(Group{Name: "Star Force"}); err != nil {
		log.Fatal(err)
	}

	var grp Group
	// make a query
	r := m.Query(grp, `SELECT * FROM 'group'`)
	defer r.Close()

	for r.Next() {
		var grp Group
		// scan is so simple
		r.Scan(&grp)
		fmt.Printf("Group name: %s\n", grp.Name)
	}

	// also support partial load
	r = m.Query(grp, `SELECT id FROM 'group'`)
	defer r.Close()
	for r.Next() {
		var grp Group
		r.Scan(&grp)
		fmt.Printf("Group name is empty? %v\n", grp.Name == "")
		fmt.Printf("Group ID: %d\n", grp.ID)
	}

	// output: Group name: Star Force
	// Group name is empty? true
	// Group ID: 1
	//
}

func ExampleManager_Build() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	m := New(db, sqlite3.New())

	// register types
	var mem Member
	var grp Group
	m.Reg(mem, grp)

	// create table in db
	m.CreateTables()

	data := Group{Name: "Sun Moon Lake"}
	m.Build(data, `INSERT INTO %table% (%cols%) VALUES (%vals%)`)

	// verify db using Go's sql package
	var cnt int
	row := db.QueryRow(`SELECT COUNT(*) FROM 'group'`)
	if err := row.Scan(&cnt); err != nil {
		fmt.Printf("Cannot scan COUNT(*) for build: %s", err)
		return
	}
	fmt.Printf("Got %d record", cnt)
	// output: Got 1 record
}

func ExampleManager_Proxify() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	m := New(db, sqlite3.New())

	// register types
	var mem Member
	var grp Group
	m.Reg(mem, grp)

	// create table in db
	m.CreateTables()

	m.Insert(Group{Name: "Moon Moon"})

	// a super complex query
	qstr := `SELECT * FROM 'group' WHERE id>0`
	r, err := db.Query(qstr)
	if err != nil {
		// error handling
		log.Fatal(err)
	}

	rows := m.Proxify(r, grp)
	defer rows.Close()
	for rows.Next() {
		var data Group
		rows.Scan(&data)
		fmt.Println(data.Name)
	}

	// output: Moon Moon
	//
}
