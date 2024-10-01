package middleware

import (
	"github.com/kibu-sh/kibu/pkg/transport"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func newTestMW(_ string) transport.Middleware {
	return transport.NewMiddleware(func(tctx transport.Context, next transport.Handler) error {
		return next.Serve(tctx)
	})
}

func TestNewMiddlewareRegistry(t *testing.T) {
	t.Skip("FIXME: middleware registry needs to be reinvented")
	registry := NewRegistry()
	mw1 := newTestMW("mw1")
	mw2 := newTestMW("mw2")
	mw3 := newTestMW("mw3")
	mw4 := newTestMW("mw4")
	globalMW := newTestMW("global")

	t.Logf("mw1 %p", mw1)
	t.Logf("mw2 %p", mw2)
	t.Logf("mw3 %p", mw3)
	t.Logf("mw4 %p", mw4)
	t.Logf("globalMW %p", globalMW)

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
	require.Contains(t, registry.cache, "database")
	require.Len(t, registry.cache["auth"], 3)
	require.Len(t, registry.cache["global"], 1)
	require.Len(t, registry.cache["database"], 1)

	t.Run("should include global and auth by default", func(t *testing.T) {
		results := registry.Get(GetParams{})
		require.Len(t, results, 4)
		requireMiddlewareEq(t, globalMW, results[0])
		requireMiddlewareEq(t, mw2, results[1])
		requireMiddlewareEq(t, mw1, results[2])
		requireMiddlewareEq(t, mw3, results[3])
	})

	t.Run("should not include the same auth middleware twice", func(t *testing.T) {
		results := registry.Get(GetParams{
			Tags: []string{"auth", "cache"},
		})
		require.Len(t, results, 4)
		requireMiddlewareEq(t, globalMW, results[0])
		requireMiddlewareEq(t, mw2, results[1])
		requireMiddlewareEq(t, mw1, results[2])
		requireMiddlewareEq(t, mw3, results[3])
	})

	t.Run("should include auth and global", func(t *testing.T) {
		results := registry.Get(GetParams{
			Tags: []string{"cache"},
		})
		require.Len(t, results, 4)
		requireMiddlewareEq(t, globalMW, results[0])
		requireMiddlewareEq(t, mw2, results[1])
		requireMiddlewareEq(t, mw1, results[2])
		requireMiddlewareEq(t, mw4, results[1])
	})

	t.Run("should exclude auth middleware", func(t *testing.T) {
		results := registry.Get(GetParams{
			Tags:        []string{"cache"},
			ExcludeAuth: true,
		})
		require.Len(t, results, 2)
		requireMiddlewareEq(t, globalMW, results[0])
		requireMiddlewareEq(t, mw4, results[1])
	})
}

func Test__requireMiddlewareEq(t *testing.T) {
	mw1 := newTestMW("mw1")
	mw2 := newTestMW("mw2")
	mw3 := mw1
	requireMiddlewareEq(t, mw1, mw3)
	require.True(t, middlewareEq(mw1, mw3), "middleware pointers should be equal")
	require.False(t, middlewareEq(mw1, mw2), "middleware pointers should not be equal")
}

func middlewareEq(a, b transport.Middleware) bool {
	return ptrOf(a) == ptrOf(b)
}

func ptrOf(mw transport.Middleware) uintptr {
	return reflect.ValueOf(mw).Pointer()
}

func requireMiddlewareEq(t *testing.T, a, b transport.Middleware) {
	require.Truef(t, middlewareEq(a, b), "middleware pointers should be equal expected: %p to be: %p", a, b)
}
