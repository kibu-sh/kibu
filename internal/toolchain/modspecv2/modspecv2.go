package modspecv2

import (
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"path/filepath"
)

type Artifact interface {
	// File returns a generated jen.File that will be written to disk
	File() *jen.File

	// OutputPath returns a string to a file and its extension relative to the module root
	// i.e., example.com/foo/bar/baz.go -> foo/bar/baz.gen.go
	OutputPath() string
}

type PackageArtifact struct {
	file *jen.File
	pass *analysis.Pass
	ext  string
}

func (p *PackageArtifact) File() *jen.File {
	return p.file
}

func (p *PackageArtifact) OutputPath() string {
	return filepath.Join(RelPathFromPass(p.pass), GenGoExt(p.pass.Pkg.Name()+p.ext))
}

func NewPackageArtifact(file *jen.File, pass *analysis.Pass, ext string) *PackageArtifact {
	return &PackageArtifact{
		file: file,
		pass: pass,
		ext:  ext,
	}
}

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
