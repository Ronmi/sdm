package readstruct

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
)

func parseFromDir(fset *token.FileSet, dir string) (ret []*ast.File, err error) {
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}
	dir += "*.go"
	files, err := filepath.Glob(dir)
	if err != nil {
		return
	}
	ret = make([]*ast.File, 0, len(files))

	for _, f := range files {
		if a, e := parseFromFile(fset, f); e == nil {
			ret = append(ret, a)
		}
	}

	return
}

func parseFromFile(fset *token.FileSet, fn string) (ret *ast.File, err error) {
	ret, err = parser.ParseFile(fset, fn, nil, parser.ParseComments)
	return
}

func parseFromReader(fset *token.FileSet, r io.Reader) (ret *ast.File, err error) {
	ret, err = parser.ParseFile(fset, "wtf", r, parser.ParseComments)
	return
}
