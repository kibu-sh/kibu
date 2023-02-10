package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/discernhq/devx/internal/parser"
)

var (
	googleWire = "github.com/google/wire"
)

func BuildWireSet(opts *PipelineOptions) (err error) {
	f := opts.FileSet.Get(devxGenWireSetPath(opts))
	f.Var().Id("WireSet").Op("=").Qual(googleWire, "NewSet").CustomFunc(multiLineParen(), func(g *jen.Group) {
		g.Id("ProvideHTTPHandlers")
		g.Id("ProvideWorkers")
		g.Id("ProvideMiddleware")
		g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
			g.New(jen.Id("HTTPHandlerFactoryDeps"))
			g.Lit("*")
		})
		g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
			g.New(jen.Id("WorkerFactoryDeps"))
			g.Lit("*")
		})
		g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
			g.New(jen.Id("MiddlewareDeps"))
			g.Lit("*")
		})

		for _, svc := range opts.Services {
			g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
				g.New(jen.Qual(svc.PackagePath(), svc.Name))
				g.Lit("*")
			})
		}

		for _, wrk := range opts.Workers {
			g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
				g.New(jen.Qual(wrk.PackagePath(), wrk.Name))
				g.Lit("*")
			})

			g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
				g.New(jen.Qual(wrk.PackagePath(), workerProxyName(wrk)))
				g.Lit("*")
			})

			if wrk.Type == parser.WorkflowType {
				g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
					g.New(jen.Qual(wrk.PackagePath(), workerClientName(wrk)))
					g.Lit("*")
				})
			}
		}

		for _, prv := range opts.Providers {
			switch prv.Type {
			case parser.FunctionProviderType:
				g.Qual(prv.PackagePath(), prv.Name)
			case parser.StructProviderType:
				g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
					g.New(jen.Qual(prv.PackagePath(), prv.Name))
					g.Lit("*")
				})
			}
		}

		for _, named := range getMiddlewareReceivers(opts) {
			g.Qual(googleWire, "Struct").CallFunc(func(g *jen.Group) {
				g.New(jen.Qual(named.Obj().Pkg().Path(), named.Obj().Name()))
				g.Lit("*")
			})
		}
	})
	return
}
