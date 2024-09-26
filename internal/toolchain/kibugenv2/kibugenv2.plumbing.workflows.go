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
	f.Add(workflowsProxyInterface(pkg))
	f.Add(workflowsClientInterface(pkg))

	f.Line()
	f.Comment("workflow implementations")
	buildWorkflowsClientImplementation(f, pkg)
	buildWorkflowsProxyImplementation(f, pkg)
	return
}

func workflowsProxyInterface(pkg *kibumod.Package) jen.Code {
	return jen.Type().Id("WorkflowsProxy").InterfaceFunc(func(g *jen.Group) {
		for _, svc := range pkg.Services {
			if svc.Decorators.Some(isKibuWorkflow) {
				g.Id(svc.Name).Params().Id(childClientName(svc.Name))
			}
		}
	})
}

func workflowsClientInterface(pkg *kibumod.Package) jen.Code {
	return jen.Type().Id("WorkflowsClient").InterfaceFunc(func(g *jen.Group) {
		for _, svc := range pkg.Services {
			if svc.Decorators.Some(isKibuWorkflow) {
				g.Id(svc.Name).Params().Id(clientName(svc.Name))
			}
		}
	})
}

func buildWorkflowsClientImplementation(f *jen.File, pkg *kibumod.Package) {
	f.Type().Id("workflowsClient").Struct(
		jen.Id("client").Qual(temporalClientImportName, "Client"),
	)

	for _, svc := range pkg.Services {
		if !svc.Decorators.Some(isKibuWorkflow) {
			continue
		}

		f.Func().Params(jen.Id("w").Op("*").Id("workflowsClient")).Id(svc.Name).Params().Id(clientName(svc.Name)).Block(
			jen.Return(jen.Op("&").Id(firstToLower(clientName(svc.Name))).Values(
				jen.Id("client").Op(":").Id("w").Dot("client"),
			)),
		)

		buildWorkflowClientImplementation(f, svc)
	}
}

func buildWorkflowClientImplementation(f *jen.File, svc *kibumod.Service) {
	clientStructName := firstToLower(clientName(svc.Name))

	f.Type().Id(clientStructName).Struct(
		jen.Id("client").Qual(temporalClientImportName, "Client"),
	)

	buildExecuteMethod(f, svc)
	buildGetHandleMethod(f, svc)
	buildExecuteWithSignalMethods(f, svc)
}

func buildExecuteMethod(f *jen.File, svc *kibumod.Service) {
	clientStructName := firstToLower(clientName(svc.Name))
	executeMethod, _ := findExecuteMethod(svc)
	executeReq := paramToExpOrAny(paramAtIndex(executeMethod.Params, 1))

	f.Func().Params(jen.Id("c").Op("*").Id(clientStructName)).Id("Execute").
		ParamsFunc(func(g *jen.Group) {
			g.Add(namedStdContextParam())
			g.Id("req").Add(executeReq)
			g.Id("mods").Op("...").Add(qualKibuTemporalWorkflowOptionFunc())
		}).
		Params(jen.Id(runName(svc.Name)), jen.Error()).
		BlockFunc(func(g *jen.Group) {
			g.Id("options").Op(":=").Qual(kibuTemporalImportName, "NewWorkflowOptionsBuilder").Call().Dot("WithProvidersWhenSupported").Call(jen.Id("req")).Dot("WithOptions").Call(jen.Id("mods").Op("...")).Dot("WithTaskQueue").Call(jen.Id(packageNameConst())).Dot("AsStartOptions").Call()
			g.Line()
			g.List(jen.Id("we"), jen.Err()).Op(":=").Id("c").Dot("client").Dot("ExecuteWorkflow").Call(
				jen.Id("ctx"),
				jen.Id("options"),
				jen.Id(svcConstName(svc)),
				jen.Id("req"),
			)
			g.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Nil(), jen.Err()),
			)
			g.Line()
			g.Return(
				jen.Op("&").Id(firstToLower(runName(svc.Name))).Values(
					jen.Id("client").Op(":").Id("c").Dot("client"),
					jen.Id("workflowRun").Op(":").Id("we"),
				),
				jen.Nil(),
			)
		})
}

func buildGetHandleMethod(f *jen.File, svc *kibumod.Service) {
	clientStructName := firstToLower(clientName(svc.Name))

	f.Func().Params(jen.Id("c").Op("*").Id(clientStructName)).Id("GetHandle").
		Params(namedStdContextParam(), jen.Id("ref").Qual(kibuTemporalImportName, "GetHandleOpts")).
		Params(jen.Id(runName(svc.Name)), jen.Error()).
		Block(
			jen.Return(
				jen.Op("&").Id(firstToLower(runName(svc.Name))).Values(
					jen.Id("client").Op(":").Id("c").Dot("client"),
					jen.Id("workflowRun").Op(":").Id("c").Dot("client").Dot("GetWorkflow").Call(
						jen.Id("ctx"),
						jen.Id("ref").Dot("WorkflowID"),
						jen.Id("ref").Dot("RunID"),
					),
				),
				jen.Nil(),
			),
		)
}

func buildExecuteWithSignalMethods(f *jen.File, svc *kibumod.Service) {
	clientStructName := firstToLower(clientName(svc.Name))
	executeMethod, _ := findExecuteMethod(svc)
	executeReq := paramToExpOrAny(paramAtIndex(executeMethod.Params, 1))

	signalMethods := filterSignalMethods(svc.Operations)
	for _, op := range signalMethods {
		signalReq := paramToExp(paramAtIndex(op.Params, 1))

		f.Func().Params(jen.Id("c").Op("*").Id(clientStructName)).Id(executeWithName(op.Name)).
			ParamsFunc(func(g *jen.Group) {
				g.Add(namedStdContextParam())
				g.Id("req").Add(executeReq)
				g.Id("sig").Add(signalReq)
				g.Id("mods").Op("...").Add(qualKibuTemporalWorkflowOptionFunc())
			}).
			Params(jen.Id(runName(svc.Name)), jen.Error()).
			BlockFunc(func(g *jen.Group) {
				g.Id("options").Op(":=").Qual(kibuTemporalImportName, "NewWorkflowOptionsBuilder").Call().Dot("WithProvidersWhenSupported").Call(jen.Id("req")).Dot("WithOptions").Call(jen.Id("mods").Op("...")).Dot("WithTaskQueue").Call(jen.Id(packageNameConst())).Dot("AsStartOptions").Call()
				g.Line()
				g.List(jen.Id("run"), jen.Err()).Op(":=").Id("c").Dot("client").Dot("SignalWithStartWorkflow").Call(
					jen.Id("ctx"),
					jen.Id("options").Dot("ID"),
					jen.Id(operationConstName(svc, op)),
					jen.Id("sig"),
					jen.Id("options"),
					jen.Id(svcConstName(svc)),
					jen.Id("req"),
				)
				g.If(jen.Err().Op("!=").Nil()).Block(
					jen.Return(jen.Nil(), jen.Err()),
				)
				g.Line()
				g.Return(
					jen.Op("&").Id(firstToLower(runName(svc.Name))).Values(
						jen.Id("client").Op(":").Id("c").Dot("client"),
						jen.Id("workflowRun").Op(":").Id("run"),
					),
					jen.Nil(),
				)
			})
	}
}

func buildWorkflowsProxyImplementation(f *jen.File, pkg *kibumod.Package) {
	f.Type().Id("workflowsProxy").Struct()

	for _, svc := range pkg.Services {
		if !svc.Decorators.Some(isKibuWorkflow) {
			continue
		}

		f.Func().Params(jen.Id("w").Op("*").Id("workflowsProxy")).Id(svc.Name).Params().Id(childClientName(svc.Name)).Block(
			jen.Return(jen.Op("&").Id(firstToLower(childClientName(svc.Name))).Values()),
		)

		buildWorkflowChildClientImplementation(f, svc)
		buildWorkflowChildRunImplementation(f, svc)
		buildWorkflowExternalRunImplementation(f, svc)
	}
}

func buildWorkflowChildClientImplementation(f *jen.File, svc *kibumod.Service) {
	childClientStructName := firstToLower(childClientName(svc.Name))

	f.Type().Id(childClientStructName).Struct()

	buildChildClientExecuteMethod(f, svc)
	buildChildClientExecuteAsyncMethod(f, svc)
	buildChildClientExternalMethod(f, svc)
}

func buildChildClientExecuteMethod(f *jen.File, svc *kibumod.Service) {
	childClientStructName := firstToLower(childClientName(svc.Name))
	executeMethod, _ := findExecuteMethod(svc)
	executeReq := paramToExpOrAny(paramAtIndex(executeMethod.Params, 1))
	executeRes := paramToExpOrAny(paramAtIndex(executeMethod.Results, 0))

	f.Func().Params(jen.Id("c").Op("*").Id(childClientStructName)).Id("Execute").
		ParamsFunc(func(g *jen.Group) {
			g.Add(namedWorkflowContextParam())
			g.Id("req").Add(executeReq)
			g.Id("mods").Op("...").Add(qualKibuTemporalWorkflowOptionFunc())
		}).
		Params(executeRes, jen.Error()).
		Block(
			jen.Return(jen.Id("c").Dot("ExecuteAsync").Call(
				jen.Id("ctx"),
				jen.Id("req"),
				jen.Id("mods").Op("..."),
			).Dot("Get").Call(jen.Id("ctx"))),
		)
}

func buildChildClientExecuteAsyncMethod(f *jen.File, svc *kibumod.Service) {
	childClientStructName := firstToLower(childClientName(svc.Name))
	executeMethod, _ := findExecuteMethod(svc)
	executeReq := paramToExpOrAny(paramAtIndex(executeMethod.Params, 1))

	f.Func().Params(jen.Id("c").Op("*").Id(childClientStructName)).Id("ExecuteAsync").
		ParamsFunc(func(g *jen.Group) {
			g.Add(namedWorkflowContextParam())
			g.Id("req").Add(executeReq)
			g.Id("mods").Op("...").Add(qualKibuTemporalWorkflowOptionFunc())
		}).
		Params(jen.Id(childRunName(svc.Name))).
		BlockFunc(func(g *jen.Group) {
			g.Id("options").Op(":=").Qual(kibuTemporalImportName, "NewWorkflowOptionsBuilder").Call().
				Dot("WithProvidersWhenSupported").Call(jen.Id("req")).
				Dot("WithOptions").Call(jen.Id("mods").Op("...")).
				Dot("WithTaskQueue").Call(jen.Id(packageNameConst())).
				Dot("AsChildOptions").Call()

			g.Id("ctx").Op("=").Qual(temporalWorkflowImportName, "WithChildOptions").Call(
				jen.Id("ctx"),
				jen.Id("options"),
			)

			g.Id("childFuture").Op(":=").Qual(temporalWorkflowImportName, "ExecuteChildWorkflow").Call(
				jen.Id("ctx"),
				jen.Id(svcConstName(svc)),
				jen.Id("req"),
			)

			g.Return(jen.Op("&").Id(firstToLower(childRunName(svc.Name))).Values(
				jen.Id("childFuture").Op(":").Id("childFuture"),
			))
		})
}

func buildChildClientExternalMethod(f *jen.File, svc *kibumod.Service) {
	childClientStructName := firstToLower(childClientName(svc.Name))

	f.Func().Params(jen.Id("c").Op("*").Id(childClientStructName)).Id("External").
		Params(jen.Id("ref").Qual(kibuTemporalImportName, "GetHandleOpts")).
		Params(jen.Id(externalRunName(svc.Name))).
		Block(
			jen.Return(jen.Op("&").Id(firstToLower(externalRunName(svc.Name))).Values(
				jen.Id("workflowID").Op(":").Id("ref").Dot("WorkflowID"),
				jen.Id("runID").Op(":").Id("ref").Dot("RunID"),
			)),
		)
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

func buildWorkflowChildRunImplementation(f *jen.File, svc *kibumod.Service) {
	childRunStructName := firstToLower(childRunName(svc.Name))

	f.Type().Id(childRunStructName).Struct(
		jen.Id("childFuture").Qual(temporalWorkflowImportName, "ChildWorkflowFuture"),
	)

	// Implement methods for childRunStructName
	buildChildRunWorkflowIDMethod(f, svc)
	buildChildRunIsReadyMethod(f, svc)
	buildChildRunUnderlyingMethod(f, svc)
	buildChildRunGetMethod(f, svc)
	buildChildRunWaitStartMethod(f, svc)
	buildChildRunSelectMethod(f, svc)
	buildChildRunSelectStartMethod(f, svc)

	// Implement signal methods
	signalMethods := filterSignalMethods(svc.Operations)
	for _, op := range signalMethods {
		buildChildRunSignalMethod(f, svc, op)
	}
}

func buildWorkflowExternalRunImplementation(f *jen.File, svc *kibumod.Service) {
	externalRunStructName := firstToLower(externalRunName(svc.Name))

	f.Type().Id(externalRunStructName).Struct(
		jen.Id("workflowID").String(),
		jen.Id("runID").String(),
	)

	// Implement methods for externalRunStructName
	buildExternalRunWorkflowIDMethod(f, svc)
	buildExternalRunRunIDMethod(f, svc)
	buildExternalRunRequestCancellationMethod(f, svc)

	// Implement signal methods
	signalMethods := filterSignalMethods(svc.Operations)
	for _, op := range signalMethods {
		buildExternalRunSignalMethod(f, svc, op)
		buildExternalRunSignalAsyncMethod(f, svc, op)
	}
}

// Add implementation for child run methods (WorkflowID, IsReady, Underlying, Get, WaitStart, Select, SelectStart)
// Add implementation for external run methods (WorkflowID, RunID, RequestCancellation, signal methods)

func buildChildRunWorkflowIDMethod(f *jen.File, svc *kibumod.Service) {
	childRunStructName := firstToLower(childRunName(svc.Name))

	f.Func().Params(jen.Id("r").Op("*").Id(childRunStructName)).Id("WorkflowID").Params().Params(jen.String()).Block(
		jen.Return(jen.Id("r").Dot("childFuture").Dot("GetChildWorkflowExecution").Call().Dot("WorkflowID")),
	)
}

func buildChildRunIsReadyMethod(f *jen.File, svc *kibumod.Service) {
	childRunStructName := firstToLower(childRunName(svc.Name))

	f.Func().Params(jen.Id("r").Op("*").Id(childRunStructName)).Id("IsReady").Params().Params(jen.Bool()).Block(
		jen.Return(jen.Id("r").Dot("childFuture").Dot("IsReady").Call()),
	)
}

func buildChildRunUnderlyingMethod(f *jen.File, svc *kibumod.Service) {
	childRunStructName := firstToLower(childRunName(svc.Name))

	f.Func().Params(jen.Id("r").Op("*").Id(childRunStructName)).Id("Underlying").Params().Params(qualWorkflowChildRunFuture()).Block(
		jen.Return(jen.Id("r").Dot("childFuture")),
	)
}

func buildChildRunGetMethod(f *jen.File, svc *kibumod.Service) {
	childRunStructName := firstToLower(childRunName(svc.Name))
	executeMethod, _ := findExecuteMethod(svc)
	executeRes := paramToExpOrAny(paramAtIndex(executeMethod.Results, 0))

	f.Func().Params(jen.Id("r").Op("*").Id(childRunStructName)).Id("Get").
		Params(namedWorkflowContextParam()).
		ParamsFunc(func(g *jen.Group) {
			g.Add(executeRes)
			g.Error()
		}).
		Block(
			jen.Var().Id("result").Add(executeRes),
			jen.Err().Op(":=").Id("r").Dot("childFuture").Dot("Get").Call(jen.Id("ctx"), jen.Op("&").Id("result")),
			jen.Return(jen.Id("result"), jen.Err()),
		)
}

func buildChildRunWaitStartMethod(f *jen.File, svc *kibumod.Service) {
	childRunStructName := firstToLower(childRunName(svc.Name))

	f.Func().Params(jen.Id("r").Op("*").Id(childRunStructName)).Id("WaitStart").
		Params(namedWorkflowContextParam()).
		Params(jen.Op("*").Add(qualWorkflowExecution()), jen.Error()).
		Block(
			jen.Return(jen.Id("r").Dot("childFuture").Dot("GetChildWorkflowExecution").Call()),
		)
}

func buildChildRunSelectMethod(f *jen.File, svc *kibumod.Service) {
	childRunStructName := firstToLower(childRunName(svc.Name))

	f.Func().Params(jen.Id("r").Op("*").Id(childRunStructName)).Id("Select").
		Params(
			namedWorkflowSelectorParam(),
			jen.Id("fn").Func().Params(jen.Id(childRunName(svc.Name))),
		).
		Params(namedWorkflowSelectorParam()).
		Block(
			jen.Return(jen.Id("sel").Dot("AddFuture").Call(
				jen.Id("r").Dot("childFuture"),
				jen.Func().Params(jen.Qual(temporalWorkflowImportName, "Future")).Block(
					jen.Id("fn").Call(jen.Id("r")),
				),
			)),
		)
}

func buildChildRunSelectStartMethod(f *jen.File, svc *kibumod.Service) {
	childRunStructName := firstToLower(childRunName(svc.Name))

	f.Func().Params(jen.Id("r").Op("*").Id(childRunStructName)).Id("SelectStart").
		Params(
			namedWorkflowSelectorParam(),
			jen.Id("fn").Func().Params(jen.Id(childRunName(svc.Name))),
		).
		Params(namedWorkflowSelectorParam()).
		Block(
			jen.Return(jen.Id("sel").Dot("AddFuture").Call(
				jen.Id("r").Dot("childFuture").Dot("GetChildWorkflowExecution").Call(),
				jen.Func().Params(jen.Qual(temporalWorkflowImportName, "Future")).Block(
					jen.Id("fn").Call(jen.Id("r")),
				),
			)),
		)
}

func buildChildRunSignalMethod(f *jen.File, svc *kibumod.Service, op *kibumod.Operation) {
	childRunStructName := firstToLower(childRunName(svc.Name))
	signalReq := paramToExp(paramAtIndex(op.Params, 1))

	f.Func().Params(jen.Id("r").Op("*").Id(childRunStructName)).Id(op.Name).
		ParamsFunc(func(g *jen.Group) {
			g.Add(namedWorkflowContextParam())
			g.Id("req").Add(signalReq)
		}).
		Params(jen.Error()).
		Block(
			jen.Return(jen.Id("r").Dot("childFuture").Dot("SignalChildWorkflow").Call(
				jen.Id("ctx"),
				jen.Id(operationConstName(svc, op)),
				jen.Id("req"),
			)),
		)
}

func buildExternalRunWorkflowIDMethod(f *jen.File, svc *kibumod.Service) {
	externalRunStructName := firstToLower(externalRunName(svc.Name))

	f.Func().Params(jen.Id("r").Op("*").Id(externalRunStructName)).Id("WorkflowID").Params().Params(jen.String()).Block(
		jen.Return(jen.Id("r").Dot("workflowID")),
	)
}

func buildExternalRunRunIDMethod(f *jen.File, svc *kibumod.Service) {
	externalRunStructName := firstToLower(externalRunName(svc.Name))

	f.Func().Params(jen.Id("r").Op("*").Id(externalRunStructName)).Id("RunID").Params().Params(jen.String()).Block(
		jen.Return(jen.Id("r").Dot("runID")),
	)
}

func buildExternalRunRequestCancellationMethod(f *jen.File, svc *kibumod.Service) {
	externalRunStructName := firstToLower(externalRunName(svc.Name))

	f.Func().Params(jen.Id("r").Op("*").Id(externalRunStructName)).Id("RequestCancellation").
		Params(namedWorkflowContextParam()).
		Params(jen.Error()).
		Block(
			jen.Return(jen.Qual(temporalWorkflowImportName, "RequestCancelExternalWorkflow").Call(
				jen.Id("ctx"),
				jen.Id("r").Dot("workflowID"),
				jen.Id("r").Dot("runID"),
			)),
		)
}

func buildExternalRunSignalMethod(f *jen.File, svc *kibumod.Service, op *kibumod.Operation) {
	externalRunStructName := firstToLower(externalRunName(svc.Name))
	signalReq := paramToExp(paramAtIndex(op.Params, 1))

	f.Func().Params(jen.Id("r").Op("*").Id(externalRunStructName)).Id(op.Name).
		ParamsFunc(func(g *jen.Group) {
			g.Add(namedWorkflowContextParam())
			g.Id("req").Add(signalReq)
		}).
		Params(jen.Error()).
		Block(
			jen.Return(jen.Qual(temporalWorkflowImportName, "SignalExternalWorkflow").Call(
				jen.Id("ctx"),
				jen.Id("r").Dot("workflowID"),
				jen.Id("r").Dot("runID"),
				jen.Id(operationConstName(svc, op)),
				jen.Id("req"),
			)),
		)
}

func buildExternalRunSignalAsyncMethod(f *jen.File, svc *kibumod.Service, op *kibumod.Operation) {
	externalRunStructName := firstToLower(externalRunName(svc.Name))
	signalReq := paramToExp(paramAtIndex(op.Params, 1))

	f.Func().Params(jen.Id("r").Op("*").Id(externalRunStructName)).Id(nameAsync(op.Name)).
		ParamsFunc(func(g *jen.Group) {
			g.Add(namedWorkflowContextParam())
			g.Id("req").Add(signalReq)
		}).
		Params(qualWorkflowFuture()).
		Block(
			jen.Return(jen.Qual(temporalWorkflowImportName, "SignalExternalWorkflow").Call(
				jen.Id("ctx"),
				jen.Id("r").Dot("workflowID"),
				jen.Id("r").Dot("runID"),
				jen.Id(operationConstName(svc, op)),
				jen.Id("req"),
			)),
		)
}
