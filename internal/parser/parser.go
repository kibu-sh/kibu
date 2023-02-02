package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
)

func defaultPackageWalker(pkg *packages.Package, dir string) (*Package, error) {
	return walkPackage(pkg,
		parseDirectives,
		buildFuncIdCache,
		collectProviders,
		collectByDefinition(
			collectServices,
		),
		collectByDefinition(
			collectWorkers,
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
		for ident, dirList := range dirs {
			p.directiveCache[ident] = dirList
		}
	}
	return
}

func buildFuncIdCache(p *Package) (err error) {
	for ident, object := range p.GoPackage.TypesInfo.Defs {
		if _, ok := object.(*types.Func); ok {
			p.funcIdCache[object.(*types.Func)] = ident
		}
	}
	return
}
