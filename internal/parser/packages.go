package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/pkg/errors"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
)

type Package struct {
	Name      string
	GoPackage *packages.Package
	Services  *orderedmap.OrderedMap[*ast.Ident, *Service]
	Workers   *orderedmap.OrderedMap[*ast.Ident, *Worker]
	Providers *orderedmap.OrderedMap[*ast.Ident, *Provider]

	funcIdCache    *orderedmap.OrderedMap[*types.Func, *ast.Ident]
	directiveCache *orderedmap.OrderedMap[*ast.Ident, directive.List]
}

type PackageList map[string]*Package

func NewPackage(p *packages.Package) *Package {
	return &Package{
		Name:           p.Name,
		GoPackage:      p,
		Services:       orderedmap.NewOrderedMap[*ast.Ident, *Service](),
		Workers:        orderedmap.NewOrderedMap[*ast.Ident, *Worker](),
		Providers:      orderedmap.NewOrderedMap[*ast.Ident, *Provider](),
		funcIdCache:    orderedmap.NewOrderedMap[*types.Func, *ast.Ident](),
		directiveCache: orderedmap.NewOrderedMap[*ast.Ident, directive.List](),
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

func ExperimentalParse(dir string, patterns ...string) (pkgList PackageList, err error) {
	pkgList = make(map[string]*Package)
	stat, err := os.Stat(dir)
	if err != nil {
		return
	}

	if !stat.IsDir() {
		err = errors.Errorf("parser entrypoint must be a directory got %s", dir)
		return
	}

	config := &packages.Config{
		Dir:   dir,
		Tests: false,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedModule |
			packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedDeps |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}

	loaded, err := packages.Load(config, patterns...)
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

	for _, l := range loaded {
		var pkg *Package
		pkg, err = defaultPackageWalker(l, dir)
		if err != nil {
			return
		}
		pkgList[pkg.GoPackage.PkgPath] = pkg
	}

	return
}
