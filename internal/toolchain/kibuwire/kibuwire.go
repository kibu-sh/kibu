package kibuwire

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"github.com/samber/lo"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"path/filepath"
	"reflect"
)

type Provider struct {
	Symbol       ast.Decl
	ProviderLine decorators.Line
	Decorators   decorators.List
	GoPackage    *types.Package
	Group        *Group
}

func (p *Provider) SymbolName() string {
	switch decl := p.Symbol.(type) {
	case *ast.FuncDecl:
		return decl.Name.Name
	case *ast.GenDecl:
		return decl.Specs[0].(*ast.TypeSpec).Name.Name
	default:
		panic(fmt.Errorf("unsupported provider symbol type: %T", decl))
	}
}

type Group struct {
	Name   string
	Import string
}

func (g *Group) FQN() string {
	return fmt.Sprintf("%s.%s", g.Import, g.Name)
}

var _ modspecv2.Artifact = (*Artifact)(nil)

type Artifact struct {
	Providers ProviderList
	*modspecv2.PackageArtifact
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

var resultType = reflect.TypeOf((*Artifact)(nil))

func FromPass(pass *analysis.Pass) (*Artifact, bool) {
	result, ok := pass.ResultOf[Analyzer].(*Artifact)
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
	file := modspecv2.NewJenFileFromPackage(pass.Pkg)
	walk := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	var result = &Artifact{
		PackageArtifact: modspecv2.NewPackageArtifact(file, pass, ".wire"),
		Providers:       ProviderList{},
	}

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

		result.Providers = append(result.Providers, &Provider{
			Symbol:       decl,
			ProviderLine: providerLine,
			GoPackage:    pass.Pkg,
			Decorators:   decor,
			Group:        groupFromProviderOptions(providerLine.Options),
		})

	})

	generateWireSet(file, result.Providers)

	if result.Providers.Len() == 0 {
		return nil, nil
	}

	return result, nil
}

func generateWireSet(f *jen.File, providers ProviderList) {
	f.Var().Id("WireSet").Op("=").Qual("github.com/google/wire", "NewSet").
		CustomFunc(modspecv2.MultiLineParen(), func(g *jen.Group) {
			for _, provider := range providers {
				generateSymbolForFunc(g, provider)
				generateSymbolForStruct(g, provider)
			}
		})
}

const wireImportName = "github.com/google/wire"

func generateSymbolForStruct(g *jen.Group, provider *Provider) {
	genDecl, ok := provider.Symbol.(*ast.GenDecl)
	if !ok {
		return
	}

	if len(genDecl.Specs) == 0 {
		return
	}

	_, ok = genDecl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return
	}

	// generate pointer wire struct provider
	// wire.Struct(new(ServiceController), "*"),
	g.Qual(wireImportName, "Struct").CallFunc(func(g *jen.Group) {
		g.New(jen.Id(provider.SymbolName()))
		g.Lit("*")
	})
	return
}

func generateSymbolForFunc(g *jen.Group, provider *Provider) {
	_, ok := provider.Symbol.(*ast.FuncDecl)
	if !ok {
		return
	}

	g.Id(provider.SymbolName())
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

var _ modspecv2.Artifact = (*Module)(nil)

type Module struct {
	file *jen.File
	out  string
}

func (k *Module) File() *jen.File {
	return k.file
}

func (k *Module) OutputPath() string {
	return filepath.Join(k.out, "kibuwire/kibuwire.gen.go")
}

func buildKibuWireModule(outDir string, artifacts []*Artifact) modspecv2.Artifact {
	var module = &Module{
		out:  outDir,
		file: jen.NewFile("kibuwire"),
	}

	module.file.HeaderComment("Code generated by kibu. DO NOT EDIT.")
	providers := lo.Reduce(artifacts, func(list ProviderList, a *Artifact, _ int) ProviderList {
		return append(list, a.Providers...)
	}, ProviderList{})

	buildGroupedProviderSet(module.file, providers)
	buildSuperSet(module.file, providers)

	return module
}

// groupQualName returns the fully qualified name of the group interface
// github.com/kibu-sh/kibu/pkg/transport/httpx.HandlerFactory → httpx.HandlerFactory
func groupQualName(group *Group) *jen.Statement {
	return jen.Qual(group.Import, group.Name)
}

// groupQualParam returns the fully qualified name of the group interface as a slice
// github.com/kibu-sh/kibu/pkg/transport/httpx.HandlerFactory → []httpx.HandlerFactory
func groupQualParam(group *Group) *jen.Statement {
	return jen.Index().Add(groupQualName(group))
}

// groupProviderName returns the name of the group provider
// github.com/kibu-sh/kibu/pkg/transport/httpx.HandlerFactory → HttpxHandlerFactoryGroup
func groupProviderName(group *Group) string {
	return lo.PascalCase(groupQualName(group).GoString()) + "Group"
}

// providerQualName returns the fully qualified name of the provider
// example.com/billingv1.ServiceController → billingv1.ServiceController
func providerQualName(provider *Provider) *jen.Statement {
	return jen.Qual(provider.GoPackage.Path(), provider.SymbolName())
}

// providerQualParam returns the fully qualified name of the provider as a pascal case field name
// example.com/billingv1.ServiceController → Billingv1ServiceController
func groupedProviderFieldName(provider *Provider) string {
	return lo.PascalCase(providerQualName(provider).GoString())
}

// groupProviderFuncName returns the name of the group provider function
// github.com/kibu-sh/kibu/pkg/transport/httpx.HandlerFactory → HttpxHandlerFactoryGroupProvider
func groupProviderFuncName(group *Group) string {
	return lo.PascalCase(groupQualName(group).GoString()) + "Provider"
}

func buildGroupedProviderSet(f *jen.File, providers ProviderList) {
	grouped := providers.GroupBy(GroupByFQN())

	for key := range grouped.KeysFromOldest() {
		groupedList, _ := grouped.Get(key)
		group := groupedList[0].Group
		f.Type().Id(groupProviderName(group)).StructFunc(func(g *jen.Group) {
			for _, provider := range groupedList {
				g.Id(groupedProviderFieldName(provider)).Add(providerQualName(provider))
			}
		})

		f.Func().Id(groupProviderFuncName(group)).Params(
			jen.Id("group").Op("*").Id(groupProviderName(group)),
		).Params(groupQualParam(group)).BlockFunc(func(g *jen.Group) {
			g.ReturnFunc(func(g *jen.Group) {
				g.Add(groupQualParam(group)).CustomFunc(modspecv2.MultiLineCurly(), func(g *jen.Group) {
					for _, provider := range groupedList {
						g.Id("group").Dot(groupedProviderFieldName(provider))
					}
				})
			})
		})
	}

	f.Var().Id("GroupSet").Op("=").Qual(wireImportName, "NewSet").
		CustomFunc(modspecv2.MultiLineParen(), func(g *jen.Group) {
			for key := range grouped.KeysFromOldest() {
				groupedList, _ := grouped.Get(key)
				group := groupedList[0].Group
				g.Qual(wireImportName, "Struct").Call(jen.New(jen.Id(groupProviderName(group))), jen.Lit("*"))
				g.Id(groupProviderFuncName(group))
			}
		})
}

func buildSuperSet(f *jen.File, providers ProviderList) {
	f.Var().Id("SuperSet").Op("=").Qual(wireImportName, "NewSet").
		CustomFunc(modspecv2.MultiLineParen(), func(g *jen.Group) {
			// includes the group set
			g.Id("GroupSet")

			for _, provider := range providers.Filter(FilterUngrouped) {
				g.Qual(provider.GoPackage.Path(), "WireSet")
			}
		})
}

func FilterUngrouped(p *Provider) bool {
	return p.Group == nil
}
