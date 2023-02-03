package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"github.com/elliotchance/orderedmap/v2"
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
		var dirs = orderedmap.NewOrderedMap[*ast.Ident, directive.List]()
		dirs, err = directive.FromDecls(f.Decls)
		if err != nil {
			return
		}
		for _, ident := range dirs.Keys() {
			dirList, _ := dirs.Get(ident)
			p.directiveCache.Set(ident, dirList)
		}
	}
	return
}

func buildFuncIdCache(p *Package) (err error) {
	for ident, object := range p.GoPackage.TypesInfo.Defs {
		if _, ok := object.(*types.Func); ok {
			p.funcIdCache.Set(object.(*types.Func), ident)
		}
	}
	return
}
