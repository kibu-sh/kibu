package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"net/http"
	"net/url"
	"strings"

	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
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
	result := strings.Replace(base, fmt.Sprintf("%s.", pkgPath), "", 1)
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
		name = fmt.Sprintf("%s.%s", rec.Origin().Name(), name)
	}

	return name
}

func (t TypeMeta) ID() string {
	return fmt.Sprintf("%s.%s", t.PackagePath(), t.QualifiedName())
}

func (t TypeMeta) PackagePath() string {
	pkg := t.Object.Pkg()
	// pkg is nil for objects in Universe scope and possibly types
	// introduced via Eval (see also comment in object.sameId)
	return resolvePackagePath(pkg)
}

func resolvePackagePath(pkg *types.Package) string {
	if pkg != nil && pkg.Path() != "" {
		return pkg.Path()
	}
	return "_"
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
	Directives decorators.List
	Public     bool
}

type Service struct {
	*TypeMeta
	Name       string
	Directives decorators.List
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

		_, ok = n.Underlying().(*types.Struct)
		if !ok {
			return
		}

		dirs, ok := pkg.directiveCache[ident]
		if !ok {
			return
		}

		// skip this struct if it doesn't have the service directive
		if !dirs.Some(decorators.HasKey("kibu", "service")) {
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

func defaultEndpointPath(pkg *Package, ident *ast.Ident) string {
	result, _ := url.JoinPath("/", pkg.Name, ident.Name)
	return result
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

		dir, isEndpoint := dirs.Find(decorators.HasKey("kibu", "endpoint"))
		if !isEndpoint {
			continue
		}

		tags, _ := dir.Options.ListValues("tag", []string{})

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

		ep.Path = dir.Options.Lookup("path").Or(defaultEndpointPath(pkg, ident))
		ep.Methods, _ = dir.Options.ListValues("method", []string{http.MethodGet})

		endpoints[ident] = ep
	}
	return
}
