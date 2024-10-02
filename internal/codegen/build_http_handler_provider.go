package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/huandu/xstrings"
	"path/filepath"
	"strings"
)

func BuildHTTPHandlerProviders(opts *PipelineOptions) (err error) {
	f := opts.FileSet.Get(kibuGenWireSetPath(opts))
	f.Type().Id("HTTPHandlerFactoryDeps").StructFunc(func(g *jen.Group) {
		for _, svc := range opts.Services {
			g.Id(buildPackageScopedID(svc.PackagePath(), svc.Name)).Op("*").Qual(svc.PackagePath(), svc.Name)
		}
		g.Id("MiddlewareRegistry").Op("*").Qual(kibuTransportMiddleware, "Registry")
		return
	})

	f.Func().Id("ProvideHTTPHandlers").Params(
		jen.Id("deps").Op("*").Id("HTTPHandlerFactoryDeps"),
	).ParamsFunc(func(g *jen.Group) {
		g.Id("handlers").Index().Op("*").Qual("github.com/kibu-sh/kibu/pkg/transport/httpx", "Handler")
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

const (
	kibugenName     = "kibugen"
	kibuGenFileName = "kibu.gen.go"
)

func kibugenPackageDir(opts GenerateParams, segments ...string) string {
	joinOpts := []string{opts.OutputDir, kibugenName}
	joinOpts = append(joinOpts, segments...)
	return filepath.Join(joinOpts...)
}

func kibuGenFilePath(opts *PipelineOptions, fileName string) (FilePath, PackageName) {
	return FilePath(kibugenPackageDir(opts.GenerateParams, fileName)), kibugenName
}

func kibuGenWireSetPath(opts *PipelineOptions) (FilePath, PackageName) {
	return kibuGenFilePath(opts, kibuGenFileName)
}

func buildPackageScopedID(pkg, name string) string {
	return xstrings.ToPascalCase(strings.Replace(jen.Qual(pkg, name).GoString(), ".", "_", -1))
}
