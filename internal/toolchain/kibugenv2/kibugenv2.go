package kibugenv2

import (
	"errors"
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
	"golang.org/x/tools/go/analysis"
	"reflect"
)

type Artifact struct {
	Files []*jen.File
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
	pkgs, ok := kibumod.FromPass(pass)
	if !ok {
		return nil, missingPackageError
	}

	result := new(Artifact)
	generators := []GenFunc{
		generatePlumbing,
	}

	return result, generate(&GenParams{
		Pass:     pass,
		Pkg:      pkgs,
		Artifact: result,
	}, generators...)
}

type GenParams struct {
	Pass     *analysis.Pass
	Pkg      *kibumod.Package
	Artifact *Artifact
}

type GenFunc func(params *GenParams) error

func generate(params *GenParams, generators ...GenFunc) error {
	for _, generator := range generators {
		if err := generator(params); err != nil {
			return err
		}
	}
	return nil
}
