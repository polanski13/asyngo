package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/polanski13/asyngo/schema"
)

type packagesDefinitions struct {
	fset     *token.FileSet
	files    map[string]*ast.File
	packages map[string]*ast.Package
	types    map[string]*schema.TypeDef
}

func newPackagesDefinitions() *packagesDefinitions {
	return &packagesDefinitions{
		fset:     token.NewFileSet(),
		files:    make(map[string]*ast.File),
		packages: make(map[string]*ast.Package),
		types:    make(map[string]*schema.TypeDef),
	}
}

func (pd *packagesDefinitions) CollectFiles(dirs []string, excludes []string) error {
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if err := pd.walkDir(dir, excludes); err != nil {
			return fmt.Errorf("collecting files in %s: %w", dir, err)
		}
	}
	return nil
}

func (pd *packagesDefinitions) walkDir(dir string, excludes []string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == ".git" || base == "node_modules" {
				return filepath.SkipDir
			}
			for _, exc := range excludes {
				if matched, _ := filepath.Match(exc, base); matched {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return pd.parseFile(path)
	})
}

func (pd *packagesDefinitions) parseFile(path string) error {
	f, err := parser.ParseFile(pd.fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}
	pd.files[path] = f

	pkgName := f.Name.Name
	if _, ok := pd.packages[pkgName]; !ok {
		pd.packages[pkgName] = &ast.Package{
			Name:  pkgName,
			Files: make(map[string]*ast.File),
		}
	}
	pd.packages[pkgName].Files[path] = f

	return nil
}

func (pd *packagesDefinitions) CatalogTypes() {
	for _, f := range pd.files {
		pkgName := f.Name.Name
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, s := range genDecl.Specs {
				typeSpec, ok := s.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if typeSpec.Doc == nil && genDecl.Doc != nil {
					typeSpec.Doc = genDecl.Doc
				}
				td := &schema.TypeDef{
					File:     f,
					TypeSpec: typeSpec,
					PkgPath:  pkgName,
				}
				pd.types[pkgName+"."+typeSpec.Name.Name] = td
				pd.types[typeSpec.Name.Name] = td
			}
		}
	}
}

func (pd *packagesDefinitions) FindTypeSpec(typeName string, file *ast.File) (*schema.TypeDef, error) {
	if strings.Contains(typeName, ".") {
		parts := strings.SplitN(typeName, ".", 2)
		pkgAlias := parts[0]
		name := parts[1]

		realPkg := resolveImportAlias(file, pkgAlias)
		if realPkg == "" {
			realPkg = pkgAlias
		}

		pkgName := realPkg
		if idx := strings.LastIndex(realPkg, "/"); idx >= 0 {
			pkgName = realPkg[idx+1:]
		}

		if td, ok := pd.types[pkgName+"."+name]; ok {
			return td, nil
		}

		if td, ok := pd.types[name]; ok {
			return td, nil
		}

		return nil, fmt.Errorf("%w: %s (package %s)", schema.ErrUnresolvedType, typeName, realPkg)
	}

	if file != nil {
		currentPkg := file.Name.Name
		if td, ok := pd.types[currentPkg+"."+typeName]; ok {
			return td, nil
		}
	}

	if td, ok := pd.types[typeName]; ok {
		return td, nil
	}

	return nil, fmt.Errorf("%w: %s", schema.ErrUnresolvedType, typeName)
}

func resolveImportAlias(file *ast.File, alias string) string {
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if imp.Name != nil {
			if imp.Name.Name == alias {
				return path
			}
		} else {
			parts := strings.Split(path, "/")
			if parts[len(parts)-1] == alias {
				return path
			}
		}
	}
	return ""
}

func (pd *packagesDefinitions) FileSet() *token.FileSet {
	return pd.fset
}

func (pd *packagesDefinitions) Files() map[string]*ast.File {
	return pd.files
}

func (pd *packagesDefinitions) GetFile(path string) *ast.File {
	return pd.files[path]
}
