package pipeline

import (
	"errors"
	"github.com/kibu-sh/kibu/internal/toolchain/dag"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
	"sync"
)

type Runner struct {
	graph *dag.AcyclicGraph
}

var GraphError = errors.New("graph error")
var GraphWalkError = errors.Join(GraphError, errors.New("walk error"))

func (r *Runner) Execute(pass *analysis.Pass) error {
	if err := r.graph.Validate(); err != nil {
		return errors.Join(GraphError, err)
	}

	if err := r.graph.Walk(r.walkFunc(pass)); err != nil {
		return errors.Join(GraphWalkError, err)
	}

	return nil
}

func (r *Runner) walkFunc(pass *analysis.Pass) func(v dag.Vertex) error {
	mtx := new(sync.Mutex)
	return func(v dag.Vertex) error {
		analyzer := v.(*analysis.Analyzer)
		pass.Analyzer = analyzer

		result, err := analyzer.Run(pass)
		if err != nil {
			return err
		}

		mtx.Lock()
		pass.ResultOf[analyzer] = result
		mtx.Unlock()
		return nil
	}
}

func buildGraph(graph *dag.AcyclicGraph, analyzers ...*analysis.Analyzer) {
	for _, node := range analyzers {
		graph.Add(node)
		for _, dep := range node.Requires {
			graph.Add(dep)
			graph.Connect(dag.BasicEdge(node, dep))
			buildGraph(graph, dep)
		}
	}
}

func NewRunner(analyzers ...*analysis.Analyzer) *Runner {
	graph := &dag.AcyclicGraph{}
	buildGraph(graph, analyzers...)
	return &Runner{
		graph: graph,
	}
}

func NewAnalysisPass(pkg *packages.Package, store FactStore) *analysis.Pass {
	return &analysis.Pass{
		Fset:              pkg.Fset,
		Files:             pkg.Syntax,
		OtherFiles:        pkg.OtherFiles,
		IgnoredFiles:      pkg.IgnoredFiles,
		Pkg:               pkg.Types,
		TypesInfo:         pkg.TypesInfo,
		TypesSizes:        pkg.TypesSizes,
		TypeErrors:        pkg.TypeErrors,
		Module:            maybeModule(pkg),
		Report:            store.Report,
		ResultOf:          make(map[*analysis.Analyzer]any),
		ImportObjectFact:  store.ImportObjectFact,
		ExportObjectFact:  store.ExportObjectFact,
		ImportPackageFact: store.ImportPackageFact,
		ExportPackageFact: store.ExportPackageFact,
		AllObjectFacts:    store.AllObjectFacts,
		AllPackageFacts:   store.AllPackageFacts,
	}
}

func maybeModule(pkg *packages.Package) *analysis.Module {
	if pkg.Module == nil {
		return nil
	}
	return &analysis.Module{
		Path:      pkg.Module.Path,
		Version:   pkg.Module.Version,
		GoVersion: pkg.Module.GoVersion,
	}
}

func Run(config *Config) error {
	pkgs, err := LoadPackages(config)
	if err != nil {
		return err
	}

	runner := NewRunner(config.Analyzers...)
	for _, pkg := range pkgs {
		pass := NewAnalysisPass(pkg, config.FactStore)
		if err := runner.Execute(pass); err != nil {
			return err
		}
	}
	return nil
}
