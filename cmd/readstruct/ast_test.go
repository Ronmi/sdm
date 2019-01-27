package readstruct

import (
	"go/token"
	"log"
	"testing"
)

func TestExtractFromAst(t *testing.T) {
	fset := token.NewFileSet()
	files, err := parseFromDir(fset, "testdata/")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	info := extractFromAst(files[0])
	if l := len(info); l != 2 {
		log.Fatalf("Expected 2 structs, got %d", l)
	}

	t.Logf("%#v", info)
}
