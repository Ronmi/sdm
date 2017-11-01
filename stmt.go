package sdm

import (
	"database/sql"
	"reflect"

	"github.com/Ronmi/sdm/driver"
)

// Stmt warps sql.Stmt
//
// QueryRow is not supported since it need some modifications to sdm core.
type Stmt struct {
	*sql.Stmt
	def     map[string]driver.Column
	t       reflect.Type
	drv     driver.Driver
	columns []string
}

func (s *Stmt) Query(args ...interface{}) *Rows {
	r, err := s.Stmt.Query(args...)
	var c []string
	if err == nil {
		c, err = r.Columns()
	}

	return &Rows{
		rows:    r,
		def:     s.def,
		columns: c,
		e:       err,
		t:       s.t,
		drv:     s.drv,
	}
}

func (s *Stmt) QueryRow(args ...interface{}) *Row {
	return &Row{r: s.Query(args...)}
}
