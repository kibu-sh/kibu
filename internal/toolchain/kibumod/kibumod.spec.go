package kibumod

import (
	"github.com/kibu-sh/kibu/internal/parser/directive"
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
	field *ast.Field
}

type Operation struct {
	Name       string
	Params     []Type
	Results    []Type
	method     *ast.Field
	Doc        string
	Decorators directive.List
}

type Service struct {
	Name       string
	Operations []*Operation
	Decorators directive.List
	Doc        string
	decl       *ast.GenDecl
	iface      *ast.InterfaceType
	tspec      *ast.TypeSpec
}
