package parser

import (
	"github.com/kibu-sh/kibu/internal/parser/directive"
	"github.com/pkg/errors"
	"go/ast"
	"go/types"
	"strings"
)

type Method struct {
	*TypeMeta
	Name       string
	Directives directive.List
	Request    *Var
	Response   *Var
}

type WorkerType string

type Worker struct {
	*TypeMeta
	Name       string
	Type       WorkerType
	TaskQueue  string
	Methods    map[*ast.Ident]*Method
	Directives directive.List
}

var (
	WorkflowType = WorkerType("workflow")
	ActivityType = WorkerType("activity")
)

func NewWorker(name, queue string, meta *TypeMeta) *Worker {
	return &Worker{
		Name:      name,
		TaskQueue: queue,
		Methods:   make(map[*ast.Ident]*Method),
		TypeMeta:  meta,
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

		taskQueue, _ := dir.Options.GetOne("task_queue", "default")
		wrk := NewWorker(ident.Name, taskQueue, NewTypeMeta(ident, obj, pkg))

		wrk.Directives = dirs

		if !dir.Options.HasOneOf("workflow", "activity") {
			err = errors.Errorf("worker must specify one of (activity or workflow) %s",
				wrk.Position().String(),
			)
			return
		}

		if dir.Options.Has("workflow") {
			wrk.Type = WorkflowType
		} else {
			wrk.Type = ActivityType
		}

		wrk.Methods, err = collectWorkerMethods(pkg, n)
		if err != nil {
			return
		}
		pkg.Workers[ident] = wrk

		return
	}
}

func collectWorkerMethods(pkg *Package, n *types.Named) (methods map[*ast.Ident]*Method, err error) {
	methods = make(map[*ast.Ident]*Method)

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

		_, isMethod := dirs.Find(directive.OneOf(
			directive.HasKey("devx", "workflow"),
			directive.HasKey("devx", "activity"),
		))

		if !isMethod {
			continue
		}

		sig := m.Type().(*types.Signature)

		if sig.Params().Len() != 2 || sig.Results().Len() != 2 {
			err = errors.Errorf("%s \n\tworker methods must match func %s(ctx context.Context, req Req) (res Res, err error)", pkg.GoPackage.Fset.Position(ident.Pos()).String(), ident.Name)
			return
		}

		req := sig.Params().At(1)
		res := sig.Results().At(0)

		method := &Method{
			Name:       ident.Name,
			Directives: dirs,
			Request:    &Var{Var: req},
			Response:   &Var{Var: res},
			TypeMeta:   NewTypeMeta(ident, m, pkg),
		}

		methods[ident] = method
	}
	return
}

func getTypeNameWithoutPackage(pkg *Package, v *types.Var) string {
	return strings.Replace(
		strings.Replace(v.Type().String(), pkg.GoPackage.PkgPath, "", 1), ".", "", 1)
}
