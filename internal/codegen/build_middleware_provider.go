package codegen

import (
	"github.com/dave/jennifer/jen"
	"go/types"
)

func getMiddlewareReceivers(opts *PipelineOptions) (receivers []*types.Named) {
	var seen = make(map[*types.Named]bool)
	for _, middleware := range opts.Middleware {
		named := middleware.RecvNamed()
		if seen[named] {
			continue
		}
		seen[named] = true
		receivers = append(receivers, named)
	}
	return
}

func BuildMiddlewareProvider(opts *PipelineOptions) (err error) {
	f := opts.FileSet.Get(kibueGenWireSetPath(opts))
	f.Type().Id("MiddlewareDeps").StructFunc(func(g *jen.Group) {
		for _, rec := range getMiddlewareReceivers(opts) {
			// middleware.Package.GoPackage.TypesInfo.TypeOf()
			pkg := rec.Obj().Pkg().Path()
			name := rec.Obj().Name()
			g.Id(buildPackageScopedID(pkg, name)).Op("*").Qual(pkg, name)
		}
	})

	f.Func().Id("ProvideMiddleware").Params(
		jen.Id("deps").Op("*").Id("MiddlewareDeps"),
	).Params(
		jen.Id("reg").Op("*").Qual(kibueTransportMiddleware, "Registry"),
	).BlockFunc(func(g *jen.Group) {
		g.Id("reg").Op("=").Qual(kibueTransportMiddleware, "NewRegistry").Call()
		for _, mw := range opts.Middleware {
			// pkg := mw.PackagePath()
			rec := mw.RecvNamed()
			recName := buildPackageScopedID(rec.Obj().Pkg().Path(), rec.Obj().Name())
			g.Id("reg").Dot("Register").Call(
				jen.Qual(kibueTransportMiddleware, "RegistryItem").CustomFunc(multiLineCurly(), func(g *jen.Group) {
					g.Id("Order").Op(":").Lit(mw.Order)
					g.Id("Tags").Op(":").Index().String().ValuesFunc(func(g *jen.Group) {
						for _, tag := range mw.Tags {
							g.Lit(tag)
						}
						return
					})
					g.Id("Middleware").Op(":").Qual(kibueTransport, "NewMiddleware").Call(
						jen.Id("deps").Dot(recName).Dot(mw.Name),
					)
				}),
			)
		}
		g.Return()
	})
	return
}
