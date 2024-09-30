package kibugenv2

import (
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
)

func buildWorkerController(f *jen.File, pkg *modspecv2.Package) {
	f.Comment("//kibu:provider group=temporal.WorkerFactory")
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

	f.Func().Id("NewWorkflowsClient").Params(
		jen.Id("client").Qual(temporalClientImportName, "Client"),
	).Id("WorkflowsClient").Block(
		jen.Return(jen.Op("&").Id("workflowsClient").Values(
			jen.Id("client").Op(":").Id("client"),
		)),
	)

	f.Comment("//kibu:provider")
	f.Var().Id("GenWireSet").Op("=").Qual(wireImportName, "NewSet").CustomFunc(modspecv2.MultiLineParen(), func(g *jen.Group) {
		g.Id("NewActivitiesProxy")
		g.Id("NewWorkflowsProxy")
		g.Id("NewWorkflowsClient")
		g.Qual(wireImportName, "Struct").Call(jen.Op("new").Parens(jen.Id("WorkerController")), jen.Lit("*"))
		g.Qual(wireImportName, "Struct").Call(jen.Op("new").Parens(jen.Id("ServiceController")), jen.Lit("*"))
		g.Qual(wireImportName, "Struct").Call(jen.Op("new").Parens(jen.Id("ActivitiesController")), jen.Lit("*"))
		for _, svc := range pkg.Services {
			if svc.Decorators.Some(isKibuWorkflow) {
				g.Qual(wireImportName, "Struct").Call(jen.Op("new").Parens(jen.Id(suffixController(svc.Name))), jen.Lit("*"))
			}
		}
	})
}
