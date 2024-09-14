package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/huandu/xstrings"
	"github.com/kibu-sh/kibu/internal/parser"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
	"sort"
	"strings"
)

var (
	temporalSdkWorkflow     = "go.temporal.io/sdk/workflow"
	temporalSDKClient       = "go.temporal.io/sdk/client"
	kibuTransport           = "github.com/kibu-sh/kibu/pkg/transport"
	kibuTransportMiddleware = "github.com/kibu-sh/kibu/pkg/transport/middleware"
	kibuTemporal            = "github.com/kibu-sh/kibu/pkg/transport/temporal"
)

func BuildWorkerProxies(
	opts *PipelineOptions,
) (err error) {
	for _, wrk := range opts.Workers {
		f := opts.FileSet.Get(packageScopedFilePath(wrk.Package))
		wrkID := jen.Id(workerTypeID(wrk.Type))

		f.Func().Params(wrkID.Clone().Op("*").Id(wrk.Name)).
			Id("WorkerFactory").Params().
			Params(jen.Index().Op("*").Qual(kibuTemporal, "Worker")).
			BlockFunc(buildWorkerFactoryBlockFunc(wrk))

		switch wrk.Type {
		case parser.ActivityType:
			buildActivityProxy(f, wrk)
		case parser.WorkflowType:
			buildWorkflowClient(f, wrk)
			buildWorkflowProxy(f, wrk)
		default:
			err = errors.Errorf("unknown worker type: %s", wrk.Type)
			return
		}
	}
	return
}

func buildWorkerFactoryBlockFunc(wrk *parser.Worker) func(g *jen.Group) {
	return func(g *jen.Group) {
		g.ReturnFunc(func(g *jen.Group) {
			g.Index().Op("*").Qual(kibuTemporal, "Worker").CustomFunc(multiLineCurly(), func(g *jen.Group) {
				methods := toSlice(wrk.Methods)
				sort.Slice(methods, sortByID(methods))

				for _, method := range methods {
					g.Op("&").Qual(kibuTemporal, "Worker").CustomFunc(multiLineCurly(), func(g *jen.Group) {
						g.Id("Name").Op(":").Lit(workerRegistrationName(wrk.Package, wrk, method))
						g.Id("Type").Op(":").Lit(string(wrk.Type))
						g.Id("TaskQueue").Op(":").Lit(wrk.TaskQueue)
						g.Id("Handler").Op(":").Id(workerTypeID(wrk.Type)).Dot(method.Name)
					})
				}
			})
			return
		})
		return
	}
}
func buildWorkflowClient(f *jen.File, wrk *parser.Worker) {
	methods := toSlice(wrk.Methods)
	scope := wrk.Package.GoPackage
	f.Type().Id(workerClientName(wrk)).StructFunc(func(g *jen.Group) {
		g.Id("ref").Qual("", wrk.Name)
		g.Id("Temporal").Qual(temporalSDKClient, "Client")
	})

	sort.Slice(methods, sortByID(methods))

	for _, method := range methods {
		f.Func().
			Params(jen.Id("c").Id(workerClientName(wrk))).Id(method.Name).
			Params(
				jen.Id("ctx").Qual("context", "Context"),
				jen.Id("id").String(),
				parserVarAsNamedParam("req", scope, method.Request),
			).
			Params(
				jen.Qual(kibuTemporal, "WorkflowRun").Types(
					parserVarAsTypeParam(scope, method.Response),
				),
				jen.Error(),
			).
			BlockFunc(func(g *jen.Group) {
				g.Return().Qual(kibuTemporal, "NewWorkflowRunWithErr").Types(parserVarAsTypeParam(scope, method.Response)).CustomFunc(multiLineParen(), func(g *jen.Group) {
					g.Id("c").Dot("Temporal").Dot("ExecuteWorkflow").CallFunc(func(g *jen.Group) {
						g.Id("ctx")
						g.Qual(temporalSDKClient, "StartWorkflowOptions").CustomFunc(multiLineCurly(), func(g *jen.Group) {
							g.Id("ID").Op(":").Id("id")
							g.Id("TaskQueue").Op(":").Lit(wrk.TaskQueue)
						})
						g.Id("c").Dot("ref").Dot(method.Name)
						g.Id("req")
						return
					})
				})
			})
	}
}
func buildWorkflowProxy(f *jen.File, wrk *parser.Worker) {
	scope := wrk.Package.GoPackage
	f.Type().Id(workerProxyName(wrk)).StructFunc(func(g *jen.Group) {
		g.Id("ref").Qual("", wrk.Name)
	})

	methods := toSlice(wrk.Methods)
	sort.Slice(methods, sortByID(methods))

	for _, method := range methods {
		f.Func().
			Params(jen.Id("p").Id(workerProxyName(wrk))).Id(method.Name).
			Params(
				jen.Id("ctx").Qual(temporalSdkWorkflow, "Context"),
				parserVarAsNamedParam("req", scope, method.Request),
			).
			Params(
				jen.Qual(kibuTemporal, "ChildWorkflowFuture").Types(
					parserVarAsTypeParam(scope, method.Response),
				),
			).
			BlockFunc(func(g *jen.Group) {
				g.Return().Qual(kibuTemporal, "NewChildWorkflowFuture").Types(parserVarAsTypeParam(scope, method.Response)).CustomFunc(multiLineParen(), func(g *jen.Group) {
					g.Qual(temporalSdkWorkflow, "ExecuteChildWorkflow").CallFunc(func(g *jen.Group) {
						g.Id("ctx")
						g.Id("p").Dot("ref").Dot(method.Name)
						g.Id("req")
						return
					})
				})
			})
	}
}

func buildActivityProxy(f *jen.File, wrk *parser.Worker) {
	methods := toSlice(wrk.Methods)
	scope := wrk.Package.GoPackage
	f.Type().Id(workerProxyName(wrk)).StructFunc(func(g *jen.Group) {})
	sort.Slice(methods, sortByID(methods))

	for _, method := range methods {
		f.Func().
			Params(jen.Id("p").Id(workerProxyName(wrk))).Id(method.Name).
			Params(
				jen.Id("ctx").Qual(temporalSdkWorkflow, "Context"),
				parserVarAsNamedParam("req", scope, method.Request),
			).
			Params(
				jen.Qual(kibuTemporal, "Future").Types(
					parserVarAsTypeParam(scope, method.Response),
				),
			).
			BlockFunc(func(g *jen.Group) {
				g.Return(jen.Qual(kibuTemporal, "NewFuture").Types(parserVarAsTypeParam(scope, method.Response)).CustomFunc(multiLineParen(), func(g *jen.Group) {
					g.Qual(temporalSdkWorkflow, "ExecuteActivity").CallFunc(func(g *jen.Group) {
						g.Id("ctx")
						g.Lit(workerRegistrationName(wrk.Package, wrk, method))
						g.Id("req")
						return
					})
				}))
			})
	}
}

func multiLineCurly() jen.Options {
	return jen.Options{
		Open:      "{",
		Close:     "}",
		Separator: ",",
		Multi:     true,
	}
}

func multiLineParen() jen.Options {
	return jen.Options{
		Open:      "(",
		Close:     ")",
		Separator: ",",
		Multi:     true,
	}
}

func workerRegistrationName(pkg *parser.Package, wrk *parser.Worker, method *parser.Method) string {
	return strings.Join([]string{
		pkg.Name,
		wrk.Name,
		method.Name,
	}, ".")
}

func workerTypeID(wrkType parser.WorkerType) string {
	if wrkType == parser.ActivityType {
		return "act"
	}
	return "wkr"
}

func workerTypeCamelCase(wrkType parser.WorkerType) string {
	return xstrings.ToCamelCase(string(wrkType))
}

func workerFactoryName(wrkType parser.WorkerType) string {
	return workerTypeCamelCase(wrkType) + "Factory"
}

func workerProxyName(wrk *parser.Worker) string {
	return strings.Join([]string{
		wrk.Name,
		"Proxy",
	}, "__")
}

func workerClientName(wrk *parser.Worker) string {
	return strings.Join([]string{
		wrk.Name,
		"Client",
	}, "__")
}

func parserVarAsNamedParam(id string, scope *packages.Package, v *parser.Var) *jen.Statement {
	if scope.PkgPath == v.TypePkgPath() {
		return jen.Id(id).Id(v.TypeName())
	}
	return jen.Id(id).Qual(v.TypePkgPath(), v.TypeName())
}

func parserVarAsTypeParam(scope *packages.Package, v *parser.Var) *jen.Statement {
	if scope.PkgPath == v.TypePkgPath() {
		return jen.Id(v.TypeName())
	}
	return jen.Qual(v.TypePkgPath(), v.TypeName())
}
