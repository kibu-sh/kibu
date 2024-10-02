package dagv2

import (
	"golang.org/x/sync/syncmap"
)

type AcyclicGraph[T comparable] struct {
	Vertices syncmap.Map
	Edges    syncmap.Map
}

func Clone[T comparable](g *AcyclicGraph[T]) *AcyclicGraph[T] {
	// clone the graph
	panic("implement me")
}

func (g *AcyclicGraph[T]) AddVertex(v T) {
	g.Vertices.Store(v, v)
}

func (g *AcyclicGraph[T]) AddEdge(from, to T) {
	g.Edges.Store(from, to)
}

func (g *AcyclicGraph[T]) TransitiveReduction(target T) AcyclicGraph[T] {
	// clone the graph and transitively reduce it
	// transitively reduce the graph for target T
	// we should not make any unnecessary hops
	panic("implement me")
}

type SortFunc[T any] func(a T, b T) bool

func (g *AcyclicGraph[T]) TopologicalSort(sortFunc SortFunc[T]) (result []T) {
	// implement a topological sort
	// return the nodes in a topical order
	return
}

func (g *AcyclicGraph[T]) Validate() error {
	// check if the graph is cyclic
	// if so return cyclic error
	return nil
}

type WalkFunc[T any] func(T)

func (g *AcyclicGraph[T]) Walk(target T, visitor WalkFunc[T]) {
	// implement a depth-first search using transitive reduction
	// attempting to reach target T
	return
}
