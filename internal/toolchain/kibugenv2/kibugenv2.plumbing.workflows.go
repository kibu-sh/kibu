package kibugenv2

import (
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

func buildWorkflowInterfaces(f *jen.File, pkg *kibumod.Package) {
	f.Comment("workflow interfaces")
	for _, svc := range pkg.Services {
		f.Add(workflowRunInterface(svc))
		f.Add(workflowChildRunInterface(svc))
		f.Add(workflowExternalRunInterface(svc))
		f.Add(workflowClientInterface(svc))
		f.Add(workflowChildClientInterface(svc))
	}
	return
}

func workflowRunInterface(svc *kibumod.Service) jen.Code {
	if !svc.Decorators.Some(isKibuWorkflow) {
		return jen.Null()
	}

	return jen.Type().Id(runName(svc.Name)).InterfaceFunc(func(g *jen.Group) {
		g.Id("WorkflowID").Params().Params(jen.String())
		g.Id("RunID").Params().Params(jen.String())
		g.Id("Get").Params(namedStdContextParam()).ParamsFunc(mapWorkflowExecuteResults(svc))

		signalsAndQueries := filterSignalAndQueryMethods(svc.Operations)
		lo.ForEach(signalsAndQueries, func(op *kibumod.Operation, i int) {
			g.Id(op.Name).
				ParamsFunc(mapWorkflowOperationArgsForRunIface(op)).
				ParamsFunc(mapWorkflowOperationResultsForRunIface(op))
		})

		updateMethods := filterUpdateMethods(svc.Operations)
		// generate sync methods
		lo.ForEach(updateMethods, func(op *kibumod.Operation, i int) {
			g.Id(op.Name).
				ParamsFunc(mapWorkflowOperationArgsForRunIface(op)).
				ParamsFunc(mapWorkflowOperationResultsForRunIface(op))
		})

		// generate async methods with the update handle
		lo.ForEach(updateMethods, func(op *kibumod.Operation, i int) {
			g.Id(op.Name).
				ParamsFunc(mapWorkflowOperationArgsForRunIface(op)).
				ParamsFunc(func(g *jen.Group) {
					for idx, result := range op.Results {
						if idx == 0 {
							g.Id(result.Name).Qual(kibuTemporalImportName, "UpdateHandle").
								Types(exprToJen(result.Field.Type))
						} else {
							g.Id(result.Name).Add(exprToJen(result.Field.Type))
						}
					}
				})
		})
	})
}

func workflowChildRunInterface(svc *kibumod.Service) jen.Code {
	if !svc.Decorators.Some(isKibuWorkflow) {
		return jen.Null()
	}

	return jen.Type().Id(childRunName(svc.Name)).InterfaceFunc(func(g *jen.Group) {
		g.Id("WorkflowID").Params().Params(jen.String())
		g.Id("IsReady").Params().Params(jen.Bool())
		g.Id("Underlying").Params().Params(qualWorkflowChildRunFuture())
		g.Id("Get").Params(namedWorkflowContextParam()).ParamsFunc(mapWorkflowExecuteResults(svc))
		g.Id("WaitStart").Params(namedWorkflowContextParam()).
			Params(jen.Op("*").Add(qualWorkflowExecution()), jen.Error())

		g.Id("Select").
			Params(namedWorkflowSelectorParam(),
				jen.Id("fn").Func().Params(jen.Id(childRunName(svc.Name)))).
			Params(namedWorkflowSelectorParam())

		g.Id("SelectStart").
			Params(namedWorkflowSelectorParam(),
				jen.Id("fn").Func().Params(jen.Id(childRunName(svc.Name)))).
			Params(namedWorkflowSelectorParam())

		// only signals are supported on child workflows
		// queries must happen using a client inside an activity
		// https://docs.temporal.io/docs/go/workflows#child-workflows
		signalMethods := filterSignalMethods(svc.Operations)
		lo.ForEach(signalMethods, func(op *kibumod.Operation, i int) {
			g.Id(op.Name).
				ParamsFunc(mapWorkflowOperationArgsForChildRunIface(op)).
				ParamsFunc(mapWorkflowOperationResultsForRunIface(op))
		})
	})
}

func workflowExternalRunInterface(svc *kibumod.Service) jen.Code {
	if !svc.Decorators.Some(isKibuWorkflow) {
		return jen.Null()
	}

	return jen.Type().Id(externalRunName(svc.Name)).InterfaceFunc(func(g *jen.Group) {
		g.Id("WorkflowID").Params().Params(jen.String())
		g.Id("RunID").Params().Params(jen.String())
		g.Id("RequestCancellation").Params(namedWorkflowContextParam()).Params(jen.Error())

		// only signals are supported on child workflows
		// queries must happen using a client inside an activity
		// https://docs.temporal.io/docs/go/workflows#child-workflows
		signalMethods := filterSignalMethods(svc.Operations)
		lo.ForEach(signalMethods, func(op *kibumod.Operation, i int) {
			g.Id(op.Name).
				ParamsFunc(mapWorkflowOperationArgsForChildRunIface(op)).
				Params(jen.Error())

			g.Id(nameAsync(op.Name)).
				ParamsFunc(mapWorkflowOperationArgsForChildRunIface(op)).
				Params(qualWorkflowFuture())
		})
	})
}

func workflowClientInterface(svc *kibumod.Service) jen.Code {
	if !svc.Decorators.Some(isKibuWorkflow) {
		return jen.Null()
	}

	return jen.Type().Id(clientName(svc.Name)).InterfaceFunc(func(g *jen.Group) {
		g.Id("GetHandle").
			Params(namedStdContextParam(), namedGetHandleOpts()).
			Params(jen.Id(runName(svc.Name)), jen.Error())

		executeMethod, _ := findExecuteMethod(svc)
		executeReq := paramToExpOrAny(paramAtIndex(executeMethod.Params, 1))

		g.Id("Execute").
			ParamsFunc(func(g *jen.Group) {
				g.Add(namedStdContextParam())
				g.Id("req").Add(executeReq)
				g.Id("mods").Op("...").Add(qualKibuTemporalWorkflowOptionFunc())

			}).
			ParamsFunc(func(g *jen.Group) {
				g.Id(runName(svc.Name))
				g.Error()
			})

		//// only signals are supported when starting starting workflows
		//// update with start may be coming soon
		//// https://docs.temporal.io/docs/go/workflows
		signalMethods := filterSignalMethods(svc.Operations)
		lo.ForEach(signalMethods, func(op *kibumod.Operation, i int) {
			g.Id(executeWithName(op.Name)).
				ParamsFunc(func(g *jen.Group) {
					g.Add(namedStdContextParam())

					signalReq := paramToExp(paramAtIndex(op.Params, 1))
					g.Id("req").Add(executeReq)
					g.Id("sig").Add(signalReq)
					g.Id("mods").Op("...").Add(qualKibuTemporalWorkflowOptionFunc())
				}).
				ParamsFunc(func(g *jen.Group) {
					g.Id(runName(svc.Name))
					g.Error()
				})
		})
	})
}
func workflowChildClientInterface(svc *kibumod.Service) jen.Code {
	if !svc.Decorators.Some(isKibuWorkflow) {
		return jen.Null()
	}

	return jen.Type().Id(childClientName(svc.Name)).InterfaceFunc(func(g *jen.Group) {
		g.Id("External").
			Params(namedGetHandleOpts()).
			Params(jen.Id(externalRunName(svc.Name)))

		executeMethod, _ := findExecuteMethod(svc)
		executeReq := paramToExpOrAny(paramAtIndex(executeMethod.Params, 1))
		executeRes := paramToExpOrAny(paramAtIndex(executeMethod.Results, 0))

		g.Id("Execute").
			ParamsFunc(func(g *jen.Group) {
				g.Add(namedWorkflowContextParam())
				g.Id("req").Add(executeReq)
				g.Id("mods").Op("...").Add(qualKibuTemporalWorkflowOptionFunc())
			}).
			ParamsFunc(func(g *jen.Group) {
				g.Add(executeRes)
				g.Error()
			})

		g.Id(nameAsync("Execute")).
			ParamsFunc(func(g *jen.Group) {
				g.Add(namedWorkflowContextParam())
				g.Id("req").Add(executeReq)
				g.Id("mods").Op("...").Add(qualKibuTemporalWorkflowOptionFunc())
			}).
			ParamsFunc(func(g *jen.Group) {
				g.Id(childRunName(svc.Name))
				g.Error()
			})
	})
}

func mapWorkflowOperationArgsForRunIface(op *kibumod.Operation) func(g *jen.Group) {
	return func(g *jen.Group) {
		g.Add(namedStdContextParam())
		g.Add(paramToMaybeNamedExp(paramAtIndex(op.Params, 1)))

		if op.Decorators.Some(isKibuWorkflowUpdate) {
			g.Id("mods").Op("...").Qual(kibuTemporalImportName, "UpdateOptionFunc")
		}
	}
}

func mapWorkflowOperationResultsForRunIface(op *kibumod.Operation) func(*jen.Group) {
	return func(g *jen.Group) {
		for _, result := range op.Results {
			g.Id(result.Name).Add(exprToJen(result.Field.Type))
		}
	}
}

func mapWorkflowExecuteResults(svc *kibumod.Service) func(g *jen.Group) {
	return func(g *jen.Group) {
		exec, found := lo.Find(svc.Operations, func(op *kibumod.Operation) bool {
			return op.Decorators.Some(isKibuWorkflowExecute)
		})
		if !found {
			return
		}

		for _, result := range exec.Results {
			g.Add(paramToMaybeNamedExp(mo.Some(result)))
		}
	}
}

func mapWorkflowOperationArgsForChildRunIface(op *kibumod.Operation) func(g *jen.Group) {
	return func(g *jen.Group) {
		g.Add(namedWorkflowContextParam())
		g.Add(paramToMaybeNamedExp(paramAtIndex(op.Params, 1)))
	}
}

func buildSignalChannelFuncs(f *jen.File, pkg *kibumod.Package) {
	f.Comment("signal channel providers")
	for _, svc := range pkg.Services {
		for _, op := range svc.Operations {
			if op.Decorators.Some(isKibuWorkflowSignal) {
				f.Add(signalChannelProviderFunc(svc, op))
			}
		}
	}
	return
}
