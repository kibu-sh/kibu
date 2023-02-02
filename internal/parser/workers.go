package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"github.com/pkg/errors"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

type Method struct {
	Name       string
	Directives directive.List
	Request    *Var
	Response   *Var
}

type Worker struct {
	Name       string
	Type       string
	TaskQueue  string
	Methods    map[string]*Method
	Directives directive.List
	File       *token.File
	Position   token.Position
}

func NewWorker(name, queue string) *Worker {
	return &Worker{
		Name:      name,
		TaskQueue: queue,
		Methods:   make(map[string]*Method),
	}
}

func collectWorkers(pkg *Package) defMapperFunc {
	return func(ident *ast.Ident, obj types.Object) (err error) {
		_, ok := obj.(*types.TypeName)
		if !ok {
			return
		}

		n, ok := obj.Type().(*types.Named)
		if !ok {
			return
		}

		_, ok = n.Underlying().(*types.Struct)
		if !ok {
			return
		}

		dirs, ok := pkg.directiveCache[ident]
		if !ok {
			return
		}

		// TODO: inject logger
		// fmt.Printf("inspecting %s\n", n.String())
		// skip this struct if it doesn't have the service directive
		dir, isWorker := dirs.Find(directive.HasKey("devx", "worker"))
		if !isWorker {
			return
		}

		taskQueue, _ := dir.Options.Find("task_queue", "default")
		wrk := NewWorker(ident.Name, taskQueue)

		wrk.Directives = dirs
		wrk.File = pkg.GoPackage.Fset.File(ident.Pos())
		wrk.Position = wrk.File.Position(ident.Pos())

		if !dir.Options.HasOneOf("workflow", "activity") {
			err = errors.Errorf("worker must specify one of (activity or workflow) %s:%d",
				wrk.Position.Filename, wrk.Position.Line)
			return
		}

		if dir.Options.Has("workflow") {
			wrk.Type = "workflow"
		} else {
			wrk.Type = "activity"
		}

		wrk.Methods, err = collectWorkerMethods(pkg, n)
		if err != nil {
			return
		}
		pkg.Workers[ident] = wrk

		return
	}
}

func collectWorkerMethods(pkg *Package, n *types.Named) (methods map[string]*Method, err error) {
	methods = make(map[string]*Method)

	for i := 0; i < n.NumMethods(); i++ {
		m := n.Method(i)
		ident, ok := pkg.funcIdCache[m]
		if !ok {
			continue
		}

		dirs, ok := pkg.directiveCache[ident]
		if !ok {
			continue
		}

		sig := m.Type().(*types.Signature)
		req := sig.Params().At(1)
		res := sig.Results().At(0)

		_, isMethod := dirs.Find(directive.OneOf(
			directive.HasKey("devx", "workflow"),
			directive.HasKey("devx", "activity"),
		))

		if !isMethod {
			return
		}

		ep := &Method{
			Name:       ident.Name,
			Directives: dirs,
			Request: &Var{
				Name: req.Name(),
				Type: getTypeNameWithoutPackage(pkg, req),
			},
			Response: &Var{
				Name: res.Name(),
				Type: getTypeNameWithoutPackage(pkg, res),
			},
		}

		methods[ident.Name] = ep
	}
	return
}

func getTypeNameWithoutPackage(pkg *Package, v *types.Var) string {
	return strings.Replace(
		strings.Replace(v.Type().String(), pkg.GoPackage.PkgPath, "", 1), ".", "", 1)

}
