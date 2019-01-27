package readstruct

import (
	"go/ast"
	"go/token"
	"os"
	"testing"
)

func dumpType(t *testing.T, typ *ast.TypeSpec) {
	t.Logf("      Type name: %s", typ.Name)
	if x, ok := typ.Type.(*ast.StructType); ok {
		dumpStruct(t, x)
	} else {
		t.Logf("      Type type: %#v", typ.Type)
	}
}

func dumpStruct(t *testing.T, typ *ast.StructType) {
	t.Logf("      Struct incomplete: %v", typ.Incomplete)
	for _, f := range typ.Fields.List {
		if len(f.Names) != 1 {
			dumpEmbedField(t, f)
			continue
		}
		t.Logf("        Names: %+v", f.Names)
		t.Logf("          Type: %#v", f.Type)
		if f.Tag != nil {
			t.Logf("          Tags: %s", f.Tag.Value)
		}
	}
}

func dumpEmbedField(t *testing.T, f *ast.Field) {
	if x, ok := f.Type.(*ast.StarExpr); ok {
		dumpPtrEmbedField(t, f, x)
		return
	}
	x := f.Type.(*ast.Ident)
	t.Logf("        Names: %s", x.Name)
	t.Logf("          Type: %s", x.Name)
	if f.Tag != nil {
		t.Logf("          Tags: %s", f.Tag.Value)
	}
}

func dumpPtrEmbedField(t *testing.T, f *ast.Field, x *ast.StarExpr) {
	y := x.X.(*ast.Ident)
	t.Logf("        Names: %s", y.Name)
	t.Logf("          Type: *%s", y.Name)
	if f.Tag != nil {
		t.Logf("          Tags: %s", f.Tag.Value)
	}
}

func TestParseFromFile(t *testing.T) {
	data := []struct {
		fn  string
		err bool
	}{
		{fn: "testdata/a.go", err: false},
		{fn: "testdata/broken.go", err: true},
	}

	for _, d := range data {
		fset := token.NewFileSet()
		f, err := parseFromFile(fset, d.fn)
		if d.err && err == nil {
			t.Fatalf("Expected error returned for %s, got nothing.", d.fn)
		}
		if !d.err && err != nil {
			t.Fatalf("Unexpected error for %s: %s", d.fn, err)
		}

		if !d.err {
			t.Log("Dump ast:")
			t.Logf("  Package name: %s", f.Name.Name)
			t.Log("  Contents:")
			for name, obj := range f.Scope.Objects {
				t.Logf("    %s:", name)
				t.Logf("      Kind: %s", obj.Kind.String())
				t.Logf("      Decl: %#v", obj.Decl)
				t.Logf("      Data: %#v", obj.Data)
				t.Logf("      Type: %#v", obj.Type)

				if obj.Kind == ast.Typ {
					dumpType(t, obj.Decl.(*ast.TypeSpec))
				}
			}

			t.Logf("  Raw: %#v", f)
		}
	}
}

func TestParseFromDir(t *testing.T) {
	fset := token.NewFileSet()
	files, err := parseFromDir(fset, "testdata/")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if l := len(files); l != 1 {
		t.Fatalf("Expected only 1 result, got %d.", l)
	}

	fn := fset.File(files[0].Package).Name()
	if fn != "testdata/a.go" {
		t.Fatalf("Expected to read a.go, go %s", fn)
	}
}

func TestParseFromReader(t *testing.T) {
	data := []struct {
		fn  string
		err bool
	}{
		{fn: "testdata/a.go", err: false},
		{fn: "testdata/broken.go", err: true},
	}

	for _, d := range data {
		fset := token.NewFileSet()
		r, err := os.Open(d.fn)
		defer r.Close()

		_, err = parseFromReader(fset, r)
		if d.err && err == nil {
			t.Fatalf("Expected error returned for %s, got nothing.", d.fn)
		}
		if !d.err && err != nil {
			t.Fatalf("Unexpected error for %s: %s", d.fn, err)
		}
	}
}
