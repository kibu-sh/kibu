package middleware

import (
	"github.com/discernhq/devx/pkg/transport"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func newTestMW(example string) transport.Middleware {
	return transport.NewMiddleware(func(ctx transport.Context, next transport.Handler) error {
		return next.Serve(ctx)
	})
}

func TestNewMiddlewareRegistry(t *testing.T) {
	registry := NewRegistry()
	mw1 := newTestMW("mw1")
	mw2 := newTestMW("mw2")
	mw3 := newTestMW("mw3")
	mw4 := newTestMW("mw4")
	globalMW := newTestMW("global")

	registry.Register(RegistryItem{
		Tags:       []string{"auth"},
		Middleware: mw2,
	})

	registry.Register(RegistryItem{
		Tags:       []string{"auth"},
		Middleware: mw1,
	})

	registry.Register(RegistryItem{
		Tags:       []string{"auth", "cache"},
		Middleware: mw3,
	})

	registry.Register(RegistryItem{
		Tags:       []string{"global"},
		Middleware: globalMW,
	})

	registry.Register(RegistryItem{
		Tags:       []string{"database"},
		Middleware: mw4,
	})

	require.Contains(t, registry.cache, "auth")
	require.Contains(t, registry.cache, "global")
	require.Len(t, registry.cache["auth"], 3)
	require.Len(t, registry.cache["global"], 1)

	t.Run("should include global and auth by default", func(t *testing.T) {
		results := registry.Get(GetParams{})
		require.Len(t, results, 4)
		requireMiddlewareEq(t, results[0], globalMW)
		requireMiddlewareEq(t, results[1], mw2)
		requireMiddlewareEq(t, results[2], mw1)
		requireMiddlewareEq(t, results[3], mw3)
	})

	t.Run("should not include the same auth middleware twice", func(t *testing.T) {
		results := registry.Get(GetParams{
			Tags: []string{"auth", "cache"},
		})
		require.Len(t, results, 4)
		requireMiddlewareEq(t, results[0], globalMW)
		requireMiddlewareEq(t, results[1], mw2)
		requireMiddlewareEq(t, results[2], mw1)
		requireMiddlewareEq(t, results[3], mw3)
	})

	t.Run("should include auth and global", func(t *testing.T) {
		results := registry.Get(GetParams{
			Tags: []string{"cache"},
		})
		require.Len(t, results, 4)
		requireMiddlewareEq(t, results[0], globalMW)
		requireMiddlewareEq(t, results[1], mw2)
		requireMiddlewareEq(t, results[2], mw1)
		requireMiddlewareEq(t, results[1], mw4)
	})

	t.Run("should exclude auth middleware", func(t *testing.T) {
		results := registry.Get(GetParams{
			Tags:        []string{"cache"},
			ExcludeAuth: true,
		})
		require.Len(t, results, 2)
		requireMiddlewareEq(t, results[0], globalMW)
		requireMiddlewareEq(t, results[1], mw4)
	})
}

// TODO: fixme, this doesn't prove anything
func requireMiddlewareEq(t *testing.T, a, b transport.Middleware) {
	require.Equal(t, reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer())
}
