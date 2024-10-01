package kibumod

import (
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"github.com/samber/lo"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"reflect"
	"strings"
)

var returnType = reflect.TypeOf((*modspecv2.Package)(nil))

func FromPass(pass *analysis.Pass) (*modspecv2.Package, bool) {
	result, ok := pass.ResultOf[Analyzer].(*modspecv2.Package)
	return result, ok
}

var Analyzer = &analysis.Analyzer{
	Name:             "kibumod",
	Doc:              "Analyzes go source code for kibu service definitions",
	Run:              run,
	ResultType:       returnType,
	RunDespiteErrors: true,
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	walk := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	pkg := &modspecv2.Package{
		Name:     pass.Pkg.Name(),
		GoPkg:    pass.Pkg,
		GoModule: pass.Module,
	}

	filter := []ast.Node{
		(*ast.GenDecl)(nil),
	}

	walk.Preorder(filter, func(n ast.Node) {
		decl, ok := n.(*ast.GenDecl)
		if !ok {
			return
		}

		// no doc comments
		if decl.Doc == nil || len(decl.Doc.List) == 0 {
			return
		}

		// comments don't contain kibu decorators
		decor, _ := decorators.FromCommentGroup(decl.Doc)
		if !decor.Some(decorators.HasTool("kibu")) {
			return
		}

		// not a valid type specification for something with a kibu decorator
		// perhaps we should raise here
		ts, ok := decl.Specs[0].(*ast.TypeSpec)
		if !ok {
			return
		}

		// must be attached to an interface
		iface, ok := ts.Type.(*ast.InterfaceType)
		if !ok {
			return
		}

		pkg.Services = append(pkg.Services,
			extractServiceFromInterface(pass,
				extractServiceFromInterfaceParams{
					decl:  decl,
					iface: iface,
					tspec: ts,
				}))
		return
	})

	_ = pkg
	return pkg, nil
}

type extractServiceFromInterfaceParams struct {
	decl  *ast.GenDecl
	iface *ast.InterfaceType
	tspec *ast.TypeSpec
}

func extractServiceFromInterface(pass *analysis.Pass, opts extractServiceFromInterfaceParams) *modspecv2.Service {
	return &modspecv2.Service{
		Name:       opts.tspec.Name.Name,
		Doc:        extractDoc(pass, opts.decl.Doc),
		Decorators: extractDecorators(pass, opts.decl.Doc),
		Operations: extractOperations(pass, opts.iface),
		Decl:       opts.decl,
		Iface:      opts.iface,
		Tspec:      opts.tspec,
	}
}

func extractOperations(pass *analysis.Pass, iface *ast.InterfaceType) (result []*modspecv2.Operation) {
	if iface.Methods == nil {
		return nil
	}

	for _, method := range iface.Methods.List {
		result = append(result, extractOperation(pass, method))
	}
	return
}

func extractOperation(pass *analysis.Pass, method *ast.Field) *modspecv2.Operation {
	// must be fn expression on interface;
	// other types of expressions are embedded
	// not supported at the moment, we shouldn't care

	fnt, ok := method.Type.(*ast.FuncType)
	if !ok {
		pass.Reportf(method.Pos(), "expected func type")
		return nil
	}
	return &modspecv2.Operation{
		Name:       tryName(method),
		Method:     method,
		Doc:        extractDoc(pass, method.Doc),
		Params:     extractTypesFromFieldList(pass, fnt.Params),
		Results:    extractTypesFromFieldList(pass, fnt.Results),
		Decorators: extractDecorators(pass, method.Doc),
	}
}

func extractDecorators(pass *analysis.Pass, doc *ast.CommentGroup) decorators.List {
	if doc == nil {
		return nil
	}
	dir, err := decorators.FromCommentGroup(doc)
	pass.Reportf(doc.Pos(), "%v", err)
	return dir
}

func extractDoc(pass *analysis.Pass, doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}

	var lines []string
	for _, comment := range doc.List {
		lines = append(lines, comment.Text)
	}

	return strings.Join(lines, "\n")
}

func tryName(param *ast.Field) string {
	return lo.FromPtr(lo.FirstOrEmpty(param.Names)).Name
}

func extractTypeFromField(pass *analysis.Pass, field *ast.Field) modspecv2.Type {
	return modspecv2.Type{
		Name:  tryName(field),
		Field: field,
	}
}

func extractTypesFromFieldList(pass *analysis.Pass, fields *ast.FieldList) (result []modspecv2.Type) {
	// TODO: types of CTX can vary between different service types
	// for example, we shouldn't be using context.Context in workflows
	for _, field := range fields.List {
		result = append(result, extractTypeFromField(pass, field))
	}
	return
}
