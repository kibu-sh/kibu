package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
)

func defaultPackageWalker(pkg *packages.Package, dir string) (*Package, error) {
	return walkPackage(dir, pkg,
		parseDirectives,
		buildFuncIdCache,
		collectByDefinition(
			collectServices,
		),
		collectByDefinition(
			collectWorkers,
		),
		collectByDefinition(
			collectProviders,
		),
		collectByDefinition(
			collectMiddleware,
		),
	)
}

func collectByDefinition(mapperFuncs packageDefMapperFunc) packageMutationFunc {
	return func(p *Package) error {
		for ident, object := range p.GoPackage.TypesInfo.Defs {
			if err := mapperFuncs(p)(ident, object); err != nil {
				return err
			}
		}
		return nil
	}
}

func parseDirectives(p *Package) (err error) {
	for _, f := range p.GoPackage.Syntax {
		var dirs = make(map[*ast.Ident]directive.List)
		dirs, err = directive.FromDecls(f.Decls)
		if err != nil {
			return
		}
		for key, dirList := range dirs {
			p.directiveCache[key] = dirList
		}
	}
	return
}

func buildFuncIdCache(p *Package) (err error) {
	for ident, object := range p.GoPackage.TypesInfo.Defs {
		if f, ok := object.(*types.Func); ok {
			p.funcIdCache[f] = ident
		}
	}
	return
}
