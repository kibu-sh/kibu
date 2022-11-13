package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"go/ast"
)

type Application struct {
	Packages map[string]*Package
	Files    map[string]*File
}

type Package struct {
	pkg  *ast.Package
	Name string
}

type File struct {
	Path string
	file *ast.File
	pkg  *Package
}

type Decl struct {
	decl ast.Decl
}

type StructDecl struct {
	spec       *ast.TypeSpec
	decl       *ast.GenDecl
	Directives directive.List
	Methods    map[string]*FuncDecl
}

func (s *StructDecl) Name() string {
	return s.spec.Name.Name
}

type FuncDecl struct {
	decl       *ast.FuncDecl
	Directives directive.List
}

func (f *FuncDecl) Name() string {
	return f.decl.Name.Name
}

func pkgsFromAst(pkgs map[string]*ast.Package) (result map[string]*Package) {
	result = make(map[string]*Package, len(pkgs))
	for path, pkg := range pkgs {
		result[path] = &Package{
			pkg:  pkg,
			Name: pkg.Name,
		}
	}
	return
}

func filesFromPackage(p *Package) (result map[string]*File) {
	result = make(map[string]*File, len(p.pkg.Files))
	for path, file := range p.pkg.Files {
		result[path] = &File{
			Path: path,
			file: file,
			pkg:  p,
		}
	}
	return
}

func declsFromFile(f *File) (result []*Decl) {
	for _, decl := range f.file.Decls {
		result = append(result, &Decl{
			decl: decl,
		})
	}
	return
}

func structsFromDecls(decls []*Decl) (result map[string]*StructDecl, err error) {
	result = make(map[string]*StructDecl)
	for _, decl := range decls {
		switch dt := decl.decl.(type) {
		case *ast.GenDecl:
			for _, spec := range dt.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					switch s.Type.(type) {
					case *ast.StructType:
						str := &StructDecl{
							spec: s,
							decl: dt,
						}

						str.Directives, err = directive.FromCommentGroup(str.decl.Doc)
						if err != nil {
							return
						}

						result[s.Name.Name] = str
					}
				}
			}
		}
	}
	return
}
