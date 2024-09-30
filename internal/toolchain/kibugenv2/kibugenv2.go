package kibugenv2

import (
	"errors"
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"golang.org/x/tools/go/analysis"
	"reflect"
)

var resultType = reflect.TypeOf((modspecv2.Artifact)(nil))

func FromPass(pass *analysis.Pass) (modspecv2.Artifact, bool) {
	result, ok := pass.ResultOf[Analyzer].(modspecv2.Artifact)
	return result, ok
}

var Analyzer = &analysis.Analyzer{
	Name:             "kibugenv2",
	Doc:              "Analyzes go source code and generates system plumbing for kibu applications",
	Run:              run,
	ResultType:       resultType,
	RunDespiteErrors: true,
	Requires: []*analysis.Analyzer{
		kibumod.Analyzer,
	},
}

var missingPackageError = errors.New("missing result of kibumod analyzer")

func run(pass *analysis.Pass) (any, error) {
	pkg, ok := kibumod.FromPass(pass)
	if !ok {
		return nil, missingPackageError
	}

	genFile := modspecv2.NewJenFileFromPackage(pass.Pkg)
	result := modspecv2.NewPackageArtifact(genFile, pass, "")

	generate(genFile, pkg,
		buildPkgCompilerAssertions,
		buildPkgConstants,
		buildSignalChannelFuncs,
		buildWorkflowInterfaces,
		buildActivityInterfaces,
		buildActivityImplementations,
		buildWorkflowControllers,
		buildActivitiesControllers,
		buildServiceControllers,
		buildWorkerController,
	)

	return result, nil
}

type genFunc func(genFile *jen.File, pkg *modspecv2.Package)

func generate(genFile *jen.File, pkg *modspecv2.Package, genFuncs ...genFunc) {
	for _, genFunc := range genFuncs {
		genFunc(genFile, pkg)
	}
	return
}
