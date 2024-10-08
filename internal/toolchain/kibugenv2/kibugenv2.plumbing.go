package kibugenv2

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"go/ast"
	"net/http"
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
	kibuMiddlewareImportName   = "github.com/kibu-sh/kibu/pkg/transport/middleware"
	temporalActivityImportName = "go.temporal.io/sdk/activity"
	temporalClientImportName   = "go.temporal.io/sdk/client"
	temporalWorkerImportName   = "go.temporal.io/sdk/worker"
	temporalWorkflowImportName = "go.temporal.io/sdk/workflow"
	timeImportName             = "time"
)

var (
	isKibuService       = decorators.HasKey(kibuPrefix, serviceName)
	isKibuServiceMethod = decorators.HasKey(kibuPrefix, serviceName, "method")

	isKibuActivity       = decorators.HasKey(kibuPrefix, activityName)
	isKibuActivityMethod = decorators.HasKey(kibuPrefix, activityName, "method")

	isKibuWorkflow        = decorators.HasKey(kibuPrefix, workflowName)
	isKibuWorkflowExecute = decorators.HasKey(kibuPrefix, workflowName, workflowExecuteName)
	isKibuWorkflowUpdate  = decorators.HasKey(kibuPrefix, workflowName, workflowUpdateName)
	isKibuWorkflowQuery   = decorators.HasKey(kibuPrefix, workflowName, workflowQueryName)
	isKibuWorkflowSignal  = decorators.HasKey(kibuPrefix, workflowName, workflowSignalName)
	isActivityOrWorkflow  = decorators.OneOf(isKibuWorkflow, isKibuActivity)
)

func firstToUpper(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func suffixProxy(name string) string {
	return firstToUpper(fmt.Sprintf("%sProxy", name))
}

func suffixRun(name string) string {
	return firstToUpper(fmt.Sprintf("%sRun", name))
}

func suffixChildRun(name string) string {
	return firstToUpper(fmt.Sprintf("%sChildRun", name))
}

func suffixController(name string) string {
	return firstToUpper(fmt.Sprintf("%sController", name))
}

func suffixAsync(name string) string {
	return firstToUpper(fmt.Sprintf("%sAsync", name))
}

func suffixChannel(name string) string {
	return firstToUpper(fmt.Sprintf("%sChannel", name))
}

func suffixFactory(name string) string {
	return firstToUpper(fmt.Sprintf("%sFactory", name))
}

func suffixInput(name string) string {
	return firstToUpper(fmt.Sprintf("%sInput", name))
}

func suffixExternalRun(name string) string {
	return firstToUpper(fmt.Sprintf("%sExternalRun", name))
}

func suffixChildClient(name string) string {
	return firstToUpper(fmt.Sprintf("%sChildClient", name))
}

func suffixClient(name string) string {
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
func buildPkgCompilerAssertions(f *jen.File, pkg *modspecv2.Package) {
	f.Comment("compiler assertions")
	for _, svc := range pkg.Services {
		if svc.Decorators.Some(isKibuActivity) {
			f.Add(compilerAssertionToInterface(
				suffixProxy(svc.Name), firstToLower(suffixProxy(svc.Name))))
		}

		if svc.Decorators.Some(isKibuWorkflow) {
			f.Add(compilerAssertionToInterface(
				suffixChildRun(svc.Name), firstToLower(suffixChildRun(svc.Name))))

			f.Add(compilerAssertionToInterface(
				suffixClient(svc.Name), firstToLower(suffixClient(svc.Name))))
		}
	}
	return
}

func svcConstName(svc *modspecv2.Service) string {
	return privateConstName(svc.Name)
}

func operationConstName(svc *modspecv2.Service, op *modspecv2.Operation) string {
	return privateConstName(fmt.Sprintf("%s%s", svc.Name, firstToUpper(op.Name)))
}

func svcConstLiteral(pkg *modspecv2.Package, svc *modspecv2.Service) string {
	return fmt.Sprintf("%s.%s", pkg.Name, svc.Name)
}

func operationConstLiteral(pkg *modspecv2.Package, svc *modspecv2.Service, op *modspecv2.Operation) string {
	return fmt.Sprintf("%s.%s.%s", pkg.Name, svc.Name, op.Name)
}

// buildPkgConstants a set of constant references for later code
func buildPkgConstants(f *jen.File, pkg *modspecv2.Package) {
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

func buildActivityInterfaces(f *jen.File, pkg *modspecv2.Package) {
	f.Comment("activity interfaces")
	for _, svc := range pkg.Services {
		f.Add(activityProxyInterface(svc))
	}
	return
}

func activityProxyInterface(svc *modspecv2.Service) jen.Code {
	if !svc.Decorators.Some(isKibuActivity) {
		return jen.Null()
	}

	return jen.Type().Id(suffixProxy(svc.Name)).InterfaceFunc(func(g *jen.Group) {
		for _, op := range svc.Operations {
			g.Id(op.Name).
				ParamsFunc(func(g *jen.Group) {
					g.Add(namedWorkflowContextParam())
					g.Add(paramToMaybeNamedExp(paramAtIndex(op.Params, 1)))
					g.Id("mods").Op("...").Add(qualKibuTemporalActivityOptionFunc())
				}).
				ParamsFunc(func(g *jen.Group) {
					g.Add(paramToExp(paramAtIndex(op.Results, 0)))
					g.Error()
				})

			g.Id(suffixAsync(op.Name)).
				ParamsFunc(func(g *jen.Group) {
					g.Add(namedWorkflowContextParam())
					g.Add(paramToMaybeNamedExp(paramAtIndex(op.Params, 1)))
					g.Id("mods").Op("...").Add(qualKibuTemporalActivityOptionFunc())
				}).
				ParamsFunc(func(g *jen.Group) {
					g.Add(qualKibuTemporalFuture(paramToExpOrAny(paramAtIndex(op.Results, 0))))
				})
		}
	})
}

func qualKibuTemporalWorkflowOptionFunc() jen.Code {
	return jen.Qual(kibuTemporalImportName, "WorkflowOptionFunc")
}

func qualKibuTemporalExecuteWithSignalParams(req jen.Code, sig jen.Code) jen.Code {
	return jen.Qual(kibuTemporalImportName, "ExecuteWithSignalParams").Types(req, sig)
}

func qualKibuTemporalExecuteParams(req jen.Code) jen.Code {
	return jen.Qual(kibuTemporalImportName, "ExecuteParams").Types(req)
}

func qualKibuTemporalFuture(res jen.Code) jen.Code {
	return jen.Qual(kibuTemporalImportName, "Future").Types(res)
}

func executeWithName(name string) string {
	return firstToUpper(fmt.Sprintf("ExecuteWith%s", firstToUpper(name)))
}

func findExecuteMethod(svc *modspecv2.Service) (*modspecv2.Operation, bool) {
	return lo.Find(svc.Operations, func(op *modspecv2.Operation) bool {
		return op.Decorators.Some(isKibuWorkflowExecute)
	})
}

func namedGetHandleOpts() jen.Code {
	return jen.Id("opts").Qual(kibuTemporalImportName, "GetHandleOpts")
}

func qualWorkflowExecution() jen.Code {
	return jen.Qual(temporalWorkflowImportName, "Execution")
}

func filterSignalMethods(operations []*modspecv2.Operation) []*modspecv2.Operation {
	return lo.Filter(operations, func(op *modspecv2.Operation, _ int) bool {
		return op.Decorators.Some(isKibuWorkflowSignal)
	})
}

func qualWorkflowChildRunFuture() jen.Code {
	return jen.Qual(temporalWorkflowImportName, "ChildWorkflowFuture")
}

func qualWorkflowFuture() jen.Code {
	return jen.Qual(temporalWorkflowImportName, "Future")
}

func filterUpdateMethods(operations []*modspecv2.Operation) []*modspecv2.Operation {
	return lo.Filter(operations, func(op *modspecv2.Operation, _ int) bool {
		return op.Decorators.Some(isKibuWorkflowUpdate)
	})
}

func filterSignalAndQueryMethods(operations []*modspecv2.Operation) []*modspecv2.Operation {
	return lo.Filter(operations, func(op *modspecv2.Operation, _ int) bool {
		return op.Decorators.Some(decorators.OneOf(
			isKibuWorkflowSignal,
			isKibuWorkflowQuery,
		))
	})
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
	return jen.Id("sel").Add(qualWorkflowSelector())
}

func qualWorkflowSelector() jen.Code {
	return jen.Qual(temporalWorkflowImportName, "Selector")
}

func signalChannelProviderFunc(svc *modspecv2.Service, op *modspecv2.Operation) jen.Code {
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

type optionalParam = mo.Option[modspecv2.Type]

func paramAtIndex(params []modspecv2.Type, index int) optionalParam {
	if index < 0 || index >= len(params) {
		return mo.None[modspecv2.Type]()
	}

	return mo.Some[modspecv2.Type](params[index])
}

func paramToExp(param optionalParam) jen.Code {
	if param.IsAbsent() {
		return jen.Null()
	}
	return exprToJen(param.MustGet().Field.Type)
}

func paramToExpOrAny(param optionalParam) jen.Code {
	if param.IsAbsent() {
		return jen.Any()
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

func signalChannelProviderFuncName(svc *modspecv2.Service, op *modspecv2.Operation) string {
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
func buildServiceControllers(f *jen.File, pkg *modspecv2.Package) {
	for _, svc := range pkg.Services {
		if !svc.Decorators.Some(isKibuService) {
			continue
		}

		f.Comment("//kibu:provider group=HandlerFactory import=github.com/kibu-sh/kibu/pkg/transport/httpx")
		f.Type().Id(suffixController(svc.Name)).Struct(
			jen.Id("Service").Id(svc.Name),
		)

		f.Func().Params(
			jen.Id("svc").Op("*").Id(suffixController(svc.Name)),
		).Id("HTTPHandlerFactory").Params(jen.Id("_").Op("*").Qual(kibuMiddlewareImportName, "Registry")).Params(
			jen.Index().Op("*").Qual(kibuHttpxImportName, "Handler"),
		).BlockFunc(func(g *jen.Group) {
			g.ReturnFunc(func(g *jen.Group) {
				g.Index().Op("*").Qual(kibuHttpxImportName, "Handler").CustomFunc(modspecv2.MultiLineCurly(), func(g *jen.Group) {
					for _, op := range svc.Operations {
						methodDecorator, _ := op.Decorators.Find(isKibuServiceMethod)

						// TODO: warn on analysis pass that there's a duplicate path detected
						// 	this is due to multiple Service interfaces defined in the same Package
						path, _ := methodDecorator.Options.GetOne("path",
							fmt.Sprintf("/%s/%s", pkg.Name, op.Name))

						// TODO: support more than one method per service call
						//  although this usually should be POST since JSON serialization will be most common
						method, _ := methodDecorator.Options.GetOne("method",
							http.MethodPost)

						g.Id("httpx").Dot("NewHandler").
							Call(jen.Lit(path),
								jen.Qual(kibuTransportImportName, "NewEndpoint").
									Call(jen.Id("svc").Dot("Service").Dot(op.Name)),
							).Dot("WithMethods").Call(jen.Lit(method))

					}
				})
			})
		})
	}
}

func buildActivitiesControllers(f *jen.File, pkg *modspecv2.Package) {
	for _, svc := range pkg.Services {
		if !svc.Decorators.Some(isKibuActivity) {
			continue
		}

		f.Comment("//kibu:provider")
		f.Type().Id(suffixController(svc.Name)).Struct(
			jen.Id("Activities").Id(svc.Name),
		)

		f.Func().Params(
			jen.Id("act").Op("*").Id(suffixController(svc.Name)),
		).Id("Build").Params(
			jen.Id("registry").Qual(temporalWorkerImportName, "ActivityRegistry"),
		).BlockFunc(func(g *jen.Group) {
			for _, op := range svc.Operations {
				g.Id("registry").Dot("RegisterActivityWithOptions").Call(
					jen.Id("act").Dot("Activities").Dot(op.Name),
					jen.Qual(temporalActivityImportName, "RegisterOptions").Values(jen.DictFunc(func(d jen.Dict) {
						d[jen.Id("Name")] = jen.Id(operationConstName(svc, op))
						d[jen.Id("DisableAlreadyRegisteredCheck")] = jen.True()
					})),
				)
			}
		})
	}
}
