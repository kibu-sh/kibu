package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/discernhq/devx/internal/parser"
	"sort"
)

func BuildServiceHTTPHandlerFactories(
	opts *PipelineOptions,
) (err error) {
	for _, svc := range opts.Services {
		f := opts.FileSet.Get(packageScopedFilePath(svc.Package))
		svcID := jen.Id("svc")

		f.Func().Params(svcID.Clone().Id("*" + svc.Name)).
			Id("HTTPHandlerFactory").Params(
			jen.Id("middlewareReg").Op("*").Qual(devxTransportMiddleware, "Registry"),
		).
			Params(jen.Id("[]*httpx.Handler")).
			BlockFunc(buildHTTPHandlerGroup(svc, svcID))
	}
	return
}

func buildHTTPHandlerGroup(svc *parser.Service, svcID *jen.Statement) func(g *jen.Group) {
	return func(g *jen.Group) {
		g.ReturnFunc(func(g *jen.Group) {
			g.Index().Op("*").Id("httpx.Handler").CustomFunc(multiLineCurly(), func(g *jen.Group) {
				endpoints := toSlice(svc.Endpoints)
				sort.Slice(endpoints, sortByID(endpoints))

				for _, endpoint := range endpoints {
					buildHTTPHandlerStatement(g, endpoint, svcID)
				}
			})
			return
		})
		return
	}
}

func buildHTTPHandlerStatement(g *jen.Group, ep *parser.Endpoint, svcID *jen.Statement) *jen.Statement {
	return httpxNewHandler(g).CallFunc(func(g *jen.Group) {
		g.Lit(ep.Path)
		transportNewEndpoint(g, ep.Raw).Call(
			svcID.Clone().Dot(ep.Name),
		).Dot("WithMiddleware").CustomFunc(multiLineParen(), func(g *jen.Group) {
			g.Id("middlewareReg").Dot("Get").Call(
				jen.Qual(devxTransportMiddleware, "GetParams").CustomFunc(multiLineCurly(), func(g *jen.Group) {
					g.Id("ExcludeAuth").Op(":").Lit(ep.Public)
					g.Id("Tags").Op(":").Index().String().ValuesFunc(func(g *jen.Group) {
						for _, tag := range ep.Tags {
							g.Lit(tag)
						}
						return
					})
				}),
			).Op("...")
		})
	}).Dot("WithMethods").CallFunc(func(g *jen.Group) {
		for _, method := range ep.Methods {
			g.Lit(method)
		}
		return
	})
}

type qualifier interface {
	Qual(string, string) *jen.Statement
}

type literal interface {
	Lit(string) *jen.Statement
}

func transportNewEndpoint(q qualifier, raw bool) *jen.Statement {
	method := "NewEndpoint"
	if raw {
		method = "NewRawEndpoint"
	}
	return q.Qual("github.com/discernhq/devx/pkg/transport", method)
}

func httpxNewHandler(q qualifier) *jen.Statement {
	return q.Qual("github.com/discernhq/devx/pkg/transport/httpx", "NewHandler")
}
