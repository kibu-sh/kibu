package parser

type Workflow struct{}
type Activity struct{}

type Worker struct {
	Workflows  map[string]*Workflow
	Activities map[string]*Activity
}

// func collectWorkflows(f *Package) defMapperFunc {
// 	return func(n ast.Node) (found bool, err error) {
// 		// if s, found := n.(*ast.StructType); found {
// 		// 	str := &Worker{}
// 		//
// 		//
// 		// 	f.Workers[str.Name] = str
// 		// }
// 		return
// 	}
// }
