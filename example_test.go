package sdm

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Ronmi/sdm/driver"
	_ "github.com/Ronmi/sdm/driver/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func Example() {
	// You'll have to import sqlite3 SDM driver to use SDM with sqlite:
	// import _ "github.com/Ronmi/sdm/driver/sqlite3"
	//
	// see package driver for more info about writing your driver in seconds.

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

	// prepare db connection
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create sdm
	m := New(db, "sqlite3")

	// register types
	m.Reg(Group{}, Member{})

	// It is suggested to manage DB schema on your own.
	// If you really want, SDM can create tables.
	m.CreateTablesNotExist()

	g := Group{Name: "Star Force"}
	// insert records into table
	if _, err := m.Insert(&g); err != nil { // pointer type, will try to fill ID
		log.Fatal(err)
	}
	// value type, does not fill ID
	if _, err := m.Insert(Member{GroupID: g.ID, Position: "DD", ColdDown: 8}); err != nil {
		log.Fatal(err)
	}

	var grp Group
	// make a query
	r := m.Query(grp, `SELECT * FROM 'group'`)
	// bit faster but acts same as
	// r := m.Query(grp, `SELECT %cols% FROM %table%`)
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

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	m := New(db, "sqlite3")

	// register types
	var grp Group
	m.Reg(grp)

	// create table in db
	m.CreateTables()

	data := Group{Name: "Sun Moon Lake"}
	m.Build(data, `INSERT INTO %table% (%cols%) VALUES (%vals%)`, driver.QInsert)

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

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	m := New(db, "sqlite3")

	// register types
	var grp Group
	m.Reg(grp)

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

func ExampleManager_Insert() {
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

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	m := New(db, "sqlite3")

	// register types
	grp := Group{Name: "Dolan"}
	m.Reg(grp)

	// create table in db
	m.CreateTables()

	// calling insert with pointer type, let SDM fill ID field for you
	m.Insert(&grp)

	fmt.Printf("Group ID: %d", grp.ID)
	// output: Group ID: 1
}
