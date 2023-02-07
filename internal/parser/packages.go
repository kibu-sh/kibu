package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"github.com/discernhq/devx/internal/parser/smap"
	"github.com/pkg/errors"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Package struct {
	Name      string
	Path      PackagePath
	GoPackage *packages.Package
	Services  smap.Map[*ast.Ident, *Service]
	Workers   smap.Map[*ast.Ident, *Worker]
	Providers smap.Map[*ast.Ident, *Provider]

	funcIdCache    smap.Map[*types.Func, *ast.Ident]
	directiveCache smap.Map[*ast.Ident, directive.List]
}

type PackagePath string

func (p PackagePath) String() string {
	return string(p)
}

func NewPackage(p *packages.Package, dir string) *Package {
	return &Package{
		Name:      p.Name,
		GoPackage: p,
		Path: PackagePath(filepath.Join(
			dir,
			strings.Replace(p.PkgPath, p.Module.Path, "", 1),
		)),
		Services:       smap.NewMap[*ast.Ident, *Service](),
		Workers:        smap.NewMap[*ast.Ident, *Worker](),
		Providers:      smap.NewMap[*ast.Ident, *Provider](),
		funcIdCache:    smap.NewMap[*types.Func, *ast.Ident](),
		directiveCache: smap.NewMap[*ast.Ident, directive.List](),
	}
}

type packageMutationFunc func(p *Package) error
type defMapperFunc func(ident *ast.Ident, object types.Object) (err error)
type packageDefMapperFunc func(p *Package) defMapperFunc

func walkPackage(
	dir string,
	p *packages.Package,
	mutationFuncs ...packageMutationFunc,
) (result *Package, err error) {
	result = NewPackage(p, dir)

	for _, mutationFunc := range mutationFuncs {
		if err = mutationFunc(result); err != nil {
			return
		}
	}

	return
}

func ExperimentalParse(dir string, patterns ...string) (pkgList smap.Map[PackagePath, *Package], err error) {
	pkgList = smap.NewMap[PackagePath, *Package]()
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

	// deterministically walk packages
	sort.Slice(loaded, func(i, j int) bool {
		return loaded[i].PkgPath < loaded[j].PkgPath
	})

	for _, pkg := range loaded {
		if pkg.Errors != nil {
			for _, e := range pkg.Errors {
				err = errors.Wrap(e, "error loading package")
			}
		}
	}

	// TODO: think about removing this
	// it makes the developer experience annoying
	// it may be required for the parser to work (how else would we know if the syntax is correct)
	if err != nil {
		return
	}

	for _, l := range loaded {
		var pkg *Package
		pkg, err = defaultPackageWalker(l, dir)
		if err != nil {
			return
		}
		pkgList.Set(pkg.Path, pkg)
	}

	return
}
