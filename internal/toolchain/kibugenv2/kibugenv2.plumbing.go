package kibugenv2

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/parser/directive"
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"go/ast"
	"go/types"
	"unicode"
)

const (
	kibuPrefix          = "kibu"
	workflowName        = "workflow"
	workflowExecuteName = "execute"
	workflowUpdateName  = "update"
	workflowQueryName   = "query"
	workflowSignalName  = "signal"
	activityName        = "activity"
	serviceName         = "service"

	ctxImportName              = "context"
	wireImportName             = "github.com/google/wire"
	kibuTransportImportName    = "github.com/kibu-sh/kibu/pkg/transport"
	kibuTemporalImportName     = "github.com/kibu-sh/kibu/pkg/transport/temporal"
	kibuHttpxImportName        = "github.com/kibu-sh/kibu/pkg/transport/httpx"
	temporalActivityImportName = "go.temporal.io/sdk/activity"
	temporalClientImportName   = "go.temporal.io/sdk/client"
	temporalWorkerImportName   = "go.temporal.io/sdk/worker"
	temporalWorkflowImportName = "go.temporal.io/sdk/workflow"
	timeImportName             = "time"
)

var (
	isKibuService         = directive.HasKey(kibuPrefix, serviceName)
	isKibuActivity        = directive.HasKey(kibuPrefix, activityName)
	isKibuWorkflow        = directive.HasKey(kibuPrefix, workflowName)
	isKibuWorkflowExecute = directive.HasKey(kibuPrefix, workflowName, workflowExecuteName)
	isKibuWorkflowUpdate  = directive.HasKey(kibuPrefix, workflowName, workflowUpdateName)
	isKibuWorkflowQuery   = directive.HasKey(kibuPrefix, workflowName, workflowQueryName)
	isKibuWorkflowSignal  = directive.HasKey(kibuPrefix, workflowName, workflowSignalName)
	isActivityOrWorkflow  = directive.OneOf(isKibuWorkflow, isKibuActivity)
)

func WithGenExt(name string) string {
	return fmt.Sprintf("%s.gen.go",
		lo.SnakeCase(name),
	)
}

func firstToUpper(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func newGenFile(pkg *types.Package) *jen.File {
	f := jen.NewFilePath(pkg.Path())
	f.HeaderComment("Code generated by kibugenv2. DO NOT EDIT.")
	return f
}

func proxyName(name string) string {
	return firstToUpper(fmt.Sprintf("%sProxy", name))
}

func runName(name string) string {
	return firstToUpper(fmt.Sprintf("%sRun", name))
}

func childRunName(name string) string {
	return firstToUpper(fmt.Sprintf("%sChildRun", name))
}

func externalRunName(name string) string {
	return firstToUpper(fmt.Sprintf("%sExternalRun", name))
}

func clientName(name string) string {
	return firstToUpper(fmt.Sprintf("%sClient", name))
}

func constName(name string) string {
	return fmt.Sprintf("%sName", name)
}

func privateConstName(name string) string {
	return firstToLower(constName(name))
}

func packageNameConst() string {
	return constName("package")
}

// buildPkgCompilerAssertions creates compiler assertions for all services
func buildPkgCompilerAssertions(f *jen.File, pkg *kibumod.Package) {
	f.Comment("compiler assertions")
	for _, svc := range pkg.Services {
		f.Add(matchingCompilerAssertion(svc.Name))

		if svc.Decorators.Some(isKibuActivity) {
			f.Add(compilerAssertionToInterface(
				proxyName(svc.Name), firstToLower(proxyName(svc.Name))))
		}

		if svc.Decorators.Some(isKibuWorkflow) {
			f.Add(compilerAssertionToInterface(
				runName(svc.Name), firstToLower(runName(svc.Name))))

			f.Add(compilerAssertionToInterface(
				childRunName(svc.Name), firstToLower(childRunName(svc.Name))))

			f.Add(compilerAssertionToInterface(
				clientName(svc.Name), firstToLower(clientName(svc.Name))))
		}
	}
	return
}

func svcConstName(svc *kibumod.Service) string {
	return privateConstName(svc.Name)
}

func operationConstName(svc *kibumod.Service, op *kibumod.Operation) string {
	return privateConstName(fmt.Sprintf("%s%s", svc.Name, firstToUpper(op.Name)))
}

func svcConstLiteral(pkg *kibumod.Package, svc *kibumod.Service) string {
	return fmt.Sprintf("%s.%s", pkg.Name, svc.Name)
}

func operationConstLiteral(pkg *kibumod.Package, svc *kibumod.Service, op *kibumod.Operation) string {
	return fmt.Sprintf("%s.%s.%s", pkg.Name, svc.Name, op.Name)
}

// buildPkgConstants a set of constant references for later code
func buildPkgConstants(f *jen.File, pkg *kibumod.Package) {
	f.Comment("system constants")
	f.Const().DefsFunc(func(g *jen.Group) {
		// packageName = "billingv1"
		// in the future this might want to be fully qualified
		g.Id(packageNameConst()).Op("=").Lit(pkg.Name)

		// service constants
		for _, svc := range pkg.Services {
			// serviceName = "package.service"
			g.Id(svcConstName(svc)).Op("=").
				Lit(svcConstLiteral(pkg, svc))

			for _, op := range svc.Operations {
				// serviceOperationName = "package.service.operation"
				g.Id(operationConstName(svc, op)).Op("=").
					Lit(operationConstLiteral(pkg, svc, op))
			}
		}
	})

	return
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

func buildWorkflowInterfaces(f *jen.File, pkg *kibumod.Package) {
	f.Comment("workflow interfaces")
	for _, svc := range pkg.Services {
		f.Add(workflowRunInterface(svc))
		f.Add(workflowChildRunInterface(svc))
		f.Add(workflowExternalRunInterface(svc))
		//f.Add(workflowClientInterface(svc))
		//f.Add(workflowChildClientInterface(svc))
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

	return jen.Type().Id(childRunName(svc.Name)).InterfaceFunc(func(g *jen.Group) {
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

func qualWorkflowExecution() jen.Code {
	return jen.Qual(temporalWorkflowImportName, "Execution")
}

func filterSignalMethods(operations []*kibumod.Operation) []*kibumod.Operation {
	return lo.Filter(operations, func(op *kibumod.Operation, _ int) bool {
		return op.Decorators.Some(isKibuWorkflowSignal)
	})
}

func qualWorkflowChildRunFuture() jen.Code {
	return jen.Qual(temporalWorkflowImportName, "ChildWorkflowFuture")
}

func qualWorkflowFuture() jen.Code {
	return jen.Qual(temporalWorkflowImportName, "Future")
}

func filterUpdateMethods(operations []*kibumod.Operation) []*kibumod.Operation {
	return lo.Filter(operations, func(op *kibumod.Operation, _ int) bool {
		return op.Decorators.Some(isKibuWorkflowUpdate)
	})
}

func filterSignalAndQueryMethods(operations []*kibumod.Operation) []*kibumod.Operation {
	return lo.Filter(operations, func(op *kibumod.Operation, _ int) bool {
		return op.Decorators.Some(directive.OneOf(
			isKibuWorkflowSignal,
			isKibuWorkflowQuery,
		))
	})
}

func nameAsync(name string) string {
	return firstToUpper(fmt.Sprintf("%sAsync", name))
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

func namedContextParam(importName string) jen.Code {
	return jen.Id("ctx").Qual(importName, "Context")
}
func namedStdContextParam() jen.Code {
	return namedContextParam(ctxImportName)
}

func namedWorkflowContextParam() jen.Code {
	return namedContextParam(temporalWorkflowImportName)
}

func namedWorkflowSelectorParam() jen.Code {
	return jen.Id("sel").Qual(temporalWorkflowImportName, "Selector")
}

func signalChannelProviderFunc(svc *kibumod.Service, op *kibumod.Operation) jen.Code {
	return jen.Func().Id(signalChannelProviderFuncName(svc, op)).
		Params(namedWorkflowContextParam()).
		ParamsFunc(func(g *jen.Group) {
			g.Qual(kibuTemporalImportName, "SignalChannel").
				Types(paramToExp(paramAtIndex(op.Params, 1)))
		}).
		BlockFunc(func(g *jen.Group) {
			g.ReturnFunc(func(g *jen.Group) {
				g.Qual(kibuTemporalImportName, "NewSignalChannel").
					Types(paramToExp(paramAtIndex(op.Params, 1))).
					Call(jen.Id("ctx"), jen.Id(operationConstName(svc, op)))
			})
		})
}

type optionalParam = mo.Option[kibumod.Type]
type optionCode = mo.Option[jen.Code]

func paramAtIndex(params []kibumod.Type, index int) optionalParam {
	if index < 0 || index >= len(params) {
		return mo.None[kibumod.Type]()
	}

	return mo.Some[kibumod.Type](params[index])
}

func paramToExp(param optionalParam) jen.Code {
	if param.IsAbsent() {
		return jen.Null()
	}
	return exprToJen(param.MustGet().Field.Type)
}

func paramToMaybeNamedExp(param optionalParam) jen.Code {
	if param.IsAbsent() {
		return jen.Null()
	}

	p := param.MustGet()
	exp := paramToExp(param)
	if p.Name == "" {
		return exp
	}

	return jen.Id(p.Name).Add(exp)
}

func exprToJen(expr ast.Expr) jen.Code {
	switch e := expr.(type) {
	case *ast.Ident:
		// Simple identifier
		return jen.Id(e.Name)
	case *ast.SelectorExpr:
		// Qualified identifier (e.g., pkg.Type)
		xIdent, ok := e.X.(*ast.Ident)
		if ok {
			return jen.Qual(xIdent.Name, e.Sel.Name)
		}
		// Handle other cases as needed
	case *ast.StarExpr:
		// Pointer type
		return jen.Op("*").Add(exprToJen(e.X))
	case *ast.ArrayType:
		// Array or slice type
		if e.Len != nil {
			return jen.Index(exprToJen(e.Len)).Add(exprToJen(e.Elt))
		}
		return jen.Index().Add(exprToJen(e.Elt))
	case *ast.MapType:
		// Map type
		return jen.Map(exprToJen(e.Key)).Add(exprToJen(e.Value))
	case *ast.FuncType:
		// Function type
		// For simplicity, returning "func(...)"
		return jen.Func().Params().Params()
	}
	return jen.Any()
}

func signalChannelProviderFuncName(svc *kibumod.Service, op *kibumod.Operation) string {
	return firstToUpper(fmt.Sprintf("New%sSignalChannel", firstToUpper(op.Name)))
}

// firstToLower returns a string with the first rune converted to lowercase
//
//	"Name" → "name"
func firstToLower(name string) string {
	r := []rune(name)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// ptrExpr returns a pointer to the given name
//
//	*name
func ptrExpr(name string) *jen.Statement {
	return jen.Op("*").Id(name)
}

// matchingCompilerAssertion creates an assertion that a private struct implements the service interface
//
//	var _ Name = (*name)(nil)
func matchingCompilerAssertion(name string) *jen.Statement {
	return compilerAssertionToInterface(name, firstToLower(name))
}

// compilerAssertionToInterface creates an assertion that an impl struct implements the given interface
//
//	var _ Iface = (*impl)(nil)
func compilerAssertionToInterface(iface, impl string) *jen.Statement {
	return jen.Var().Id("_").Id(iface).Op("=").
		Params(ptrExpr(impl)).
		Parens(jen.Nil())
}
