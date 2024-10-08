package codegen

import (
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
)

type PipelineOptions struct {
	FileSet        FileSet
	GenerateParams GenerateParams
	Services       []*parser.Service
	Workers        []*parser.Worker
	Providers      []*parser.Provider
	Middleware     []*parser.Middleware
}

type PipelineFunc func(opts *PipelineOptions) (err error)

type Pipeline []PipelineFunc

type GenerateParams struct {
	Dir       string
	Pipeline  Pipeline
	Patterns  []string
	OutputDir string
}

type FilePath string
type PackageName string
type FileSet map[FilePath]*jen.File

func NewFileSet() FileSet {
	return make(FileSet)
}

func (fs FileSet) NewFile(filePath FilePath, packageName PackageName) *jen.File {
	f := jen.NewFile(string(packageName))
	f.HeaderComment("Code generated by kibu. DO NOT EDIT.")
	fs[filePath] = f
	return f
}

func (fs FileSet) Get(filePath FilePath, packageName PackageName) (f *jen.File) {
	f, ok := fs[filePath]
	if ok {
		return
	}
	return fs.NewFile(filePath, packageName)
}

func NewPipelineOptions(
	fset FileSet,
	genParams GenerateParams,
	pkgList map[parser.PackagePath]*parser.Package,
) (opts *PipelineOptions) {
	opts = &PipelineOptions{
		FileSet:        fset,
		GenerateParams: genParams,
	}

	for _, pkg := range pkgList {
		opts.Services = append(opts.Services, toSlice(pkg.Services)...)
		opts.Workers = append(opts.Workers, toSlice(pkg.Workers)...)
		opts.Providers = append(opts.Providers, toSlice(pkg.Providers)...)
		opts.Middleware = append(opts.Middleware, toSlice(pkg.Middleware)...)
	}

	sort.Slice(opts.Services, sortByID(opts.Services))
	sort.Slice(opts.Workers, sortByID(opts.Workers))
	sort.Slice(opts.Providers, sortByID(opts.Providers))
	sort.Slice(opts.Middleware, sortByPos(opts.Middleware))

	return
}

func toSlice[K comparable, V any](m map[K]V) (result []V) {
	for _, v := range m {
		result = append(result, v)
	}
	return
}

type idSortable interface {
	ID() string
}

func sortByID[V idSortable](list []V) func(i, j int) bool {
	return func(i, j int) bool {
		return list[i].ID() < list[j].ID()
	}
}

type posSortable interface {
	Pos() token.Pos
}

func sortByPos[v posSortable](list []v) func(i, j int) bool {
	return func(i, j int) bool {
		return list[i].Pos() < list[j].Pos()
	}
}

func Generate(params GenerateParams) (err error) {
	fset := NewFileSet()
	packageDir := kibugenPackageDir(params)

	_ = os.RemoveAll(packageDir)
	if err = os.MkdirAll(packageDir, os.ModePerm); err != nil {
		return
	}

	pkgList, err := parser.ExperimentalParse(parser.ExperimentalParseOpts{
		Dir:      params.Dir,
		Patterns: params.Patterns,
	})
	if err != nil {
		return
	}

	pipelineOpts := NewPipelineOptions(fset, params, pkgList)

	for _, generateFunc := range params.Pipeline {
		if err = generateFunc(pipelineOpts); err != nil {
			return
		}
	}

	for filePath, file := range fset {
		fp := string(filePath)

		if err = os.MkdirAll(filepath.Dir(fp), os.ModePerm); err != nil {
			return
		}

		if err = file.Save(fp); err != nil {
			return
		}
	}

	return
}

func packageScopedFilePath(pkg *parser.Package) (FilePath, PackageName) {
	return FilePath(filepath.Join(string(pkg.Path), pkg.GoPackage.Name+".gen.go")), PackageName(pkg.GoPackage.Name)
}

func DefaultPipeline() Pipeline {
	return Pipeline{
		BuildServiceHTTPHandlerFactories,
		BuildWorkerProxies,
		BuildHTTPHandlerProviders,
		BuildWorkerProviders,
		BuildMiddlewareProvider,
		BuildWireSet,
		BuildOpenAPISpec,
	}
}
