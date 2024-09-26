package kibugenv2

import (
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
)

func buildWorkerController(f *jen.File, pkg *kibumod.Package) {
	f.Type().Id("WorkerController").StructFunc(func(g *jen.Group) {
		g.Id("Client").Qual(temporalClientImportName, "Client")
		g.Id("Options").Qual(temporalWorkerImportName, "Options")
		g.Id("ActivitiesController").Id("ActivitiesController")
		for _, svc := range pkg.Services {
			if svc.Decorators.Some(isKibuWorkflow) {
				g.Id(svc.Name + "WorkflowController").Id(svc.Name + "WorkflowController")
			}
		}
	})

	f.Func().Params(jen.Id("wc").Op("*").Id("WorkerController")).Id("Build").Params().Qual(temporalWorkerImportName, "Worker").Block(
		jen.Id("wk").Op(":=").Qual(temporalWorkerImportName, "New").Call(
			jen.Id("wc").Dot("Client"),
			jen.Id(packageNameConst()),
			jen.Id("wc").Dot("Options"),
		),
		jen.Id("wc").Dot("ActivitiesController").Dot("Build").Call(jen.Id("wk")),
		jen.Line(),
		jen.For(jen.List(jen.Id("_"), jen.Id("controller")).Op(":=").Range().Id("[]interface{}").Values(
			jen.ListFunc(func(g *jen.Group) {
				for _, svc := range pkg.Services {
					if svc.Decorators.Some(isKibuWorkflow) {
						g.Id("wc").Dot(svc.Name + "WorkflowController")
					}
				}
			}),
		)).Block(
			jen.Id("controller").Assert(jen.Interface(
				jen.Id("Build").Params(jen.Qual(temporalWorkerImportName, "WorkflowRegistry")),
			)).Dot("Build").Call(jen.Id("wk")),
		),
		jen.Return(jen.Id("wk")),
	)

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

	f.Var().Id("WireSet").Op("=").Qual(wireImportName, "NewSet").CallFunc(func(g *jen.Group) {
		g.Id("NewService")
		g.Id("NewActivities")
		g.Id("NewActivitiesProxy")
		g.Id("NewWorkflowsProxy")
		g.Id("NewWorkflowsClient")
		for _, svc := range pkg.Services {
			if svc.Decorators.Some(isKibuWorkflow) {
				g.Id("New" + svc.Name + "WorkflowFactory")
			}
		}
		g.Qual(wireImportName, "Struct").Call(jen.Op("new").Parens(jen.Id("WorkerController")), jen.Lit("*"))
		g.Qual(wireImportName, "Struct").Call(jen.Op("new").Parens(jen.Id("ServiceController")), jen.Lit("*"))
		g.Qual(wireImportName, "Struct").Call(jen.Op("new").Parens(jen.Id("ActivitiesController")), jen.Lit("*"))
		for _, svc := range pkg.Services {
			if svc.Decorators.Some(isKibuWorkflow) {
				g.Qual(wireImportName, "Struct").Call(jen.Op("new").Parens(jen.Id(svc.Name+"WorkflowController")), jen.Lit("*"))
			}
		}
	})
}
