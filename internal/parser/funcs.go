package parser

import (
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"go/ast"
)

type FuncDecl struct {
	*ast.FuncDecl
	Directives decorators.List
}

func funcsFromDecls(decls []ast.Decl) (result []*FuncDecl, err error) {
	for _, decl := range decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			fn := &FuncDecl{FuncDecl: funcDecl}

			fn.Directives, err = decorators.FromCommentGroup(funcDecl.Doc)
			if err != nil {
				return
			}

			result = append(result, fn)
		}
	}
	return
}
