package schema

import "errors"

var (
	ErrUnresolvedType  = errors.New("unresolved type reference")
	ErrUnsupportedType = errors.New("unsupported Go type")
)
