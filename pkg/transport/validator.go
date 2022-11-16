package transport

import "context"

// Validator receives a value from a Decoder and validates it.
// If an error is returned it will be passed to an ErrorEncoder.
type Validator interface {
	Validate(ctx context.Context, decoded any) error
}

// ValidatorFunc is a functional implementation to the Validator interface
type ValidatorFunc func(ctx context.Context, decoded any) error

// PayloadValidator is an alternative to Validator.
// If the return value of a Decoder implements PayloadValidator, it will be called.
// It can be used in lieu of or in tandem with Validator
type PayloadValidator interface {
	Validate() error
}
