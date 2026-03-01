package asyngo

import "github.com/polanski13/asyngo/parser"

type ParseError = parser.ParseError

var (
	ErrMissingAnnotation    = parser.ErrMissingAnnotation
	ErrDuplicateOperationID = parser.ErrDuplicateOperationID
	ErrInvalidAnnotation    = parser.ErrInvalidAnnotation
	ErrInvalidAction        = parser.ErrInvalidAction
	ErrMissingChannel       = parser.ErrMissingChannel
)
