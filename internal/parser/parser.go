package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"go/types"
	"golang.org/x/tools/go/packages"
)

func ParseDir(dir string) (pkgs map[string]*Package, err error) {
	// fs := token.NewFileSet()
	// p, err := parser.ParseDir(fs, dir, nil, parser.AllErrors|parser.ParseComments)
	// if err != nil {
	// 	return
	// }
	//
	// pkgs = make(map[string]*Package, len(p))
	// for s, a := range p {
	// 	var pkg *Package
	// 	pkg, err = defaultPackageWalker(a)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	pkgs[s] = pkg
	// }

	return
}

func defaultPackageWalker(pkg *packages.Package) (*Package, error) {
	return walkPackage(pkg,
		parseDirectives,
		buildFuncIdCache,
		collectProviders,
		collectByDefinition(
			collectServices,
		),
		// collectServices,
		// collectWorkflows,
	)
}

func collectByDefinition(mapperFuncs packageDefMapperFunc) packageMutationFunc {
	return func(p *Package) error {
		for ident, object := range p.pkg.TypesInfo.Defs {
			if err := mapperFuncs(p)(ident, object); err != nil {
				return err
			}
		}
		return nil
	}
}

func parseDirectives(p *Package) (err error) {
	for _, f := range p.pkg.Syntax {
		p.directiveCache, err = directive.FromDecls(f.Decls)
		if err != nil {
			return
		}
	}
	return
}

func buildFuncIdCache(p *Package) (err error) {
	for ident, object := range p.pkg.TypesInfo.Defs {
		if _, ok := object.(*types.Func); ok {
			p.funcIdCache[object.(*types.Func)] = ident
		}
	}
	return
}
