package sdm

import (
	"database/sql"
	"reflect"
	"sync"

	"github.com/Ronmi/sdm/driver"
)

// Stmt wraps sql.Stmt
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
		if s.columns, err = r.Columns(); err != nil {
			for x, v := range s.columns {
				s.columns[x] = s.drv.ParseColumnName(v)
			}
		}
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

// QueryRow is like Manager.QueryRow, but executes on prepared statement
func (s *Stmt) QueryRow(data interface{}, args ...interface{}) error {
	rows := s.Query(args...)
	defer rows.Close()
	if !rows.Next() {
		return nil
	}

	return rows.Scan(data)
}
