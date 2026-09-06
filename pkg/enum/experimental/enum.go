// Package experimental is an isolated constructor prototype for static enum
// discovery. It does not change the existing enum package's API.
package experimental

import "fmt"

type Scalar interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Member[T Scalar] struct {
	Value       T
	Label       string
	Description string
}

// Set has immutable membership. Input and output slices are copied.
// Its zero value is an empty set.
type Set[T Scalar] struct {
	members []Member[T]
	index   map[T]int
}

type DuplicateValueError[T Scalar] struct {
	Value         T
	First, Second int
}

func (e *DuplicateValueError[T]) Error() string {
	return fmt.Sprintf("duplicate enum value %v at members %d and %d", e.Value, e.First, e.Second)
}

type InvalidValueError[T Scalar] struct{ Value T }

func (e *InvalidValueError[T]) Error() string {
	return fmt.Sprintf("invalid enum value %v", e.Value)
}

// Define returns a usable runtime set from the same declaration inspected by
// kibuenum. It panics with *DuplicateValueError[T] for duplicate values: such a
// declaration is a programming error, also diagnosed statically by kibuenum.
func Define[T Scalar](members []Member[T]) Set[T] {
	s := Set[T]{members: append([]Member[T](nil), members...), index: make(map[T]int, len(members))}
	for i, m := range members {
		if first, ok := s.index[m.Value]; ok {
			panic(&DuplicateValueError[T]{Value: m.Value, First: first, Second: i})
		}
		s.index[m.Value] = i
	}
	return s
}

func (s Set[T]) Has(value T) bool { _, ok := s.index[value]; return ok }
func (s Set[T]) Get(value T) (Member[T], bool) {
	if i, ok := s.index[value]; ok {
		return s.members[i], true
	}
	return Member[T]{}, false
}
func (s Set[T]) Validate(value T) error {
	if s.Has(value) {
		return nil
	}
	return &InvalidValueError[T]{Value: value}
}
func (s Set[T]) Values() []T {
	values := make([]T, len(s.members))
	for i, m := range s.members {
		values[i] = m.Value
	}
	return values
}
func (s Set[T]) Members() []Member[T] { return append([]Member[T](nil), s.members...) }
