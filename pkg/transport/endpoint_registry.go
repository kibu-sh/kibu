package transport

import (
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Operation interface {
	ID() string
	Register(Service)
	WithHandler(Handler)
}

type Service interface {
	ID() string
	Register(Registry)
	WithOperations(...Operation)
}

type Registry interface {
	Register(Service)
	GetByID(string) (Service, bool)
	Filter(func(Service) bool) []Service
}

type registry struct {
	cache *orderedmap.OrderedMap[string, Service]
}

func (e *registry) Register(svc Service) {
	e.cache.Set(svc.ID(), svc)
}

func (e *registry) GetByID(id string) (Service, bool) {
	return e.cache.Get(id)
}

func (e *registry) Filter(filter func(Service) bool) (result []Service) {
	for svc := range e.cache.ValuesFromOldest() {
		if filter(svc) {
			result = append(result, svc)
		}
	}
	return
}
