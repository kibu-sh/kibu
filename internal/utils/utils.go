package utils

import "sync"

type SyncMap[T any] struct {
	m *sync.Map
}

func NewSyncMap[T any]() *SyncMap[T] {
	return &SyncMap[T]{
		m: new(sync.Map),
	}
}

func (s *SyncMap[T]) Store(key string, value *T) {
	s.m.Store(key, value)
}

func (s *SyncMap[T]) Delete(key string) {
	s.m.Delete(key)
}

func (s *SyncMap[T]) Range(f func(key string, value *T) bool) {
	s.m.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*T))
	})
}

func (s *SyncMap[T]) Has(key string) (ok bool) {
	_, ok = s.Load(key)
	return
}

func (s *SyncMap[T]) Load(key string) (value *T, ok bool) {
	v, ok := s.m.Load(key)
	if !ok {
		return
	}

	value = v.(*T)
	return
}
