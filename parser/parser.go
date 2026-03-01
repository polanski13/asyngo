package parser

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/polanski13/asyngo/schema"
	"github.com/polanski13/asyngo/spec"
)

type Parser struct {
	spec     *spec.AsyncAPI
	packages *packagesDefinitions
	schemas  *schema.Resolver

	searchDirs      []string
	mainFile        string
	excludes        []string
	strict          bool
	operationIDs    map[string]bool
	referencedTypes []typeReference
	warnings        []string
}

type typeReference struct {
	typeName string
	file     *ast.File
}

type Option func(*Parser)

func WithSearchDirs(dirs ...string) Option {
	return func(p *Parser) {
		p.searchDirs = dirs
	}
}

func WithMainFile(file string) Option {
	return func(p *Parser) {
		p.mainFile = file
	}
}

func WithExcludes(patterns ...string) Option {
	return func(p *Parser) {
		p.excludes = patterns
	}
}

func WithStrict(strict bool) Option {
	return func(p *Parser) {
		p.strict = strict
	}
}

func New(opts ...Option) *Parser {
	p := &Parser{
		spec:         spec.NewAsyncAPI(),
		packages:     newPackagesDefinitions(),
		searchDirs:   []string{"."},
		mainFile:     "main.go",
		operationIDs: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(p)
	}
	p.schemas = schema.NewResolver(p.packages)
	return p
}

func (p *Parser) Warnings() []string {
	return p.warnings
}

func (p *Parser) Parse() (*spec.AsyncAPI, error) {
	if err := p.packages.CollectFiles(p.searchDirs, p.excludes); err != nil {
		return nil, fmt.Errorf("collecting files: %w", err)
	}

	p.packages.CatalogTypes()

	if err := p.parseGeneralAPI(); err != nil {
		return nil, fmt.Errorf("parsing general API info: %w", err)
	}

	if err := p.parseHandlers(); err != nil {
		return nil, fmt.Errorf("parsing handlers: %w", err)
	}

	if err := p.buildSchemas(); err != nil {
		return nil, fmt.Errorf("building schemas: %w", err)
	}

	return p.spec, nil
}

func (p *Parser) buildSchemas() error {
	for _, ref := range p.referencedTypes {
		schemaName := ref.typeName
		if idx := strings.LastIndex(ref.typeName, "."); idx >= 0 {
			schemaName = ref.typeName[idx+1:]
		}

		if _, exists := p.spec.Components.Schemas[schemaName]; exists {
			continue
		}

		_, err := p.schemas.ResolveTypeName(ref.typeName, ref.file, p.spec.Components.Schemas)
		if err != nil {
			return fmt.Errorf("resolving schema for %s: %w", ref.typeName, err)
		}
	}
	return nil
}

func (p *Parser) findMainFile() string {
	for path := range p.packages.Files() {
		if filepath.Base(path) == p.mainFile {
			return path
		}
	}

	for _, dir := range p.searchDirs {
		candidate := filepath.Join(strings.TrimSpace(dir), p.mainFile)
		if _, ok := p.packages.Files()[candidate]; ok {
			return candidate
		}
	}

	for path := range p.packages.Files() {
		return path
	}

	return p.mainFile
}

type funcDeclInfo struct {
	decl *ast.FuncDecl
}

func asFuncDecl(decl ast.Decl) (funcDeclInfo, bool) {
	fd, ok := decl.(*ast.FuncDecl)
	if !ok {
		return funcDeclInfo{}, false
	}
	return funcDeclInfo{decl: fd}, true
}

func (f funcDeclInfo) name() string {
	return f.decl.Name.Name
}

func (f funcDeclInfo) Doc() *ast.CommentGroup {
	return f.decl.Doc
}
