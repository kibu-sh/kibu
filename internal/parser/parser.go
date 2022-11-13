package parser

import (
	"go/parser"
	"go/token"
)

func ParseDir(dir string) (pkgs map[string]*Package, err error) {
	fs := token.NewFileSet()
	p, err := parser.ParseDir(fs, dir, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return
	}

	pkgs = pkgsFromAst(p)
	return
}
