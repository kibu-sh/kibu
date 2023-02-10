package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/huandu/xstrings"
	"path/filepath"
	"strings"
)

func BuildHTTPHandlerProviders(opts *PipelineOptions) (err error) {
	f := opts.FileSet.Get(devxGenWireSetPath(opts))
	f.Type().Id("HTTPHandlerFactoryDeps").StructFunc(func(g *jen.Group) {
		for _, svc := range opts.Services {
			// TODO: id might collide
			g.Id(buildPackageScopedID(svc.PackagePath(), svc.Name)).Op("*").Qual(svc.PackagePath(), svc.Name)
		}
		g.Id("MiddlewareRegistry").Op("*").Qual(devxTransportMiddleware, "Registry")
		return
	})

	f.Func().Id("ProvideHTTPHandlers").Params(
		jen.Id("deps").Op("*").Id("HTTPHandlerFactoryDeps"),
	).ParamsFunc(func(g *jen.Group) {
		g.Id("handlers").Index().Op("*").Qual("github.com/discernhq/devx/pkg/transport/httpx", "Handler")
	}).BlockFunc(func(g *jen.Group) {
		for _, svc := range opts.Services {
			g.Id("handlers").Op("=").AppendFunc(func(g *jen.Group) {
				g.Id("handlers")
				g.Id("deps").Dot(buildPackageScopedID(svc.PackagePath(), svc.Name)).Dot("HTTPHandlerFactory").CustomFunc(multiLineParen(), func(g *jen.Group) {
					g.Id("deps").Dot("MiddlewareRegistry")
				}).Op("...")
			})
		}
		g.Return()
		return
	})

	return
}

func devxGenFilePath(opts *PipelineOptions, fileName string) (FilePath, PackageName) {
	return FilePath(filepath.Join(opts.GenerateParams.OutputDir, "devxgen", fileName)), PackageName("devxgen")
}

func devxGenWireSetPath(opts *PipelineOptions) (FilePath, PackageName) {
	return devxGenFilePath(opts, "wire_set.gen.go")
}

func buildPackageScopedID(pkg, name string) string {
	return xstrings.ToCamelCase(strings.Replace(jen.Qual(pkg, name).GoString(), ".", "_", -1))
}
