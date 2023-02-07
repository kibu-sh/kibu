package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/discernhq/devx/internal/parser"
	"github.com/huandu/xstrings"
	"github.com/pkg/errors"
	"strings"
)

var (
	temporalSdkWorkflow = "go.temporal.io/sdk/workflow"
	temporalSDKClient   = "go.temporal.io/sdk/client"
	devxTemporal        = "github.com/discernhq/devx/pkg/transport/temporal"
)

func BuildWorkerProxies(
	opts *GeneratorOptions,
) (err error) {
	for _, path := range opts.PackageList.Iterator() {
		pkg := path.Value
		f := opts.FileSet.Get(packageScopedFilePath(pkg))
		for _, ident := range pkg.Workers.Iterator() {
			wrk := ident.Value
			wrkID := jen.Id(workerTypeID(wrk.Type))

			f.Func().Params(wrkID.Clone().Op("*").Id(wrk.Name)).
				Id("WorkerFactory").Params().
				Params(jen.Index().Op("*").Qual(devxTemporal, "Worker")).
				BlockFunc(buildWorkerFactoryBlockFunc(wrk, pkg))

			switch wrk.Type {
			case parser.WorkflowType:
				buildWorkflowProxy(f, wrk, pkg, opts)
			case parser.ActivityType:
				buildActivityProxy(f, wrk, pkg, opts)
			default:
				err = errors.Errorf("unknown worker type: %s", wrk.Type)
				return
			}
		}
	}
	return
}

func buildWorkerFactoryBlockFunc(wrk *parser.Worker, pkg *parser.Package) func(g *jen.Group) {
	return func(g *jen.Group) {
		g.ReturnFunc(func(g *jen.Group) {
			g.Index().Op("*").Qual(devxTemporal, "Worker").CustomFunc(multiLineCurly(), func(g *jen.Group) {
				for _, elem := range wrk.Methods.Iterator() {
					method, _ := wrk.Methods.Get(elem.Key)
					g.Op("&").Qual(devxTemporal, "Worker").CustomFunc(multiLineCurly(), func(g *jen.Group) {
						g.Id("Name").Op(":").Lit(workerRegistrationName(pkg, wrk, method))
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
func buildWorkflowProxy(f *jen.File, wrk *parser.Worker, pkg *parser.Package, opts *GeneratorOptions) {
	f.Type().Id(workerProxyName(wrk)).StructFunc(func(g *jen.Group) {
		g.Id("Temporal").Qual(temporalSDKClient, "Client")
	})

	for _, elem := range wrk.Methods.Iterator() {
		method := elem.Value
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
					g.Lit(workerRegistrationName(pkg, wrk, method))
					g.Id(method.Request.Name)
					return
				})
				g.If(jen.Err().Op("!=").Nil()).Block(jen.Return())
				g.Id("err").Op("=").Id("run").Dot("Get").Call(jen.Id("ctx"), jen.Op("&").Id(method.Response.Name))
				g.Return()
			})
	}
}

func buildActivityProxy(f *jen.File, wrk *parser.Worker, pkg *parser.Package, opts *GeneratorOptions) {
	f.Type().Id(workerProxyName(wrk)).StructFunc(func(g *jen.Group) {})

	for _, elem := range wrk.Methods.Iterator() {
		method := elem.Value
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
					g.Lit(workerRegistrationName(pkg, wrk, method))
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
