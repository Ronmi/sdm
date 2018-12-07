package sdm

import (
	"database/sql"

	"github.com/Ronmi/sdm/driver"
)

// Executable abstracts common facilities between manager and Tx.
//
// If you are writing some queries which supports both transaction and
// non-transaction mode should depend on this instead of Manager/Tx.
type Executable interface {
	SQLIn(arr interface{}) string
	Query(typ interface{}, qstr string, args ...interface{}) *Rows
	QueryRow(data interface{}, qstr string, args ...interface{}) error
	Prepare(data interface{}, qstr string) (*Stmt, error)
	PrepareSQL(data interface{}, tmpl string, qType driver.QuotingType) (*Stmt, error)
	Insert(data interface{}) (sql.Result, error)
	Update(data interface{}, where string, whereargs ...interface{}) (sql.Result, error)
	Delete(data interface{}) (sql.Result, error)
	RunBulk(b Bulk) (sql.Result, error)
	Val(data interface{}) []interface{}
	ValIns(data interface{}) []interface{}
	Stmt(stmt *Stmt) (ret *Stmt)
}
