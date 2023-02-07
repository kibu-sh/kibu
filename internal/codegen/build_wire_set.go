package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/discernhq/devx/internal/parser"
	"github.com/discernhq/devx/internal/parser/smap"
	"go/ast"
)

var (
	googleWire = "github.com/google/wire"
)

func BuildWireSet(opts *GeneratorOptions) (err error) {
	f := opts.FileSet.Get(devxGenWireSetPath(opts))
	f.Var().Id("WireSet").Op("=").Qual(googleWire, "NewSet").CustomFunc(multiLineParen(), func(g *jen.Group) {
		g.Id("ProvideHTTPHandlers")
		g.Id("ProvideWorkers")
		g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
			g.New(jen.Id("HTTPHandlerFactoryDeps"))
			g.Lit("*")
		})
		g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
			g.New(jen.Id("WorkerFactoryDeps"))
			g.Lit("*")
		})
		// g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
		// 	g.New(jen.Id("WorkerContainer"))
		// 	g.Lit("*")
		// })

		services := smap.NewMap[*ast.Ident, *parser.Service]()
		workers := smap.NewMap[*ast.Ident, *parser.Worker]()
		providers := smap.NewMap[*ast.Ident, *parser.Provider]()
		pkgCache := smap.NewMap[*ast.Ident, *parser.Package]()

		for _, path := range opts.PackageList {
			pkg := path.Value
			for _, ident := range pkg.Services.Iterator() {
				pkgCache.Set(ident.Key, pkg)
				services.Set(ident.Key, ident.Value)
			}
			for _, ident := range pkg.Workers.Iterator() {
				pkgCache.Set(ident.Key, pkg)
				workers.Set(ident.Key, ident.Value)
			}
			for _, ident := range pkg.Providers.Iterator() {
				pkgCache.Set(ident.Key, pkg)
				providers.Set(ident.Key, ident.Value)
			}
		}

		for _, ident := range services.Iterator() {
			svc := ident.Value
			pkg, _ := pkgCache.Get(ident.Key)
			g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
				g.New(jen.Qual(pkg.GoPackage.PkgPath, svc.Name))
				g.Lit("*")
			})
		}

		for _, ident := range workers.Iterator() {
			wrk := ident.Value
			pkg, _ := pkgCache.Get(ident.Key)
			g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
				g.New(jen.Qual(pkg.GoPackage.PkgPath, wrk.Name))
				g.Lit("*")
			})

			g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
				g.New(jen.Qual(pkg.GoPackage.PkgPath, workerProxyName(wrk)))
				g.Lit("*")
			})
		}

		for _, ident := range providers.Iterator() {
			prv := ident.Value
			pkg, _ := pkgCache.Get(ident.Key)
			switch prv.Type {
			case parser.FunctionProviderType:
				g.Qual(pkg.GoPackage.PkgPath, prv.Name)
			case parser.StructProviderType:
				g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
					g.New(jen.Qual(pkg.GoPackage.PkgPath, prv.Name))
					g.Lit("*")
				})
			}

		}
	})
	return
}
