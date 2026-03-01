package gen

import (
	"fmt"
	"os"
	"strings"

	"github.com/polanski13/asyngo/parser"
	"github.com/polanski13/asyngo/spec"
)

type outputWriter func(cfg *Config, doc *spec.AsyncAPI) error

type Gen struct {
	writers map[string]outputWriter
}

func New() *Gen {
	g := &Gen{
		writers: map[string]outputWriter{
			"json": writeJSON,
			"yaml": writeYAML,
			"yml":  writeYAML,
			"go":   writeGo,
		},
	}
	return g
}

func (g *Gen) Build(cfg *Config) error {
	for _, dir := range cfg.SearchDirs {
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("search directory %q: %w", dir, err)
		}
	}

	p := parser.New(
		parser.WithSearchDirs(cfg.SearchDirs...),
		parser.WithMainFile(cfg.MainAPIFile),
		parser.WithExcludes(cfg.Excludes...),
		parser.WithStrict(cfg.Strict),
	)

	doc, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	if errs := doc.ValidateBasic(); len(errs) > 0 {
		return fmt.Errorf("validation: %w", errs[0])
	}

	if cfg.Strict {
		if errs := doc.Validate(); len(errs) > 0 {
			return fmt.Errorf("validation: %w", errs[0])
		}
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	for _, outType := range cfg.OutputTypes {
		writer, ok := g.writers[strings.ToLower(strings.TrimSpace(outType))]
		if !ok {
			continue
		}
		if err := writer(cfg, doc); err != nil {
			return fmt.Errorf("writing %s: %w", outType, err)
		}
	}

	return nil
}
