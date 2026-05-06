package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/polanski13/asyngo/schema"
)

type packagesDefinitions struct {
	fset     *token.FileSet
	files    map[string]*ast.File
	fileDirs map[*ast.File]string
	packages map[string]*ast.Package
	types    map[string]map[string]*schema.TypeDef
}

func newPackagesDefinitions() *packagesDefinitions {
	return &packagesDefinitions{
		fset:     token.NewFileSet(),
		files:    make(map[string]*ast.File),
		fileDirs: make(map[*ast.File]string),
		packages: make(map[string]*ast.Package),
		types:    make(map[string]map[string]*schema.TypeDef),
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
				matched, err := filepath.Match(exc, base)
				if err != nil {
					return fmt.Errorf("invalid exclude pattern %q: %w", exc, err)
				}
				if matched {
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
	pd.fileDirs[f] = filepath.Dir(path)

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
	for path, f := range pd.files {
		dir := filepath.Dir(path)
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
					PkgPath:  dir,
				}
				if pd.types[dir] == nil {
					pd.types[dir] = make(map[string]*schema.TypeDef)
				}
				pd.types[dir][typeSpec.Name.Name] = td
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

		pkgBase := realPkg
		if idx := strings.LastIndex(realPkg, "/"); idx >= 0 {
			pkgBase = realPkg[idx+1:]
		}

		dirs := pd.sortedDirs()
		var matches []*schema.TypeDef
		for _, dir := range dirs {
			td, ok := pd.types[dir][name]
			if !ok {
				continue
			}
			if filepath.Base(dir) != pkgBase && td.File.Name.Name != pkgBase {
				continue
			}
			matches = append(matches, td)
		}

		if len(matches) == 1 {
			return matches[0], nil
		}
		if len(matches) > 1 {
			for _, td := range matches {
				if strings.HasSuffix(filepath.ToSlash(td.PkgPath), realPkg) {
					return td, nil
				}
			}
			return matches[0], nil
		}

		return nil, fmt.Errorf("%w: %s (package %s)", schema.ErrUnresolvedType, typeName, realPkg)
	}

	if file != nil {
		if dir, ok := pd.fileDirs[file]; ok {
			if byName, ok := pd.types[dir]; ok {
				if td, ok := byName[typeName]; ok {
					return td, nil
				}
			}
		}
	}

	for _, dir := range pd.sortedDirs() {
		if td, ok := pd.types[dir][typeName]; ok {
			return td, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", schema.ErrUnresolvedType, typeName)
}

func (pd *packagesDefinitions) sortedDirs() []string {
	dirs := make([]string, 0, len(pd.types))
	for dir := range pd.types {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	return dirs
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
