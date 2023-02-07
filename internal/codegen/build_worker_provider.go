package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/discernhq/devx/internal/parser"
)

var (
	workerFactoryDepsID = "WorkerFactoryDeps"
)

func BuildWorkerProviders(opts *GeneratorOptions) (err error) {
	f := opts.FileSet.Get(devxGenWireSetPath(opts))
	f.Type().Id(workerFactoryDepsID).StructFunc(func(g *jen.Group) {
		for _, elem := range opts.PackageList.Iterator() {
			pkg := elem.Value
			for _, ident := range pkg.Workers.Iterator() {
				wrk := ident.Value
				// TODO: id might collide
				g.Id(buildPackageScopedID(pkg, wrk.Name)).Op("*").Qual(pkg.GoPackage.PkgPath, wrk.Name)
			}
		}
		return
	})

	f.Func().Id("ProvideWorkers").Params(
		jen.Id("deps").Op("*").Id(workerFactoryDepsID),
	).ParamsFunc(func(g *jen.Group) {
		g.Id("workers").Index().Op("*").Qual(devxTemporal, "Worker")
	}).BlockFunc(func(g *jen.Group) {
		for _, elem := range opts.PackageList.Iterator() {
			pkg := elem.Value
			for _, ident := range pkg.Workers.Iterator() {
				wrk := ident.Value
				g.Id("workers").Op("=").AppendFunc(func(g *jen.Group) {
					g.Id("workers")
					g.Id("deps").Dot(buildPackageScopedID(pkg, wrk.Name)).Dot("WorkerFactory").Call().Op("...")
				})
			}
		}

		g.Return()
		return
	})

	return
}

func pluralWorkerType(t parser.WorkerType) string {
	switch t {
	case parser.WorkflowType:
		return "Workflows"
	case parser.ActivityType:
		return "Activities"
	default:
		panic("unknown worker type")
	}
}
