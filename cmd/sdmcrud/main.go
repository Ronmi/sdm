package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Ronmi/sdm/cmd/readstruct"
)

func doc() {
	x := os.Args[0]
	fmt.Fprintf(
		flag.CommandLine.Output(),
		`
As %s grnerates one struct per execution, you can chain it up like

  (%s -head -p myrepo ; %s -p myrepo -d ../mydata/ -n User -k ID,Email,Name ; %s -p myrepo -d ../mydata/ -n Article -k ID,AuthorID,Title) | goimports

Example usage of generated repository:

  func updateInTransaction(repo *MyStructRepo, mydata *MyStruct) (err error) {
      tx, err := repo.Begin()
      if err != nil {
          return
      }
      defer tx.Rollback()

      if err = tx.Update(mydata); err != nil {
          return
      }
      return tx.Commit()
  }
`,
		x,
		x, x, x,
	)
}

type Code struct {
	code  string
	quote string
}

func main() {
	var (
		dir  string
		name string
		keys string
		pkg  string
		head bool
	)

	u := flag.Usage
	flag.Usage = func() {
		u()
		doc()
	}

	flag.StringVar(&dir, "d", ".", "Where to find source files. (Defult to current dir)")
	flag.StringVar(&name, "n", ".", "Struct name. (required)")
	flag.StringVar(&keys, "k", "", "Search keys for SELECT and DELETE (and first for UPDATE). Separate multiple keys with comma. (required)")
	flag.StringVar(&pkg, "p", "", "Package name. (optional, affects generated code if differ from package of the struct)")
	flag.BoolVar(&head, "head", false, "Generates package and import then exit. (requires -p)")
	flag.Parse()

	if head {
		if pkg == "" {
			flag.Usage()
			return
		}
		buf := &bytes.Buffer{}
		genPackage(buf, pkg)
		if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
			log.Fatalf("Error writing to stdout: %s", err)
		}
		return
	}

	if name == "" || keys == "" {
		flag.Usage()
		return
	}

	gen(pkg, dir, name, keys)
}

func gen(pkg, dir, name, keys string) {
	buf := &bytes.Buffer{}
	keyArr := strings.Split(keys, ",")

	// load strcut info
	info, err := readstruct.ExtractFromDir(dir)
	if err != nil {
		log.Fatalf("Failed to load struct info from source files: %s", err)
	}
	var st readstruct.Info
	for _, i := range info {
		if i.Name == name {
			st = i
		}
	}
	if st.Name == "" {
		log.Fatalf("Cannot find specified struct in %s!", dir)
	}
	if pkg == "" {
		pkg = st.Package
	}

	cCode, cStmt := genCreate(st, pkg, name)
	sCode, sStmt := genSelect(st, pkg, name, keyArr)
	uCode, uStmt := genUpdate(st, pkg, name, keyArr[0])
	dCode, dStmt := genDelete(st, pkg, name, keyArr)

	stmts := []map[string]Code{
		cStmt,
		sStmt,
		uStmt,
		dStmt,
	}
	genStruct(buf, name, stmts)
	genInit(buf, st, pkg, name, stmts)

	// methods
	for _, x := range []*bytes.Buffer{cCode, sCode, uCode, dCode} {
		buf.Write(x.Bytes())
	}

	// transactions
	genTransaction(buf, name)

	if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
		log.Fatalf("Error writing to stdout: %s", err)
	}
}

func genPackage(buf *bytes.Buffer, pkg string) {
	fmt.Fprintf(
		buf,
		`package %s

import (
	"errors"

	"github.com/Ronmi/sdm"
	"github.com/Ronmi/sdm/driver"
)

`,
		pkg,
	)
	return
}

func genStruct(buf *bytes.Buffer, name string, stmts []map[string]Code) {
	// find max length of stmt name
	max := 1
	for _, part := range stmts {
		for name, _ := range part {
			l := len(name)
			if l > max {
				max = l
			}
		}
	}
	x := "%-" + strconv.Itoa(max) + "s"

	// header
	fmt.Fprintf(buf, `type %sRepo struct {
	`+x+` sdm.Executable
`, name, "m")

	// statements
	for _, part := range stmts {
		for stmt, _ := range part {
			fmt.Fprintf(buf, `	`+x+` *sdm.Stmt`, stmt)
			buf.WriteString("\n")
		}
	}

	// footer
	buf.WriteString(`
}

`)
}

func genInit(buf *bytes.Buffer, info readstruct.Info, pkg, name string, stmts []map[string]Code) {
	typeName := name
	if info.Package != pkg {
		typeName = info.Package + "." + name
	}

	// header
	buf.WriteString(`func (r *` + name + `Repo) Init(m *sdm.Manager) (err error) {
	r.m = m
	var typ ` + typeName + "\n")

	// stmts
	for _, part := range stmts {
		for stmt, code := range part {
			fmt.Fprintf(buf, `	r.%s, err = r.m.PrepareSQL(
		typ,
		%s,
		driver.%s,
	)
	if err != nil {
		return
	}

`, stmt, code.code, code.quote)
		}
	}

	// footer
	buf.WriteString("	return\n}\n\n")
}

func genTransaction(buf *bytes.Buffer, name string) {
	fmt.Fprintf(
		buf,
		`
func (r *%sRepo) Begin() (ret *%sRepo, err error) {
	m, ok := r.m.(*sdm.Manager)
	if !ok {
		err = errors.New("[%sRepo] Already in transaction!?")
		return
	}
	x := *r
	x.m, err = m.Begin()
	ret = &x
	return
}

func (r *%sRepo) Commit() (err error) {
	m, ok := r.m.(*sdm.Tx)
	if !ok {
		err = errors.New("[%sRepo] Not in transaction!?")
		return
	}
	return m.Commit()
}

func (r *%sRepo) Rollback() (err error) {
	m, ok := r.m.(*sdm.Tx)
	if !ok {
		err = errors.New("[%sRepo] Not in transaction!?")
		return
	}
	return m.Rollback()
}
`,
		name, name,
		name,

		name,
		name,

		name,
		name,
	)
}

func genCreate(info readstruct.Info, pkg, name string) (buf *bytes.Buffer, stmts map[string]Code) {
	buf = &bytes.Buffer{}
	typeName := name
	if info.Package != pkg {
		typeName = info.Package + "." + name
	}
	buf.WriteString(`func (r *` + name + `Repo) Create` + name + `(data *` + typeName + `) (err error) {
	stmt := r.m.Stmt(r.stmtCreate)
	_, err = stmt.Exec(r.m.Val(data)...)
	return
}

`)
	stmts = map[string]Code{
		"stmtCreate": Code{
			code:  `"INSERT INTO %table% (%cols%) VALUES (%vals%)"`,
			quote: "QInsert",
		},
	}

	return
}

func genSelect(info readstruct.Info, pkg, name string, keys []string) (buf *bytes.Buffer, stmts map[string]Code) {
	stmts = make(map[string]Code)
	typeName := name
	if info.Package != pkg {
		typeName = info.Package + "." + name
	}
	buf = &bytes.Buffer{}
	for _, k := range keys {
		stmt, code := genSelectBy(buf, info, typeName, name, k)
		stmts[stmt] = code
	}

	return
}

var regSDMTag *regexp.Regexp

func init() {
	regSDMTag = regexp.MustCompile(`sdm:"([^"]+)"`)
}

func getCol(k, tags string) (col string) {
	col = k
	if tags != "" {
		// parse tag
		tagInfo := regSDMTag.FindStringSubmatch(tags)
		if len(tagInfo) == 2 {
			// found
			vals := strings.Split(tagInfo[1], ",")
			col = vals[0]
		}
	}
	return col
}

func genSelectBy(buf *bytes.Buffer, info readstruct.Info, typeName, name, k string) (stmt string, code Code) {
	var f readstruct.Field
	for _, x := range info.Fields {
		if x.Name == k {
			f = x
			break
		}
	}

	if f.Name == "" {
		log.Fatalf("Cannot find field %s in struct %s", k, name)
	}

	col := getCol(k, f.Tags)
	stmt = "stmtSelectBy" + k

	fmt.Fprintf(
		buf,
		`func (r *%sRepo) Select%sBy%s(data %s) (ret []*%s, err error) {
	stmt := r.m.Stmt(r.%s)
	ret = make([]*%s, 0)
	rows := stmt.Query(data)
	err = rows.AppendTo(&ret)
	return
}

`,
		name, name, k, f.RawType, typeName,
		stmt,
		typeName,
	)

	code = Code{
		code:  `"SELECT %cols% FROM %table% WHERE ` + col + `=?"`,
		quote: "QSelect",
	}

	return
}

func genUpdate(info readstruct.Info, pkg, name, pk string) (buf *bytes.Buffer, stmts map[string]Code) {
	typeName := name
	if info.Package != pkg {
		typeName = info.Package + "." + name
	}

	var f readstruct.Field
	for _, x := range info.Fields {
		if x.Name == pk {
			f = x
			break
		}
	}

	if f.Name == "" {
		log.Fatalf("Cannot find field %s in struct %s", pk, name)
	}

	col := getCol(pk, f.Tags)

	buf = &bytes.Buffer{}
	fmt.Fprintf(
		buf,
		`func (r *%sRepo) Update%s(data *%s) (err error) {
	stmt := r.m.Stmt(r.stmtUpdate)
	args := append(r.m.Val(data), data.%s)
	_, err = stmt.Exec(args...)
	return
}

`,
		name, name, typeName,
		pk,
	)
	stmts = map[string]Code{
		"stmtUpdate": Code{
			code:  `"UPDATE %table% SET %combined% WHERE ` + col + `=?"`,
			quote: "QUpdate",
		},
	}

	return
}

func genDelete(info readstruct.Info, pkg, name string, keys []string) (buf *bytes.Buffer, stmts map[string]Code) {
	stmts = make(map[string]Code)
	buf = &bytes.Buffer{}
	for _, k := range keys {
		stmt, code := genDeleteBy(buf, info, name, k)
		stmts[stmt] = code
	}

	return
}

func genDeleteBy(buf *bytes.Buffer, info readstruct.Info, name, k string) (stmt string, code Code) {
	var f readstruct.Field
	for _, x := range info.Fields {
		if x.Name == k {
			f = x
			break
		}
	}

	if f.Name == "" {
		log.Fatalf("Cannot find field %s in struct %s", k, name)
	}

	col := getCol(k, f.Tags)
	stmt = "stmtDeleteBy" + k

	fmt.Fprintf(
		buf,
		`func (r *%sRepo) Delete%sBy%s(data %s) (err error) {
	stmt := r.m.Stmt(r.%s)
	_, err = stmt.Exec(data)
	return
}

`,
		name, name, k, f.RawType,
		stmt,
	)

	code = Code{
		code:  `"DELETE FROM %table% WHERE ` + col + `=?"`,
		quote: "QSelect",
	}

	return
}
