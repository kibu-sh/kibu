package kibugenv2

import (
	"errors"
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
	"golang.org/x/tools/go/analysis"
	"reflect"
)

type Artifact struct {
	File *jen.File
}

var resultType = reflect.TypeOf((*Artifact)(nil))

func FromPass(pass *analysis.Pass) (*Artifact, bool) {
	result, ok := pass.ResultOf[Analyzer].(*Artifact)
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

	genFile := newGenFile(pass.Pkg)

	result := &Artifact{
		File: genFile,
	}

	buildPkgCompilerAssertions(genFile, pkg)
	buildPkgConstants(genFile, pkg)
	buildSignalChannelFuncs(genFile, pkg)
	buildWorkflowInterfaces(genFile, pkg)
	buildActivityInterfaces(genFile, pkg)
	buildActivityImplementations(genFile, pkg)
	buildWorkflowControllers(genFile, pkg)
	buildActivitiesControllers(genFile, pkg)
	buildServiceControllers(genFile, pkg)
	buildWorkerController(genFile, pkg)
	return result, nil
}
