package kibudecor

import (
	"github.com/kibu-sh/kibu/internal/parser/directive"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"reflect"
)

type Map = orderedmap.OrderedMap[*ast.Ident, directive.List]

var resultType = reflect.TypeOf((*Map)(nil))

func FromPass(pass *analysis.Pass) (*Map, bool) {
	result, ok := pass.ResultOf[Analyzer].(*Map)
	return result, ok
}

var Analyzer = &analysis.Analyzer{
	Name:             "kibudirectives",
	Doc:              "Analyzes go source code for kibu directive annotations",
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	ResultType:       resultType,
	RunDespiteErrors: true,
	Run:              run,
}

func run(pass *analysis.Pass) (any, error) {
	var result = orderedmap.New[*ast.Ident, directive.List]()
	walk := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
		(*ast.FuncDecl)(nil),
	}

	walk.Preorder(nodeFilter, func(n ast.Node) {
		decl, ok := n.(ast.Decl)
		if !ok {
			return
		}

		if err := directive.ApplyFromDecl(decl, result); err != nil {
			pass.Reportf(n.Pos(), "failed to parse directive: %v", err)
			return
		}
	})

	return result, nil
}
