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

func fileWithGenGoExt(name string) string {
	return fmt.Sprintf("%s.gen.go", name)
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

func childClientName(name string) string {
	return firstToUpper(fmt.Sprintf("%sChildClient", name))
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

func buildActivityInterfaces(f *jen.File, pkg *kibumod.Package) {
	f.Comment("activity interfaces")
	for _, svc := range pkg.Services {
		f.Add(activityProxyInterface(svc))
	}
	return
}

func activityProxyInterface(svc *kibumod.Service) jen.Code {
	if !svc.Decorators.Some(isKibuActivity) {
		return jen.Null()
	}

	return jen.Type().Id(proxyName(svc.Name)).InterfaceFunc(func(g *jen.Group) {
		for _, op := range svc.Operations {
			g.Id(op.Name).
				ParamsFunc(func(g *jen.Group) {
					g.Add(namedWorkflowContextParam())
					g.Add(paramToMaybeNamedExp(paramAtIndex(op.Params, 1)))
				}).
				ParamsFunc(func(g *jen.Group) {
					g.Add(paramToExp(paramAtIndex(op.Results, 0)))
					g.Error()
				})

			g.Id(nameAsync(op.Name)).
				ParamsFunc(func(g *jen.Group) {
					g.Add(namedWorkflowContextParam())
					g.Add(paramToMaybeNamedExp(paramAtIndex(op.Params, 1)))
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

func findExecuteMethod(svc *kibumod.Service) (*kibumod.Operation, bool) {
	return lo.Find(svc.Operations, func(op *kibumod.Operation) bool {
		return op.Decorators.Some(isKibuWorkflowExecute)
	})
}

func namedGetHandleOpts() jen.Code {
	return jen.Id("opts").Qual(kibuTemporalImportName, "GetHandleOptions")
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
