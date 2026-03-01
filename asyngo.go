package asyngo

import "github.com/polanski13/asyngo/gen"

func Generate(cfg *gen.Config) error {
	g := gen.New()
	return g.Build(cfg)
}
