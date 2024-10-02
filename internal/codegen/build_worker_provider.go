package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/parser"
)

var (
	workerFactoryDepsID = "WorkerFactoryDeps"
)

func BuildWorkerProviders(opts *PipelineOptions) (err error) {
	f := opts.FileSet.Get(kibuGenWireSetPath(opts))
	f.Type().Id(workerFactoryDepsID).StructFunc(func(g *jen.Group) {
		for _, wrk := range opts.Workers {
			// TODO: id might collide
			g.Id(buildPackageScopedID(wrk.PackagePath(), wrk.Name)).Op("*").Qual(wrk.PackagePath(), wrk.Name)
		}
		return
	})

	f.Func().Id("ProvideWorkers").Params(
		jen.Id("deps").Op("*").Id(workerFactoryDepsID),
	).ParamsFunc(func(g *jen.Group) {
		g.Id("workers").Index().Op("*").Qual(kibuTemporal, "Worker")
	}).BlockFunc(func(g *jen.Group) {
		for _, wrk := range opts.Workers {
			g.Id("workers").Op("=").AppendFunc(func(g *jen.Group) {
				g.Id("workers")
				g.Id("deps").Dot(buildPackageScopedID(wrk.PackagePath(), wrk.Name)).Dot("WorkerFactory").Call().Op("...")
			})
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
