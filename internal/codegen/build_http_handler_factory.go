package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/discernhq/devx/internal/parser"
)

func BuildServiceHTTPHandlerFactories(
	opts *GeneratorOptions,
) (err error) {
	for _, elem := range opts.PackageList.Iterator() {
		pkg := elem.Value
		f := opts.FileSet.Get(packageScopedFilePath(pkg))
		for _, elem := range pkg.Services.Iterator() {
			svc := elem.Value
			svcID := jen.Id("svc")

			f.Func().Params(svcID.Clone().Id("*" + svc.Name)).
				Id("HTTPHandlerFactory").Params().
				Params(jen.Id("[]*httpx.Handler")).
				BlockFunc(buildHTTPHandlerGroup(svc, svcID))
		}
	}
	return
}

func buildHTTPHandlerGroup(svc *parser.Service, svcID *jen.Statement) func(g *jen.Group) {
	return func(g *jen.Group) {
		g.ReturnFunc(func(g *jen.Group) {
			g.Index().Op("*").Id("httpx.Handler").CustomFunc(multiLineCurly(), func(g *jen.Group) {
				for _, elm := range svc.Endpoints.Iterator() {
					buildHTTPHandlerStatement(g, elm.Value, svcID)
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
		)
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
