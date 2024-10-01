package kibugenv2

import (
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
)

func buildWorkerController(f *jen.File, pkg *modspecv2.Package) {
	f.Comment("//kibu:provider group=WorkerFactory import=github.com/kibu-sh/kibu/pkg/transport/temporal")
	f.Type().Id("WorkerController").StructFunc(func(g *jen.Group) {
		g.Id("Client").Qual(temporalClientImportName, "Client")
		g.Id("Options").Qual(temporalWorkerImportName, "Options")

		for _, svc := range pkg.Services {
			if svc.Decorators.Some(isKibuActivity) {
				g.Id(suffixController(svc.Name)).Id(suffixController(svc.Name))
			}

			if svc.Decorators.Some(isKibuWorkflow) {
				g.Id(suffixController(svc.Name)).Id(suffixController(svc.Name))
			}
		}
	})

	f.Func().Params(jen.Id("wc").Op("*").Id("WorkerController")).Id("Build").Params().Qual(temporalWorkerImportName, "Worker").
		BlockFunc(func(g *jen.Group) {
			g.Id("wk").Op(":=").Qual(temporalWorkerImportName, "New").Call(
				jen.Id("wc").Dot("Client"),
				jen.Id(packageNameConst()),
				jen.Id("wc").Dot("Options"),
			)
			for _, svc := range pkg.Services {
				if svc.Decorators.Some(decorators.OneOf(isKibuActivity, isKibuWorkflow)) {
					g.Id("wc").Dot(suffixController(svc.Name)).Dot("Build").Call(jen.Id("wk"))
				}
			}
			g.Return(jen.Id("wk"))
		})

	f.Func().Id("NewActivitiesProxy").Params().Id("ActivitiesProxy").Block(
		jen.Return(jen.Op("&").Id("activitiesProxy").Values()),
	)

	f.Func().Id("NewWorkflowsProxy").Params().Id("WorkflowsProxy").Block(
		jen.Return(jen.Op("&").Id("workflowsProxy").Values()),
	)

	f.Comment("//kibu:provider")
	f.Func().Id("NewWorkflowsClient").Params(
		jen.Id("client").Qual(temporalClientImportName, "Client"),
	).Id("WorkflowsClient").Block(
		jen.Return(jen.Op("&").Id("workflowsClient").Values(
			jen.Id("client").Op(":").Id("client"),
		)),
	)
}
