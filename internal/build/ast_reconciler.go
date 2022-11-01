package build

import (
	"github.com/discernhq/devx/internal/codedef"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
)

func reconcileDrift(entry string, mod codedef.Module) (err error) {
	var fset = token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, entry, nil, parser.AllErrors)
	if err != nil {
		return
	}

	for _, pkg := range pkgs {
		for f, astFile := range pkg.Files {
			updatedSyntax := astutil.Apply(astFile, func(cursor *astutil.Cursor) bool {
				if cursor == nil {
					return true
				}

				for _, wf := range mod.Worker.Workflows {
					if rec := matchReceiverMethod(cursor.Node(), "Workflows"); rec != nil {
						updateFieldIdentifier(rec.Type.Params, 1, wf.Request.Name)
						updateFieldIdentifier(rec.Type.Results, 0, wf.Response.Name)
					}
				}

				for _, act := range mod.Worker.Activities {
					if rec := matchReceiverMethod(cursor.Node(), "Activities"); rec != nil {
						updateFieldIdentifier(rec.Type.Params, 1, act.Request.Name)
						updateFieldIdentifier(rec.Type.Results, 0, act.Response.Name)
					}
				}

				return true
			}, nil)

			var tmpF *os.File
			tmp := f + ".tmp"
			tmpF, err = os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return
			}

			err = printer.Fprint(tmpF, fset, updatedSyntax)
			if err != nil {
				return
			}

			err = tmpF.Close()
			if err != nil {
				return
			}

			err = os.Rename(tmp, f)
			if err != nil {
				return
			}

		}
	}

	return
}

// TODO: add support for generics
func updateFieldIdentifier(fields *ast.FieldList, idx int, name string) {
	if id, ok := fields.List[idx].Type.(*ast.Ident); ok {
		id.Name = name
	}
}

func matchReceiverMethod(node ast.Node, recID string) *ast.FuncDecl {
	switch t := node.(type) {
	case *ast.FuncDecl:
		if t.Recv != nil {
			for _, f := range t.Recv.List {
				if exp, ok := f.Type.(*ast.StarExpr); ok {
					if id, ok := exp.X.(*ast.Ident); ok && id.Name == recID {
						return t
					}
				}
			}
		}
	}
	return nil
}
