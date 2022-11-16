package httpx

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

var _ ServeMux = (*GinMux)(nil)

type GinMux struct {
	mux *gin.Engine
}

func (g GinMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	g.mux.ServeHTTP(writer, request)
}

func NewGinMux() *GinMux {
	r := gin.Default()
	r.Use(captureGinParams)
	return &GinMux{r}
}

func (g GinMux) Handle(handler *Handler) {
	for _, method := range handler.Methods {
		g.mux.Handle(method, handler.Path, gin.HandlerFunc(func(c *gin.Context) {
			handler.ServeHTTP(c.Writer, c.Request)
		}))
	}
}

var uriParamsContextKey struct{}

// PathParamsFromContext
func PathParamsFromContext(ctx context.Context) url.Values {
	if v, ok := ctx.Value(uriParamsContextKey).(url.Values); ok {
		return v
	}
	return nil
}

// ContextWithPathParams
func ContextWithPathParams(ctx context.Context, values url.Values) context.Context {
	return context.WithValue(ctx, uriParamsContextKey, values)
}

func captureGinParams(c *gin.Context) {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	ctx := ContextWithPathParams(c.Request.Context(), m)
	c.Request = c.Request.WithContext(ctx)
}
