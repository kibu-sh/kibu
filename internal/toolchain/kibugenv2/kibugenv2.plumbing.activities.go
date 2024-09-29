package kibugenv2

import (
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
)

func buildActivityImplementations(file *jen.File, pkg *kibumod.Package) {
	for _, svc := range pkg.Services {
		if !svc.Decorators.Some(isKibuActivity) {
			continue
		}

		file.Type().Id(firstToLower(proxyName(svc.Name))).Struct()

		for _, op := range svc.Operations {
			file.Add(buildActivityProxyMethod(svc, op))
			file.Add(buildActivityProxyAsyncMethod(svc, op))
		}
	}
}

func buildActivityProxyMethod(svc *kibumod.Service, op *kibumod.Operation) jen.Code {
	return jen.Func().Params(jen.Id("a").Op("*").Id(firstToLower(proxyName(svc.Name)))).Id(op.Name).
		ParamsFunc(func(g *jen.Group) {
			g.Add(namedWorkflowContextParam())
			g.Add(paramToMaybeNamedExp(paramAtIndex(op.Params, 1)))
			g.Id("mods").Op("...").Add(qualKibuTemporalActivityOptionFunc())
		}).
		ParamsFunc(func(g *jen.Group) {
			g.Id("res").Add(paramToExp(paramAtIndex(op.Results, 0)))
			g.Id("err").Error()
		}).
		Block(
			jen.Return(jen.Id("a").Dot(suffixAsync(op.Name)).Call(
				jen.Id("ctx"),
				jen.Id("req"),
				jen.Id("mods").Op("..."))).Dot("Get").Call(jen.Id("ctx")),
		)
}

func qualKibuTemporalActivityOptionFunc() jen.Code {
	return jen.Qual(kibuTemporalImportName, "ActivityOptionFunc")
}

func buildActivityProxyAsyncMethod(svc *kibumod.Service, op *kibumod.Operation) jen.Code {
	return jen.Func().Params(jen.Id("a").Op("*").Id(firstToLower(proxyName(svc.Name)))).Id(suffixAsync(op.Name)).
		ParamsFunc(func(g *jen.Group) {
			g.Add(namedWorkflowContextParam())
			g.Add(paramToMaybeNamedExp(paramAtIndex(op.Params, 1)))
			g.Id("mods").Op("...").Qual(kibuTemporalImportName, "ActivityOptionFunc")
		}).
		Params(jen.Qual(kibuTemporalImportName, "Future").Types(paramToExp(paramAtIndex(op.Results, 0)))).
		Block(
			jen.Id("options").Op(":=").Qual(kibuTemporalImportName, "NewActivityOptionsBuilder").Call().Dot("WithStartToCloseTimeout").Call(jen.Qual("time", "Second").Op("*").Lit(30)).Dot("WithTaskQueue").Call(jen.Id(packageNameConst())).Dot("WithProvidersWhenSupported").Call(jen.Id("req")).Dot("WithOptions").Call(jen.Id("mods").Op("...")).Dot("Build").Call(),
			jen.Id("ctx").Op("=").Qual(temporalWorkflowImportName, "WithActivityOptions").Call(jen.Id("ctx"), jen.Id("options")),
			jen.Id("future").Op(":=").Qual(temporalWorkflowImportName, "ExecuteActivity").Call(jen.Id("ctx"), jen.Id(operationConstName(svc, op)), jen.Id("req")),
			jen.Return(jen.Qual(kibuTemporalImportName, "NewFuture").Types(paramToExp(paramAtIndex(op.Results, 0))).Call(jen.Id("future"))),
		)
}
