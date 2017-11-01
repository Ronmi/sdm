package sdm

import (
	"database/sql"
	"reflect"
	"sync"

	"github.com/Ronmi/sdm/driver"
)

// Stmt warps sql.Stmt
//
// QueryRow is not supported since it need some modifications to sdm core.
type Stmt struct {
	stmt    *sql.Stmt
	def     map[string]driver.Column
	t       reflect.Type
	drv     driver.Driver
	columns []string
	lock    sync.Mutex
}

// Exec is identical to sql.Stmt.Exec
func (s *Stmt) Exec(args ...interface{}) (sql.Result, error) {
	return s.stmt.Exec(args...)
}

// Query is just sql.Stmt.Query, excepts it wrap the sql.Rows in sdm.Rows
func (s *Stmt) Query(args ...interface{}) *Rows {
	r, err := s.stmt.Query(args...)

	s.lock.Lock()
	if err == nil && s.columns == nil {
		s.columns, err = r.Columns()
	}
	s.lock.Unlock()

	return &Rows{
		rows:    r,
		def:     s.def,
		columns: s.columns,
		e:       err,
		t:       s.t,
		drv:     s.drv,
	}
}

// QueryRow is like sql.Stmt.QueryRow, but implementing detail differs.
//
// See sdm.Row for more info.
func (s *Stmt) QueryRow(args ...interface{}) *Row {
	return &Row{r: s.Query(args...)}
}
