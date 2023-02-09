package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/discernhq/devx/internal/parser"
	"github.com/huandu/xstrings"
	"github.com/pkg/errors"
	"sort"
	"strings"
)

var (
	temporalSdkWorkflow     = "go.temporal.io/sdk/workflow"
	temporalSDKClient       = "go.temporal.io/sdk/client"
	devxTransport           = "github.com/discernhq/devx/pkg/transport"
	devxTransportMiddleware = "github.com/discernhq/devx/pkg/transport/middleware"
	devxTemporal            = "github.com/discernhq/devx/pkg/transport/temporal"
)

func BuildWorkerProxies(
	opts *PipelineOptions,
) (err error) {
	for _, wrk := range opts.Workers {
		f := opts.FileSet.Get(packageScopedFilePath(wrk.Package))
		wrkID := jen.Id(workerTypeID(wrk.Type))

		f.Func().Params(wrkID.Clone().Op("*").Id(wrk.Name)).
			Id("WorkerFactory").Params().
			Params(jen.Index().Op("*").Qual(devxTemporal, "Worker")).
			BlockFunc(buildWorkerFactoryBlockFunc(wrk))

		switch wrk.Type {
		case parser.WorkflowType:
			buildWorkflowProxy(f, wrk)
		case parser.ActivityType:
			buildActivityProxy(f, wrk)
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
			g.Index().Op("*").Qual(devxTemporal, "Worker").CustomFunc(multiLineCurly(), func(g *jen.Group) {
				methods := toSlice(wrk.Methods)
				sort.Slice(methods, sortByID(methods))

				for _, method := range methods {
					g.Op("&").Qual(devxTemporal, "Worker").CustomFunc(multiLineCurly(), func(g *jen.Group) {
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
func buildWorkflowProxy(f *jen.File, wrk *parser.Worker) {
	f.Type().Id(workerProxyName(wrk)).StructFunc(func(g *jen.Group) {
		g.Id("Temporal").Qual(temporalSDKClient, "Client")
	})

	methods := toSlice(wrk.Methods)
	sort.Slice(methods, sortByID(methods))

	for _, method := range methods {
		f.Func().
			Params(jen.Id("p").Id(workerProxyName(wrk))).Id(method.Name).
			Params(
				jen.Id("ctx").Qual("context", "Context"),
				jen.Id("id").String(),
				jen.Id(method.Request.Name).Id(method.Request.Type),
			).
			Params(jen.Id(method.Response.Name).Id(method.Response.Type), jen.Id("err").Error()).
			BlockFunc(func(g *jen.Group) {
				g.List(jen.Id("run"), jen.Id("err")).Op(":=").Id("p").Dot("Temporal").Dot("ExecuteWorkflow").CallFunc(func(g *jen.Group) {
					g.Id("ctx")
					g.Qual(temporalSDKClient, "StartWorkflowOptions").CustomFunc(multiLineCurly(), func(g *jen.Group) {
						g.Id("ID").Op(":").Id("id")
						g.Id("TaskQueue").Op(":").Lit(wrk.TaskQueue)
					})
					g.Lit(workerRegistrationName(wrk.Package, wrk, method))
					g.Id(method.Request.Name)
					return
				})
				g.If(jen.Err().Op("!=").Nil()).Block(jen.Return())
				g.Id("err").Op("=").Id("run").Dot("Get").Call(jen.Id("ctx"), jen.Op("&").Id(method.Response.Name))
				g.Return()
			})
	}
}

func buildActivityProxy(f *jen.File, wrk *parser.Worker) {
	f.Type().Id(workerProxyName(wrk)).StructFunc(func(g *jen.Group) {})
	methods := toSlice(wrk.Methods)
	sort.Slice(methods, sortByID(methods))

	for _, method := range methods {
		f.Func().
			Params(jen.Id("p").Id(workerProxyName(wrk))).Id(method.Name).
			Params(
				jen.Id("ctx").Qual(temporalSdkWorkflow, "Context"),
				jen.Id(method.Request.Name).Id(method.Request.Type),
			).
			Params(jen.Id(method.Response.Name).Id(method.Response.Type), jen.Id("err").Error()).
			BlockFunc(func(g *jen.Group) {
				g.Id("err").Op("=").Qual(temporalSdkWorkflow, "ExecuteActivity").CallFunc(func(g *jen.Group) {
					g.Qual(temporalSdkWorkflow, "WithActivityOptions").CallFunc(func(g *jen.Group) {
						g.Id("ctx")
						g.Qual(temporalSdkWorkflow, "ActivityOptions").CustomFunc(multiLineCurly(), func(g *jen.Group) {
							g.Id("StartToCloseTimeout").Op(":").Qual("time", "Second").Op("*").Lit(30)
						})
					})
					g.Lit(workerRegistrationName(wrk.Package, wrk, method))
					g.Id(method.Request.Name)
					return
				}).Dot("Get").Call(jen.Id("ctx"), jen.Op("&").Id(method.Response.Name))
				g.Return()
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
