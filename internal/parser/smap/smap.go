package smap

import "sort"

type Comparable interface {
	comparable
	String() string
}

type String string

func (s String) String() string { return string(s) }

type Map[K Comparable, V any] map[K]*Element[K, V]
type Element[K Comparable, V any] struct {
	Key   K
	Value V
}

func NewMap[K Comparable, V any]() Map[K, V] {
	return make(Map[K, V])
}

func (m Map[K, V]) Set(key K, value V) {
	m[key] = &Element[K, V]{
		Key:   key,
		Value: value,
	}
}

func (m Map[K, V]) Get(key K) (value V, ok bool) {
	if m[key] != nil {
		return m[key].Value, true
	}
	return
}

func (m Map[K, V]) GetOrDefault(key K, def V) V {
	if m[key] != nil {
		return m[key].Value
	}
	return def
}

func (m Map[K, V]) Delete(key K) {
	delete(m, key)
}

func (m Map[K, V]) Iterator() (keys []*Element[K, V]) {
	for key := range m {
		keys = append(keys, m[key])
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Key.String() < keys[j].Key.String()
	})
	return
}
