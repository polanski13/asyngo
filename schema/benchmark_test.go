package schema

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/polanski13/asyngo/spec"
)

func BenchmarkResolveTypeName(b *testing.B) {
	src := `package test
import "time"

type Address struct {
	Street string  ` + "`json:\"street\"`" + `
	City   string  ` + "`json:\"city\"`" + `
	Zip    string  ` + "`json:\"zip\"`" + `
}

type User struct {
	Name      string    ` + "`json:\"name\" validate:\"required\"`" + `
	Email     string    ` + "`json:\"email\"`" + `
	Age       int       ` + "`json:\"age\"`" + `
	Score     float64   ` + "`json:\"score\"`" + `
	Active    bool      ` + "`json:\"active\"`" + `
	Tags      []string  ` + "`json:\"tags\"`" + `
	Address   Address   ` + "`json:\"address\"`" + `
	CreatedAt time.Time ` + "`json:\"createdAt\"`" + `
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		b.Fatal(err)
	}

	lookup := &testPackageLookup{
		files: map[string]*ast.File{"test.go": f},
		types: make(map[string]*TypeDef),
	}
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, s := range genDecl.Specs {
			ts, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}
			lookup.types[ts.Name.Name] = &TypeDef{
				File:     f,
				TypeSpec: ts,
				PkgPath:  "test",
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewResolver(lookup)
		components := make(map[string]*spec.Schema)
		_, err := r.ResolveTypeName("User", f, components)
		if err != nil {
			b.Fatal(err)
		}
	}
}
