package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/discernhq/devx/internal/parser"
	"github.com/huandu/xstrings"
	"path/filepath"
	"strings"
)

func BuildHTTPHandlerProviders(opts *GeneratorOptions) (err error) {
	f := opts.FileSet.Get(devxGenWireSetPath(opts))
	f.Type().Id("HTTPHandlerFactoryDeps").StructFunc(func(g *jen.Group) {
		for _, elem := range opts.PackageList.Iterator() {
			pkg := elem.Value
			for _, ident := range pkg.Services.Iterator() {
				svc := ident.Value
				// TODO: id might collide
				g.Id(buildPackageScopedID(pkg, svc.Name)).Op("*").Qual(pkg.GoPackage.PkgPath, svc.Name)
			}
		}
		return
	})

	f.Func().Id("ProvideHTTPHandlers").Params(
		jen.Id("deps").Op("*").Id("HTTPHandlerFactoryDeps"),
	).ParamsFunc(func(g *jen.Group) {
		g.Id("handlers").Index().Op("*").Qual("github.com/discernhq/devx/pkg/transport/httpx", "Handler")
	}).BlockFunc(func(g *jen.Group) {
		for _, elem := range opts.PackageList.Iterator() {
			pkg := elem.Value
			for _, ident := range pkg.Services.Iterator() {
				svc := ident.Value
				g.Id("handlers").Op("=").AppendFunc(func(g *jen.Group) {
					g.Id("handlers")
					g.Id("deps").Dot(buildPackageScopedID(pkg, svc.Name)).Dot("HTTPHandlerFactory").Call().Op("...")
				})
			}
		}

		g.Return()
		return
	})

	return
}

func devxGenFilePath(opts *GeneratorOptions, fileName string) (FilePath, PackageName) {
	return FilePath(filepath.Join(opts.GenerateParams.OutputDir, "devxgen", fileName)), PackageName("devxgen")
}

func devxGenWireSetPath(opts *GeneratorOptions) (FilePath, PackageName) {
	return devxGenFilePath(opts, "wire_set.gen.go")
}

func buildPackageScopedID(pkg *parser.Package, name string) string {
	return xstrings.ToCamelCase(
		strings.Join([]string{
			pkg.GoPackage.Name,
			name,
		}, "_"),
	)
}
