package parser

import (
	"fmt"
	"github.com/discernhq/devx/internal/parser/directive"
	"go/ast"
	"go/token"
	"go/types"
	"net/http"
)

type Var struct {
	Name string
	Type string
}

type Endpoint struct {
	Name       string
	Path       string
	Methods    []string
	Request    *Var
	Response   *Var
	Directives directive.List
}

type Service struct {
	Name       string
	Endpoints  map[string]*Endpoint
	Directives directive.List
	File       *token.File
	Position   token.Position
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

		// TODO: inject logger
		// fmt.Printf("inspecting %s\n", n.String())
		// skip this struct if it doesn't have the service directive
		if !dirs.Some(directive.HasKey("devx", "service")) {
			return
		}

		svc := NewService(ident.Name)
		svc.Directives = dirs
		svc.File = pkg.GoPackage.Fset.File(ident.Pos())
		svc.Position = svc.File.Position(ident.Pos())
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

		sig := m.Type().(*types.Signature)
		req := sig.Params().At(1)
		res := sig.Results().At(0)

		if !dirs.Some(directive.HasKey("devx", "endpoint")) {
			return
		}

		ep := &Endpoint{
			Name:       ident.Name,
			Directives: dirs,
			Request: &Var{
				Name: req.Name(),
				Type: getTypeNameWithoutPackage(pkg, req),
			},
			Response: &Var{
				Name: res.Name(),
				Type: getTypeNameWithoutPackage(pkg, res),
			},
		}

		for _, d := range dirs.Filter(directive.HasKey("devx", "endpoint")) {
			ep.Path, _ = d.Options.Find("path", fmt.Sprintf("/%s/%s", pkg.Name, ident.Name))
			ep.Methods, _ = d.Options.Filter("method", http.MethodGet)
		}

		endpoints[ident.Name] = ep
	}
	return
}
