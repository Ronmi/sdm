[![Build Status](https://ci.ronmi.tw/api/badges/ronmi/sdm/status.svg)](https://ci.ronmi.tw/ronmi/sdm)

SDM stands for *Struct-Database Mapping*. It helps you to play with simple Go struct and DB table.

SDM is **LIGHTYEARS FAR FROM ORM**. It's designed for people who used to write raw SQL.

```go
package main

import (
	"database/sql"
	"log"

	"git.ronmi.tw/ronmi/sdm"
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
	ColdDown int    `sdm:"cd,idx_cd"`                     // cd INTEGER
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

func main() {
	// prepare db connection
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create sdm
	m := sdm.New(db, sqlite3.New())

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
		// scan is so simple
		r.Scan(&grp)
		log.Printf("dump: %#v", grp)
	}
}
```

# License

Copyright(c) 2016-2017 Ronmi Ren (Chung-ping Jen) <ronmi.ren@gmail.com>

MIT
