package readstruct

import (
	"errors"
	"go/ast"
	"go/token"
	"path/filepath"
)

// FindFirstIn filters all go source files in dir, return first matching struct info
// which test by filter function "by"
//
// Example usage:
//
//     FindFirstIn("testdata", ByName("A"))
func FindFirstIn(dir string, by func(*Info) bool) (info Info, err error) {
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}
	dir += "*.go"
	files, err := filepath.Glob(dir)
	if err != nil {
		return
	}

	fset := token.NewFileSet()
	for _, fn := range files {
		if a, e := parseFromFile(fset, fn); e == nil {
			i, ok := filterResult(a, by)
			if ok {
				info = i
				return
			}
		}
	}

	err = errors.New("Cannot find matching structure.")
	return
}

func filterResult(a *ast.File, f func(*Info) bool) (info Info, ok bool) {
	data := extractFromAst(a)
	for _, i := range data {
		if ok = f(&i); ok {
			info = i
			return
		}
	}

	return
}

func ByName(name string) (ret func(*Info) bool) {
	return func(i *Info) (ok bool) {
		return i.Name == name
	}
}
