package sdm

import "reflect"

// Extension defines what facilities an extension should provide
//
// An extension is a helper to read data into struct according to existing
// type info maintained in SDM.
//
// Implementations MUST write detailed document to describe whether a type is
// supported or not.
type Extension interface {
	// Mean to be called by sdm.Manager. An extension might not work
	// before initialized.
	Init(typ reflect.Type, mapFieldToColumn map[int]string)
}

// ErrExtension indicates something goes wrong when calling ReadTo
//
// It implements error interface
type ErrExtension struct {
	ExtName    string // extension name
	StructName string // reading into this struct
	FieldName  string // found error on this field
	Reason     error
}

func (e *ErrExtension) Error() string {
	return e.ExtName + ": error reading into " + e.StructName + "." + e.FieldName + ": " + e.Reason.Error()
}

func (e *ErrExtension) String() string {
	return e.Error()
}

// Ext fills data ito struct using specified extension, panics if not registered and
// auto registering is not enabled
func (m *Manager) Ext(data interface{}, e Extension) {
	v := reflect.ValueOf(data)
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	info := m.getInfo(t)

	cols := make(map[int]string)
	for _, c := range info.Fields {
		cols[c.ID] = c.Name
	}

	e.Init(t, cols)
}
