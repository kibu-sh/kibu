package http

import "github.com/gin-gonic/gin"

type GinRouter struct {
	*gin.Engine
}

func captureGinParams(c *gin.Context) {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	ctx := ContextWithPathParams(c.Request.Context(), m)
	c.Request = c.Request.WithContext(ctx)
}

func NewGinRouter() Router {
	r := gin.Default()
	r.Use(captureGinParams)
	return GinRouter{r}
}

func (g GinRouter) MountController(controller Controller) {
	route := controller.Route()
	for _, method := range route.Methods {
		g.Handle(method, route.Path, gin.HandlerFunc(func(c *gin.Context) {
			controller.ServeHTTP(c.Writer, c.Request)
		}))
	}
}
