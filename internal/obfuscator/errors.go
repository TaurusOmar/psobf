package obfuscator

import "fmt"

type TransformError struct {
	Transform string
	Position  int
	Line      int
	Message   string
	Cause     error
}

func (e *TransformError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] line %d: %s: %v", e.Transform, e.Line, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] line %d: %s", e.Transform, e.Line, e.Message)
}

func (e *TransformError) Unwrap() error {
	return e.Cause
}

type ParseError struct {
	Position int
	Line     int
	Message  string
	Source   string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d (pos %d): %s", e.Line, e.Position, e.Message)
}

type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}
