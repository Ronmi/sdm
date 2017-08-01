package driver

const (
	IndexTypeIndex   = "idx"
	IndexTypeUnique  = "uniq"
	IndexTypePrimary = "pri"
)

// Index represents defination of an index
type Index struct {
	Type string
	Name string
	Cols []string
}

// HasCol determins is this index contains the column
func (i *Index) HasCol(name string) bool {
	for _, c := range i.Cols {
		if name == c {
			return true
		}
	}

	return false
}

// Column represents defination of a column, for internal use only
type Column struct {
	ID   int    // field id
	AI   bool   // auto increment
	Name string // column name
}
