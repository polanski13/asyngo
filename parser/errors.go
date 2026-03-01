package parser

import (
	"errors"
	"fmt"
)

var (
	ErrMissingAnnotation    = errors.New("missing required annotation")
	ErrDuplicateOperationID = errors.New("duplicate operation ID")
	ErrInvalidAnnotation    = errors.New("invalid annotation syntax")
	ErrInvalidAction        = errors.New("invalid operation action")
	ErrMissingChannel       = errors.New("missing @Channel annotation")
)

type ParseError struct {
	File     string
	Line     int
	Function string
	Err      error
}

func (pe *ParseError) Error() string {
	if pe.Function != "" {
		return fmt.Sprintf("%s:%d (func %s): %v", pe.File, pe.Line, pe.Function, pe.Err)
	}
	return fmt.Sprintf("%s:%d: %v", pe.File, pe.Line, pe.Err)
}

func (pe *ParseError) Unwrap() error { return pe.Err }

func newParseError(file string, line int, fn string, err error) *ParseError {
	return &ParseError{
		File:     file,
		Line:     line,
		Function: fn,
		Err:      err,
	}
}
