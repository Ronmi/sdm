package readstruct

import (
	"go/token"
	"io"
)

// ExtractFromDir extracts complete and valid struct info from all go files in dir.
//
// dir should be end with path separator, otherwise behavier is unspecified and
// might change in the future.
//
// Files causing parse error are silently ignored. The only possible error returned
// comes from filepath.Glob(). Also, incomplete strcuts are silently ignored.
func ExtractFromDir(dir string) (info []Info, err error) {
	fset := token.NewFileSet()
	files, err := parseFromDir(fset, dir)
	if err != nil {
		return
	}

	for _, f := range files {
		i := extractFromAst(f)
		info = append(info, i...)
	}

	return
}

// ExtractFromFile extracts complete and valid struct info from specified file.
//
// Only errors from parser.ParseFile is returned. Also, incomplete strcuts are
// silently ignored.
func ExtractFromFile(fn string) (info []Info, err error) {
	fset := token.NewFileSet()
	f, err := parseFromFile(fset, fn)
	if err != nil {
		return
	}

	info = extractFromAst(f)
	return
}

// ExtractFromReader extracts complete and valid struct info from io.Reader.
//
// Only errors from parser.ParseFile is returned. Also, incomplete strcuts are
// silently ignored.
func ExtractFromReader(r io.Reader) (info []Info, err error) {
	fset := token.NewFileSet()
	f, err := parseFromReader(fset, r)
	if err != nil {
		return
	}

	info = extractFromAst(f)
	return
}
