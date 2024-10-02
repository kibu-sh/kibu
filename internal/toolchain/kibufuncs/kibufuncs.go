package kibufuncs

import (
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"reflect"
)

type Map = orderedmap.OrderedMap[*types.Func, *ast.Ident]

func FromPass(pass *analysis.Pass) (*Map, bool) {
	result, ok := pass.ResultOf[Analyzer].(*Map)
	return result, ok
}

var resultType = reflect.TypeOf((*Map)(nil))

var Analyzer = &analysis.Analyzer{
	Name:             "kibufuncs",
	Doc:              "Analyzes go source code builds an index of *types.Func to *ast.Ident",
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	ResultType:       resultType,
	RunDespiteErrors: true,
	Run:              run,
}

func run(pass *analysis.Pass) (any, error) {
	var result = orderedmap.New[*types.Func, *ast.Ident]()
	for ident, object := range pass.TypesInfo.Defs {
		if f, ok := object.(*types.Func); ok {
			result.Set(f, ident)
		}
	}
	return result, nil
}
