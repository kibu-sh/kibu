package parser

import (
	"fmt"
	"go/token"
)

// PositionError wraps a sentinel error with source position context.
// errors.Is works because Unwrap returns the underlying sentinel.
type PositionError struct {
	Err error
	Pos token.Position
}

func (e *PositionError) Error() string {
	return fmt.Sprintf("%s at %s", e.Err.Error(), e.Pos.String())
}

func (e *PositionError) Unwrap() error {
	return e.Err
}

// NewPositionError wraps a sentinel error with the given source position.
func NewPositionError(err error, pos token.Position) *PositionError {
	return &PositionError{Err: err, Pos: pos}
}
