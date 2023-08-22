package parser

import (
	"errors"
	"fmt"
	"github.com/discernhq/devx/internal/parser/directive"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Package struct {
	Name       string
	Path       PackagePath
	GoPackage  *packages.Package
	Services   map[*ast.Ident]*Service
	Workers    map[*ast.Ident]*Worker
	Providers  map[*ast.Ident]*Provider
	Middleware map[*ast.Ident]*Middleware

	funcIdCache    map[*types.Func]*ast.Ident
	directiveCache map[*ast.Ident]directive.List
}

type PackagePath string

func (p PackagePath) String() string {
	return string(p)
}

func NewPackage(p *packages.Package, dir string) *Package {
	return &Package{
		Name: p.Name,
		Path: PackagePath(filepath.Join(
			dir,
			strings.Replace(p.PkgPath, p.Module.Path, "", 1),
		)),
		GoPackage:      p,
		Services:       make(map[*ast.Ident]*Service),
		Workers:        make(map[*ast.Ident]*Worker),
		Providers:      make(map[*ast.Ident]*Provider),
		Middleware:     make(map[*ast.Ident]*Middleware),
		funcIdCache:    make(map[*types.Func]*ast.Ident),
		directiveCache: make(map[*ast.Ident]directive.List),
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

type ExperimentalParseOpts struct {
	Dir      string
	Patterns []string
	Logger   *slog.Logger
}

func ExperimentalParse(opts ExperimentalParseOpts) (pkgList map[PackagePath]*Package, err error) {
	pkgList = make(map[PackagePath]*Package)
	stat, err := os.Stat(opts.Dir)
	if err != nil {
		return
	}

	if !stat.IsDir() {
		err = fmt.Errorf("parser entrypoint must be a directory got %s", opts.Dir)
		return
	}

	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	config := &packages.Config{
		Dir:   opts.Dir,
		Tests: false,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedModule |
			packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedDeps |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}

	loaded, err := packages.Load(config, opts.Patterns...)
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
				// TODO: revisit this behavior
				// ignore type errors (imports)
				// parsing errors should fail the build

				if e.Kind != packages.TypeError {
					err = errors.Join(err, e)
				} else {
					opts.Logger.Warn("package type error %s: %s", pkg.PkgPath, e.Error())
				}
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
		pkg, err = defaultPackageWalker(l, opts.Dir)
		if err != nil {
			return
		}
		pkgList[pkg.Path] = pkg
	}

	return
}
