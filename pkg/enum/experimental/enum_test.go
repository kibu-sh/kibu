package experimental

import (
	"errors"
	"reflect"
	"testing"
)

type status string

func TestSet(t *testing.T) {
	input := []Member[status]{{Value: "paid", Label: "Paid"}, {Value: "pending", Description: "Waiting"}}
	s := Define(input)
	input[0].Value = "changed"
	members := s.Members()
	members[0].Label = "changed"
	values := s.Values()
	values[0] = "changed"
	if !reflect.DeepEqual(s.Values(), []status{"paid", "pending"}) {
		t.Fatal(s.Values())
	}
	if got, ok := s.Get("paid"); !ok || got.Label != "Paid" {
		t.Fatalf("Get: %+v %v", got, ok)
	}
	if _, ok := s.Get("missing"); ok {
		t.Fatal("unexpected member")
	}
	if !s.Has("pending") || s.Has("missing") || s.Validate("paid") != nil {
		t.Fatal("membership")
	}
	var invalid *InvalidValueError[status]
	if !errors.As(s.Validate("missing"), &invalid) || invalid.Value != "missing" {
		t.Fatal("structured validation error")
	}
	var zero Set[status]
	if zero.Has("paid") || zero.Validate("paid") == nil || len(zero.Values()) != 0 {
		t.Fatal("zero set")
	}
}

func TestDuplicate(t *testing.T) {
	defer func() {
		e, ok := recover().(*DuplicateValueError[int])
		if !ok || e.Value != 7 || e.First != 0 || e.Second != 1 {
			t.Fatalf("unexpected panic: %+v", e)
		}
	}()
	Define([]Member[int]{{Value: 7}, {Value: 7}})
	t.Fatal("expected duplicate panic")
}
