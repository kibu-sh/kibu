package parser

import (
	"fmt"
	"github.com/discernhq/devx/internal/parser/directive"
	"go/ast"
	"go/types"
	"net/http"
)

type Endpoint struct {
	Path       string
	Method     string
	Name       string
	Directives directive.List
}

type Service struct {
	Name       string
	Endpoints  map[string]*Endpoint
	Directives directive.List
}

func NewService(name string) *Service {
	return &Service{
		Name:      name,
		Endpoints: make(map[string]*Endpoint),
	}
}

func collectServices(pkg *Package) defMapperFunc {
	return func(ident *ast.Ident, obj types.Object) (err error) {
		_, ok := obj.(*types.TypeName)
		if !ok {
			return
		}

		n, ok := obj.Type().(*types.Named)
		if !ok {
			return
		}

		// TODO: collectFields
		_, ok = n.Underlying().(*types.Struct)
		if !ok {
			return
		}

		dirs, ok := pkg.directiveCache[ident]
		if !ok {
			return
		}

		// skip this struct if it doesn't have the service directive
		if !dirs.Some(directive.HasKey("devx", "service")) {
			return
		}

		svc := NewService(ident.Name)
		svc.Directives = dirs
		svc.Endpoints, err = collectEndpoints(pkg, n)
		if err != nil {
			return
		}
		pkg.Services[ident] = svc

		return
	}
}

func collectEndpoints(pkg *Package, n *types.Named) (endpoints map[string]*Endpoint, err error) {
	endpoints = make(map[string]*Endpoint)

	for i := 0; i < n.NumMethods(); i++ {
		m := n.Method(i)
		ident, ok := pkg.funcIdCache[m]
		if !ok {
			continue
		}

		dirs, ok := pkg.directiveCache[ident]
		if !ok {
			continue
		}

		if !dirs.Some(directive.HasKey("devx", "endpoint")) {
			return
		}

		ep := &Endpoint{
			Name:       ident.Name,
			Directives: dirs,
		}

		for _, d := range dirs.Filter(directive.HasKey("devx", "endpoint")) {
			ep.Path, _ = d.Options.Get("path", fmt.Sprintf("/%s", ident.Name))
			ep.Method, _ = d.Options.Get("method", http.MethodGet)
		}

		endpoints[ident.Name] = ep
	}
	return
}
