package parser

import (
	"fmt"
	"github.com/discernhq/devx/internal/parser/directive"
	"github.com/discernhq/devx/internal/parser/smap"
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
	Raw        bool
	Methods    []string
	Request    *Var
	Response   *Var
	Directives directive.List
}

type Service struct {
	Name       string
	Directives directive.List
	File       *token.File
	Position   token.Position
	Endpoints  smap.Map[*ast.Ident, *Endpoint]
}

func NewService(name string) *Service {
	return &Service{
		Name:      name,
		Endpoints: smap.NewMap[*ast.Ident, *Endpoint](),
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

		dirs, ok := pkg.directiveCache.Get(ident)
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
		pkg.Services.Set(ident, svc)

		return
	}
}

func collectEndpoints(pkg *Package, n *types.Named) (endpoints smap.Map[*ast.Ident, *Endpoint], err error) {
	endpoints = smap.NewMap[*ast.Ident, *Endpoint]()

	for i := 0; i < n.NumMethods(); i++ {
		m := n.Method(i)
		ident, ok := pkg.funcIdCache.Get(m)
		if !ok {
			continue
		}

		dirs, ok := pkg.directiveCache.Get(ident)
		if !ok {
			continue
		}

		dir, isEndpoint := dirs.Find(directive.HasKey("devx", "endpoint"))

		if !isEndpoint {
			return
		}

		ep := &Endpoint{
			Name:       ident.Name,
			Directives: dirs,
			Raw:        dir.Options.Has("raw"),
		}

		if !ep.Raw {
			sig := m.Type().(*types.Signature)
			req := sig.Params().At(1)
			res := sig.Results().At(0)
			ep.Request = &Var{
				Name: req.Name(),
				Type: getTypeNameWithoutPackage(pkg, req),
			}
			ep.Response = &Var{
				Name: res.Name(),
				Type: getTypeNameWithoutPackage(pkg, res),
			}
		}

		ep.Path, _ = dir.Options.GetOne("path", fmt.Sprintf("/%s/%s", pkg.Name, ident.Name))
		ep.Methods, _ = dir.Options.GetAll("method", []string{http.MethodGet})

		endpoints.Set(ident, ep)
	}
	return
}
