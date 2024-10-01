package middleware

import (
	"github.com/kibu-sh/kibu/pkg/transport"
	"sort"
)

type RegistryItem struct {
	Order      int
	Tags       []string
	Middleware transport.Middleware
}

type Registry struct {
	cache map[string][]*RegistryItem
}

// Register takes a MiddlewareSetItem and adds it to the cache for each of its tags
func (r *Registry) Register(item RegistryItem) {
	for _, tag := range item.Tags {
		r.cache[tag] = append(r.cache[tag], &item)
		sort.Slice(r.cache[tag], func(i, j int) bool {
			return r.cache[tag][i].Order > r.cache[tag][j].Order
		})
	}
}

type GetParams struct {
	Tags          []string
	ExcludeAuth   bool
	ExcludeGlobal bool
}

// Get returns a list of Middleware for the given tags
// "global" middleware are always returned as a part of the list
// "auth" middleware are always returned if a tag of "public" is not specified
func (r *Registry) Get(params GetParams) (result []transport.Middleware) {
	var tags = params.Tags

	if !params.ExcludeGlobal {
		tags = append([]string{"global"}, tags...)
	}

	if !params.ExcludeAuth {
		tags = append([]string{"auth"}, tags...)
	}

	seen := make(map[*RegistryItem]bool)
	for _, tag := range tags {
		if items, ok := r.cache[tag]; ok {
			for _, item := range items {
				if seen[item] {
					continue
				}
				seen[item] = true
				result = append(result, item.Middleware)
			}
		}
	}

	return
}

func NewRegistry() *Registry {
	return &Registry{
		cache: map[string][]*RegistryItem{
			"global": {},
			"auth":   {},
		},
	}
}
