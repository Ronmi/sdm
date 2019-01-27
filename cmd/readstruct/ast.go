package readstruct

import (
	"go/ast"
	"strings"
)

func extractFromAst(f *ast.File) (ret []Info) {
	pkg := f.Name.Name
	ret = make([]Info, 0, len(f.Scope.Objects))

	for name, obj := range f.Scope.Objects {
		if obj.Kind != ast.Typ {
			continue
		}

		spec := obj.Decl.(*ast.TypeSpec)
		st, ok := spec.Type.(*ast.StructType)
		if !ok {
			continue
		}
		if st.Incomplete {
			continue
		}

		info := Info{
			Package: pkg,
			Name:    name,
			Fields:  make([]Field, 0, len(st.Fields.List)),
		}

		for _, f := range st.Fields.List {
			info.Fields = append(info.Fields, extractFieldFromAst(f))
		}

		ret = append(ret, info)
	}

	return
}

func extractFieldFromAst(f *ast.Field) (ret Field) {
	if len(f.Names) == 0 {
		return extractEmbedFieldFromAst(f)
	}

	ret.Name = f.Names[0].Name
	if i, ok := f.Type.(*ast.Ident); ok {
		ret.RawType = i.Name
	}
	ret.Exported = strings.ToUpper(ret.Name[0:1]) == ret.Name[0:1]
	if f.Tag != nil {
		ret.Tags = f.Tag.Value
	}

	return
}

func extractEmbedFieldFromAst(f *ast.Field) (ret Field) {
	if x, ok := f.Type.(*ast.StarExpr); ok {
		ret.Name = x.X.(*ast.Ident).Name
		ret.RawType = "*" + ret.Name
	} else {
		ret.Name = f.Type.(*ast.Ident).Name
		ret.RawType = ret.Name
	}

	ret.Exported = strings.ToUpper(ret.Name[0:1]) == ret.Name[0:1]

	if f.Tag != nil {
		ret.Tags = f.Tag.Value
	}

	return
}
