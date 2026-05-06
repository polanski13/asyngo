package gen

import (
	"errors"
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
		return fmt.Errorf("validation: %w", errors.Join(errs...))
	}

	if cfg.Strict {
		if errs := doc.Validate(); len(errs) > 0 {
			return fmt.Errorf("validation: %w", errors.Join(errs...))
		}
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	outputTypes := normalizeOutputTypes(cfg.OutputTypes)

	for _, outType := range outputTypes {
		writer, ok := g.writers[outType]
		if !ok {
			continue
		}
		if err := writer(cfg, doc); err != nil {
			return fmt.Errorf("writing %s: %w", outType, err)
		}
	}

	return nil
}

func normalizeOutputTypes(types []string) []string {
	seen := make(map[string]bool, len(types))
	out := make([]string, 0, len(types))
	hasGo := false
	for _, t := range types {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		if t == "go" {
			hasGo = true
			continue
		}
		out = append(out, t)
	}
	if hasGo && !seen["json"] {
		out = append(out, "json")
	}
	if hasGo {
		out = append(out, "go")
	}
	return out
}
