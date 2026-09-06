package kibuenum

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"sort"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const SDKPath = "github.com/kibu-sh/kibu/pkg/enum/experimental"

var Analyzer = &analysis.Analyzer{
	Name:       "kibuenum",
	Doc:        "Extract experimental typed enum constructor declarations",
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	ResultType: reflect.TypeOf((*Result)(nil)),
	Run:        run,
}

func FromPass(pass *analysis.Pass) (*Result, bool) {
	r, ok := pass.ResultOf[Analyzer].(*Result)
	return r, ok
}

func referencedObject(info *types.Info, expr ast.Expr) types.Object {
	switch e := ast.Unparen(expr).(type) {
	case *ast.Ident:
		return info.ObjectOf(e)
	case *ast.SelectorExpr:
		return info.ObjectOf(e.Sel)
	}
	return nil
}

func isSDKObject(obj types.Object, name string) bool {
	return obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == SDKPath && obj.Name() == name
}

func constructorExpression(call *ast.CallExpr) ast.Expr {
	switch f := ast.Unparen(call.Fun).(type) {
	case *ast.IndexExpr:
		return f.X
	case *ast.IndexListExpr:
		return f.X
	default:
		return f
	}
}

func run(pass *analysis.Pass) (any, error) {
	owners := packageVariableOwners(pass.Files)
	calls := constructorCalls(pass)
	enumParser := parser{pass: pass}
	result := &Result{}
	firstDeclaration := map[string]Enum{}
	conflicts := map[string]bool{}

	for _, call := range calls {
		owner := owners[call]
		if owner == nil {
			enumParser.report(call, "enum-shape", "Define must directly initialize one named package-level variable")
			continue
		}
		declaration, valid := enumParser.parseDeclaration(call, owner)
		if !valid {
			continue
		}

		typeName := declaration.Type.PackagePath + "." + declaration.Type.Name
		if first, exists := firstDeclaration[typeName]; exists {
			message := fmt.Sprintf("duplicate enum declaration for %s; first declaration at %s",
				typeName, first.Source.Start)
			enumParser.report(call, "enum-conflict", message)
			conflicts[typeName] = true
			continue
		}
		firstDeclaration[typeName] = declaration
		result.Enums = append(result.Enums, declaration)
	}
	result.Enums = excludeConflictingTypes(result.Enums, conflicts)
	return result, nil
}

func packageVariableOwners(files []*ast.File) map[*ast.CallExpr]*ast.Ident {
	// Map only direct, single-name, package-level var initializers. Calls in
	// functions, nested expressions, or multi-bindings still get diagnostics.
	owners := map[*ast.CallExpr]*ast.Ident{}
	for _, file := range files {
		for _, decl := range file.Decls {
			group, ok := decl.(*ast.GenDecl)
			if !ok || group.Tok != token.VAR {
				continue
			}
			for _, spec := range group.Specs {
				variable := spec.(*ast.ValueSpec)
				if len(variable.Names) != 1 || len(variable.Values) != 1 || variable.Names[0].Name == "_" {
					continue
				}
				if call, ok := ast.Unparen(variable.Values[0]).(*ast.CallExpr); ok {
					owners[call] = variable.Names[0]
				}
			}
		}
	}
	return owners
}

func constructorCalls(pass *analysis.Pass) []*ast.CallExpr {
	var calls []*ast.CallExpr
	walker := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	walker.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		expression := constructorExpression(call)
		function, ok := referencedObject(pass.TypesInfo, expression).(*types.Func)
		if ok && isSDKObject(function, "Define") {
			calls = append(calls, call)
		}
	})
	sort.Slice(calls, func(i, j int) bool {
		a, b := pass.Fset.Position(calls[i].Pos()), pass.Fset.Position(calls[j].Pos())
		if a.Filename != b.Filename {
			return a.Filename < b.Filename
		}
		return a.Offset < b.Offset
	})
	return calls
}

func excludeConflictingTypes(declarations []Enum, conflicts map[string]bool) []Enum {
	result := declarations[:0]
	for _, declaration := range declarations {
		typeName := declaration.Type.PackagePath + "." + declaration.Type.Name
		if !conflicts[typeName] {
			result = append(result, declaration)
		}
	}
	return result
}
