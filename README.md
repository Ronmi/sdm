[![Build Status](https://ci.ronmi.tw/api/badges/ronmi/sdm/status.svg)](https://ci.ronmi.tw/ronmi/sdm)

This package is just only experiment of mine.

```go
type mytype struct {
    MyInt int `sdm:"myint"`
    MyStr string `sdm:"mystr"`
}


db := sql.Open("some_driver", "some dsn")
defer db.Close()
db.Exec(`CREATE TABLE test (myint int, mystr varchar(10))`)

mgr := sdm.New()

// insert
myval := mytype{1, "test"}
mgr.Insert(db, "test", myval)

// select
var result mytype
qstr := `SELECT * FROM test`
dbrows := db.Query(qstr)
rows := mgr.Proxify(dbrows, result)
for rows.Next() {
    rows.Scan(&result)
}

// update
result.MyInt=10
mgr.Update(db, "test", result, `myint=? AND mystr=?`, 1, "test")

// delete
mgr.Delete(db, "test", result)
```

# License

Copyright(c) 2016 Ronmi Ren (Chung-ping Jen) <ronmi.ren@gmail.com>

GPLv3
