package asyngo

import (
	"errors"

	"github.com/polanski13/asyngo/gen"
)

var ErrNilConfig = errors.New("asyngo: Config is nil")

func Generate(cfg *gen.Config) error {
	if cfg == nil {
		return ErrNilConfig
	}
	g := gen.New()
	return g.Build(cfg)
}
