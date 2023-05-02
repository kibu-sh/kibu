package parser

import (
	"fmt"
	"github.com/discernhq/devx/internal/parser/directive"
	"go/ast"
	"go/token"
	"go/types"
	"net/http"
	"strings"
)

type Var struct {
	*types.Var
}

func (v *Var) TypePkgPath() string {
	switch t := v.Type().(type) {
	case *types.Named:
		return t.Obj().Pkg().Path()
	case *types.Pointer:
		return t.Elem().(*types.Named).Obj().Pkg().Path()
	default:
		return ""
	}
}

func (v *Var) IsStruct() bool {
	_, ok := v.Type().Underlying().(*types.Struct)
	return ok
}

func (v *Var) IsSlice() bool {
	_, ok := v.Type().Underlying().(*types.Slice)
	return ok
}

func (v *Var) TypeName() string {
	pkgPath := v.TypePkgPath()
	base := v.Type().String()
	result := strings.Replace(base, pkgPath+".", "", 1)
	return result
}

type TypeMeta struct {
	Ident   *ast.Ident
	Object  types.Object
	Package *Package
}

// Recv returns the receiver of a method, or nil if the object is not a method.
func (t TypeMeta) Recv() *types.Var {
	if sig, ok := t.Object.Type().(*types.Signature); ok {
		return sig.Recv()
	}
	return nil
}

func (t TypeMeta) RecvNamed() *types.Named {
	if rec := t.Recv(); rec != nil {
		switch n := rec.Type().(type) {
		case *types.Pointer:
			return n.Elem().(*types.Named)
		case *types.Named:
			return n
		}
	}
	return nil
}

// QualifiedName returns the qualified name of the object.
// If the object is a method, the receiver type is prepended to the name.
// MyType.Name is the qualified name.
// where MyType is a receiver of the method Name.
func (t TypeMeta) QualifiedName() string {
	name := t.Object.Name()
	if rec := t.Recv(); rec != nil {
		name = rec.Origin().Name() + "." + name
	}

	return name
}

func (t TypeMeta) ID() string {
	return t.PackagePath() + "." + t.QualifiedName()
}

func (t TypeMeta) PackagePath() string {
	path := "_"
	pkg := t.Object.Pkg()
	// pkg is nil for objects in Universe scope and possibly types
	// introduced via Eval (see also comment in object.sameId)
	if pkg != nil && pkg.Path() != "" {
		path = pkg.Path()
	}
	return path
}

func (t TypeMeta) File() *token.File {
	return t.Package.GoPackage.Fset.File(t.Ident.Pos())
}

func (t TypeMeta) Position() token.Position {
	return t.Package.GoPackage.Fset.PositionFor(t.Ident.Pos(), false)
}

func (t TypeMeta) Pos() token.Pos {
	return t.Ident.Pos()
}

func NewTypeMeta(
	ident *ast.Ident,
	obj types.Object,
	pkg *Package,
) *TypeMeta {
	return &TypeMeta{
		Ident:   ident,
		Object:  obj,
		Package: pkg,
	}
}

type Endpoint struct {
	*TypeMeta
	Name       string
	Path       string
	Raw        bool
	Tags       []string
	Methods    []string
	Request    *Var
	Response   *Var
	Directives directive.List
	Public     bool
}

type Service struct {
	*TypeMeta
	Name       string
	Directives directive.List
	Endpoints  map[*ast.Ident]*Endpoint
}

func NewService(name string, meta *TypeMeta) *Service {
	return &Service{
		Name:      name,
		TypeMeta:  meta,
		Endpoints: make(map[*ast.Ident]*Endpoint),
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

		svc := NewService(ident.Name, NewTypeMeta(ident, obj, pkg))
		svc.Directives = dirs
		svc.Endpoints, err = collectEndpoints(pkg, n)
		if err != nil {
			return
		}
		pkg.Services[ident] = svc

		return
	}
}

func collectEndpoints(pkg *Package, n *types.Named) (endpoints map[*ast.Ident]*Endpoint, err error) {
	endpoints = make(map[*ast.Ident]*Endpoint)

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

		dir, isEndpoint := dirs.Find(directive.HasKey("devx", "endpoint"))

		if !isEndpoint {
			return
		}

		tags, _ := dir.Options.GetAll("tag", []string{})

		ep := &Endpoint{
			Name:       ident.Name,
			Directives: dirs,
			Tags:       tags,
			Raw:        dir.Options.Has("raw"),
			Public:     dir.Options.Has("public"),
			TypeMeta:   NewTypeMeta(ident, pkg.GoPackage.TypesInfo.Defs[ident], pkg),
		}

		if !ep.Raw {
			sig := m.Type().(*types.Signature)
			req := sig.Params().At(1)
			res := sig.Results().At(0)
			ep.Request = &Var{Var: req}
			ep.Response = &Var{Var: res}
		}

		ep.Path, _ = dir.Options.GetOne("path", fmt.Sprintf("/%s/%s", pkg.Name, ident.Name))
		ep.Methods, _ = dir.Options.GetAll("method", []string{http.MethodGet})

		endpoints[ident] = ep
	}
	return
}
