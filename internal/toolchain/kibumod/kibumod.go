package kibumod

import (
	"github.com/kibu-sh/kibu/internal/toolchain/kibudirectives"
	"github.com/kibu-sh/kibu/internal/toolchain/kibufuncs"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"reflect"
)

type Package struct {
	Name       string
	Services   []*Service
	Funcs      *kibufuncs.Map
	Directives *kibudirectives.Map
}

type Operations struct {
	Name string
}

type Service struct {
	Name       string
	Operations []*Operations
}

var returnType = reflect.TypeOf((*Package)(nil))

var Analyzer = &analysis.Analyzer{
	Name:             "kibumod",
	Doc:              "Analyzes go source code for kibu service definitions",
	Run:              run,
	ResultType:       returnType,
	RunDespiteErrors: true,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		kibufuncs.Analyzer,
		kibudirectives.Analyzer,
	},
}

func run(pass *analysis.Pass) (any, error) {
	//walk := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	directives, _ := kibudirectives.FromPass(pass)
	funcs, _ := kibufuncs.FromPass(pass)

	pkg := &Package{
		Name:       pass.Pkg.Name(),
		Funcs:      funcs,
		Directives: directives,
	}

	return pkg, nil
}
