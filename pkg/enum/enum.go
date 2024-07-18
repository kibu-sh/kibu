package enum

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrInvalidType is returned when an invalid type is used for an enum
	// This is usually caused when a type is deserialized from a network request and we want to validate its in the Set
	ErrInvalidType = errors.New("invalid enum type")
)

type ID interface {
	~string | int
}

// Set is a generic set of enums
// This enables users to create a set of enums keyed by a custom type
// This set provides functionality for adding, getting, and validating enums of that type within the set
type Set[T ID] struct {
	m *sync.Map
}

type Item[T any] struct {
	ID    T
	Label string
}

func NewItem[T ID](id T, label string) Item[T] {
	return Item[T]{ID: id, Label: label}
}

func NewSet[T ID](items ...Item[T]) Set[T] {
	set := Set[T]{m: new(sync.Map)}
	for _, item := range items {
		_ = set.Add(item.ID, item.Label)
	}
	return set
}

func NewSetFromMap[T ID](m map[T]string) Set[T] {
	set := Set[T]{m: new(sync.Map)}
	for k, v := range m {
		_ = set.Add(k, v)
	}
	return set
}

// Add adds a new enum to the set
func (s *Set[T]) Add(id T, label string) T {
	item := Item[T]{ID: id, Label: label}
	s.m.Store(id, item)
	return id
}

// Get returns the enum item by its id
func (s *Set[T]) Get(id T) (Item[T], bool) {
	val, ok := s.m.Load(id)
	if !ok {
		return Item[T]{}, false
	}

	return val.(Item[T]), ok
}

// GetOrError returns the enum item by its id or an error if the item is not found
func (s *Set[T]) GetOrError(id T) (Item[T], error) {
	if val, ok := s.m.Load(id); ok {
		return val.(Item[T]), nil
	}

	var t T
	return Item[T]{}, errors.Join(ErrInvalidType,
		fmt.Errorf("enum for type (%T) not found by id (%v)", t, id))
}

// Has returns true if the enum item exists in the set
func (s *Set[T]) Has(id T) bool {
	_, ok := s.m.Load(id)
	return ok
}

// Validate returns an error if the enum item is not found in the set
func (s *Set[T]) Validate(id T) (err error) {
	_, err = s.GetOrError(id)
	return
}
