package sdmext

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Ronmi/sdm"
	"github.com/Ronmi/sdm/driver"
)

type testok struct {
	A int     `sdm:"a"`
	B float64 `sdm:"b"`
	C bool    `sdm:"c"`
	D string  `sdm:"d"`
	E []byte  `sdm:"e"`
	F []rune  `sdm:"f"`
}

func TestExtGeneral(t *testing.T) {
	driver.AnsiStub()
	m := sdm.New(nil, "ansistub")
	m.Reg(testok{})
}

func egok(m *sdm.Manager, t *testing.T) {
	data := map[string]string{
		"a": "1",
		"b": "1.1",
		"c": "ok",
		"d": "string",
		"e": "string2",
		"f": "string3",
	}

	f := func(k string) string {
		return data[k]
	}

	ext := new(ExtGeneral)
	actual := testok{}
	m.Ext(actual, ext)
	if err := ext.ReadTo(&actual, f); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if actual.A != 1 {
		t.Errorf("expected A to be 1, get %d", actual.A)
	}
	if actual.B != 1.1 {
		t.Errorf("expected B to be 1.1, get %f", actual.B)
	}
	if !actual.C {
		t.Errorf("expected C to be true, get %v", actual.C)
	}
	if actual.D != "string" {
		t.Errorf("expected D to be 'string', get %s", actual.D)
	}
	if string(actual.E) != "string2" {
		t.Errorf("expected E to be 'string2', get %s", actual.E)
	}
	if string(actual.F) != "string3" {
		t.Errorf("expected F to be 'string3', get %s", string(actual.F))
	}
}

func egwithempty(m *sdm.Manager, t *testing.T) {
	f := func(k string) string {
		return ""
	}

	ext := new(ExtGeneral)
	actual := testok{
		A: 1,
		B: 1.1,
		C: true,
		D: "string",
		E: []byte("string2"),
		F: []rune("string3"),
	}
	m.Ext(actual, ext)
	if err := ext.ReadTo(&actual, f); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if actual.A != 1 {
		t.Errorf("expected A to be 1, get %d", actual.A)
	}
	if actual.B != 1.1 {
		t.Errorf("expected B to be 1.1, get %f", actual.B)
	}
	if !actual.C {
		t.Errorf("expected C to be true, get %v", actual.C)
	}
	if actual.D != "string" {
		t.Errorf("expected D to be 'string', get %s", actual.D)
	}
	if string(actual.E) != "string2" {
		t.Errorf("expected E to be 'string2', get %s", actual.E)
	}
	if string(actual.F) != "string3" {
		t.Errorf("expected F to be 'string3', get %s", string(actual.F))
	}
}

func ExampleExtGeneral() {
	// a struct for test
	type myStruct struct {
		A int    `sdm:"a"`
		B string `sdm:"b"`
	}

	// create fake request for demo purpose
	r, _ := http.NewRequest("GET", "https://example.com/?a=1&b=test", nil)

	// create sdm.Manager and register our struct
	driver.MySQLStub()
	m := sdm.New(nil, "mysqlstub")
	m.Reg(myStruct{})

	// here's code should look like in your http handler
	var formData myStruct
	ext := &ExtGeneral{}
	m.Ext(formData, ext)

	r.ParseForm()
	if err := ext.ReadTo(&formData, r.FormValue); err != nil {
		// handle error here
		fmt.Print(err)
	}

	// dump struct for testing purpose
	fmt.Printf("%+v", formData)
	// output: {A:1 B:test}
}
