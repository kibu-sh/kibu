package kibumod

import (
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
)

type Package struct {
	Name     string
	Services []*Service
	GoPkg    *types.Package
	GoModule *analysis.Module
}

// Type represents an abstract type
// A can be any native Go object, we use this to identify properties or func arguments/results
type Type struct {
	Name  string
	Field *ast.Field
}

type Operation struct {
	Name       string
	Params     []Type
	Results    []Type
	Method     *ast.Field
	Doc        string
	Decorators decorators.List
}

type Service struct {
	Name       string
	Operations []*Operation
	Decorators decorators.List
	Doc        string
	Decl       *ast.GenDecl
	Iface      *ast.InterfaceType
	Tspec      *ast.TypeSpec
}
