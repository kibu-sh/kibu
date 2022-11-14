package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"go/ast"
)

type FuncDecl struct {
	*ast.FuncDecl
	Directives directive.List
}

func funcsFromDecls(decls []ast.Decl) (result []*FuncDecl, err error) {
	for _, decl := range decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			fn := &FuncDecl{FuncDecl: funcDecl}

			fn.Directives, err = directive.FromCommentGroup(funcDecl.Doc)
			if err != nil {
				return
			}

			result = append(result, fn)
		}
	}
	return
}
