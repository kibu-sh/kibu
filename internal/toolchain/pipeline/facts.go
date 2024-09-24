package pipeline

import (
	"go/types"
	"golang.org/x/tools/go/analysis"
)

type FactStore interface {
	// ReadFile returns the contents of the named file.
	//
	// The only valid file names are the elements of OtherFiles
	// and IgnoredFiles, and names returned by
	// Fset.File(f.FileStart).Name() for each f in Files.
	//
	// Analyzers must use this function (if provided) instead of
	// accessing the file system directly. This allows a driver to
	// provide a virtualized file tree (including, for example,
	// unsaved editor buffers) and to track dependencies precisely
	// to avoid unnecessary computation.
	ReadFile(filename string) ([]byte, error)

	// Report reports a Diagnostic, a finding about a specific location
	// in the analyzed source code such as a potential mistake.
	// It may be called by the Run function.
	Report(diagnostic analysis.Diagnostic)

	// -- facts --

	// ImportObjectFact retrieves a fact associated with obj.
	// Given a value ptr of type *T, where *T satisfies Fact,
	// ImportObjectFact copies the value to *ptr.
	//
	// ImportObjectFact panics if called after the pass is complete.
	// ImportObjectFact is not concurrency-safe.
	ImportObjectFact(obj types.Object, fact analysis.Fact) bool

	// ImportPackageFact retrieves a fact associated with package pkg,
	// which must be this package or one of its dependencies.
	// See comments for ImportObjectFact.
	ImportPackageFact(pkg *types.Package, fact analysis.Fact) bool

	// ExportObjectFact associates a fact of type *T with the obj,
	// replacing any previous fact of that type.
	//
	// ExportObjectFact panics if it is called after the pass is
	// complete, or if obj does not belong to the package being analyzed.
	// ExportObjectFact is not concurrency-safe.
	ExportObjectFact(obj types.Object, fact analysis.Fact)

	// ExportPackageFact associates a fact with the current package.
	// See comments for ExportObjectFact.
	ExportPackageFact(fact analysis.Fact)

	// AllPackageFacts returns a new slice containing all package
	// facts of the analysis's FactTypes in unspecified order.
	AllPackageFacts() []analysis.PackageFact

	// AllObjectFacts returns a new slice containing all object
	// facts of the analysis's FactTypes in unspecified order.
	AllObjectFacts() []analysis.ObjectFact
}

var _ FactStore = (*NoOpFactStore)(nil)

type NoOpFactStore struct{}

func (n NoOpFactStore) ReadFile(filename string) ([]byte, error) {
	return nil, nil
}

func (n NoOpFactStore) Report(diagnostic analysis.Diagnostic) {
	return
}

func (n NoOpFactStore) ImportObjectFact(obj types.Object, fact analysis.Fact) bool {
	return false
}

func (n NoOpFactStore) ImportPackageFact(pkg *types.Package, fact analysis.Fact) bool {
	return false
}

func (n NoOpFactStore) ExportObjectFact(obj types.Object, fact analysis.Fact) {
	return
}

func (n NoOpFactStore) ExportPackageFact(fact analysis.Fact) {
	return
}

func (n NoOpFactStore) AllPackageFacts() []analysis.PackageFact {
	return nil
}

func (n NoOpFactStore) AllObjectFacts() []analysis.ObjectFact {
	return nil
}
