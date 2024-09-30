package kibuwire

import (
	"fmt"
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"reflect"
)

type Group struct {
	Name   string
	Import string
}

func (g *Group) FQN() string {
	return fmt.Sprintf("%s.%s", g.Import, g.Name)
}

type Provider struct {
	Symbol       ast.Decl
	ProviderLine decorators.Line
	Decorators   decorators.List
	GoPackage    *types.Package
	Group        *Group
}

type ProviderList []*Provider

func (list ProviderList) Len() int {
	return len(list)
}

func (list ProviderList) Filter(filter func(p *Provider) bool) ProviderList {
	var result ProviderList
	for _, p := range list {
		if filter != nil && filter(p) {
			result = append(result, p)
		}
	}
	return result
}

func (list ProviderList) First() (*Provider, bool) {
	if len(list) == 0 {
		return nil, false
	}
	return list[0], true
}

func (list ProviderList) Find(predicate func(p *Provider) bool) (*Provider, bool) {
	return list.Filter(predicate).First()
}

func (list ProviderList) GroupBy(groupBy func(p *Provider) string) *orderedmap.OrderedMap[string, ProviderList] {
	result := orderedmap.New[string, ProviderList]()
	for _, p := range list {
		group := groupBy(p)
		// skip providers that are not intended to be grouped
		if group == "" {
			continue
		}

		agg, ok := result.Get(group)
		if !ok {
			result.Set(group, ProviderList{p})
			continue
		}

		result.Set(group, append(agg, p))
	}
	return result
}

func FilterByGroupName(name string) func(p *Provider) bool {
	return func(p *Provider) bool {
		return p.Group != nil && p.Group.Name == name
	}
}

func GroupByName() func(p *Provider) string {
	return func(p *Provider) string {
		return p.Group.Name
	}
}

func GroupByFQN() func(p *Provider) string {
	return func(p *Provider) string {
		if p.Group == nil {
			return ""
		}

		return p.Group.FQN()
	}
}

var resultType = reflect.TypeOf((ProviderList)(nil))

func FromPass(pass *analysis.Pass) (ProviderList, bool) {
	result, ok := pass.ResultOf[Analyzer].(ProviderList)
	return result, ok
}

var Analyzer = &analysis.Analyzer{
	Name:             "kibuwire",
	Doc:              "Analyzes go source code for kibu provider annotations",
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	ResultType:       resultType,
	RunDespiteErrors: true,
	Run:              run,
}

var (
	IsKibu         = decorators.HasPrefix("kibu")
	IsKibuProvider = decorators.HasPrefix("kibu:provider")

	IsKibuWorkflow       = decorators.HasPrefix("kibu:workflow")
	IsKibuWorkflowUpdate = decorators.HasPrefix("kibu:workflow:update")
	IsKibuWorkflowQuery  = decorators.HasPrefix("kibu:workflow:query")
	IsKibuWorkflowSignal = decorators.HasPrefix("kibu:workflow:signal")
	IsKibuWorkflowExec   = decorators.HasPrefix("kibu:workflow:execute")

	IsKibuActivity       = decorators.HasPrefix("kibu:activity")
	IsKibuActivityMethod = decorators.HasPrefix("kibu:activity:method")

	IsKibuService       = decorators.HasPrefix("kibu:service")
	IsKibuServiceMethod = decorators.HasPrefix("kibu:service:method")
)

func run(pass *analysis.Pass) (any, error) {
	var result ProviderList
	walk := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
		(*ast.FuncDecl)(nil),
	}

	walk.Preorder(nodeFilter, func(n ast.Node) {
		decl, ok := n.(ast.Decl)
		if !ok {
			return
		}

		var doc *ast.CommentGroup
		switch node := decl.(type) {
		case *ast.GenDecl:
			doc = node.Doc
		case *ast.FuncDecl:
			doc = node.Doc
		}

		decor, err := decorators.FromCommentGroup(doc)
		if err != nil {
			pass.Reportf(n.Pos(), "failed to parse directive: %v", err)
			return
		}

		providerLine, found := decor.Find(IsKibuProvider)
		if !found {
			return
		}

		result = append(result, &Provider{
			Symbol:       decl,
			ProviderLine: providerLine,
			GoPackage:    pass.Pkg,
			Decorators:   decor,
			Group:        groupFromProviderOptions(providerLine.Options),
		})

	})

	return result, nil
}

func groupFromProviderOptions(options *decorators.OptionList) *Group {
	group, ok := options.GetOne("group", "")
	if !ok {
		return nil
	}

	importPath, _ := options.GetOne("import", "")

	return &Group{
		Name:   group,
		Import: importPath,
	}
}
