package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"github.com/discernhq/devx/internal/parser/smap"
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
		var dirs = smap.NewMap[*ast.Ident, directive.List]()
		dirs, err = directive.FromDecls(f.Decls)
		if err != nil {
			return
		}
		for _, elm := range dirs.Iterator() {
			dirList, _ := dirs.Get(elm.Key)
			p.directiveCache.Set(elm.Key, dirList)
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
