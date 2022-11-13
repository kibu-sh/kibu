package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
	"go/ast"
	"testing"
)

func TestParser(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"parse": func(ts *testscript.TestScript, neg bool, args []string) {
				pkgs, err := ParseDir(args[0])
				ts.Check(err)

				for _, p := range pkgs {
					for _, f := range filesFromPackage(p) {
						var funcs []*FuncDecl
						var structs map[string]*StructDecl
						decls := declsFromFile(f)
						structs, err = structsFromDecls(decls)
						ts.Check(err)

						funcs, err = funcsFromDecls(decls)
						ts.Check(err)

						require.Len(t, structs, 1)
						require.Len(t, funcs, 1)
					}
				}
			},
		},
	})
}

func funcsFromDecls(decls []*Decl) (result []*FuncDecl, err error) {
	for _, decl := range decls {
		if funcDecl, ok := decl.decl.(*ast.FuncDecl); ok {
			fn := &FuncDecl{
				decl: funcDecl,
			}

			fn.Directives, err = directive.FromCommentGroup(funcDecl.Doc)
			if err != nil {
				return
			}

			result = append(result, fn)
		}
	}
	return
}
