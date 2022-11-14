package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"github.com/pkg/errors"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
)

type Package struct {
	pkg       *packages.Package
	Name      string
	Services  map[*ast.Ident]*Service
	Providers map[*ast.Ident]*Provider

	funcIdCache    map[*types.Func]*ast.Ident
	directiveCache map[*ast.Ident]directive.List
}

func NewPackage(p *packages.Package) *Package {
	return &Package{
		pkg:            p,
		Name:           p.Name,
		Services:       make(map[*ast.Ident]*Service),
		Providers:      make(map[*ast.Ident]*Provider),
		funcIdCache:    make(map[*types.Func]*ast.Ident),
		directiveCache: make(map[*ast.Ident]directive.List),
	}
}

type packageMutationFunc func(p *Package) error
type defMapperFunc func(ident *ast.Ident, object types.Object) (err error)
type packageDefMapperFunc func(p *Package) defMapperFunc

func walkPackage(
	p *packages.Package,
	mutationFuncs ...packageMutationFunc,
) (result *Package, err error) {
	result = NewPackage(p)

	for _, mutationFunc := range mutationFuncs {
		if err = mutationFunc(result); err != nil {
			return
		}
	}

	return
}

func experimentalParse(entry string) (pkgs map[string]*Package, err error) {
	config := &packages.Config{
		Dir:   entry,
		Tests: false,
		Mode: packages.NeedName | packages.NeedFiles |
			packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}

	loaded, err := packages.Load(config)
	if err != nil {
		return
	}

	for _, pkg := range loaded {
		if pkg.Errors != nil {
			for _, e := range pkg.Errors {
				err = errors.Wrap(e, "error loading package")
			}
		}

		if err != nil {
			return
		}
	}

	pkgs = make(map[string]*Package)
	for _, l := range loaded {
		var pkg *Package
		pkg, err = defaultPackageWalker(l)
		if err != nil {
			return
		}
		pkgs[pkg.Name] = pkg
	}

	return
}
