package kibuenum

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

type parser struct{ pass *analysis.Pass }

func (p parser) report(n ast.Node, category, message string) {
	p.pass.Report(analysis.Diagnostic{
		Pos:      n.Pos(),
		End:      n.End(),
		Category: category,
		Message:  message,
	})
}

func (p parser) span(n ast.Node) Span {
	return Span{
		Start: p.pass.Fset.Position(n.Pos()),
		End:   p.pass.Fset.Position(n.End()),
	}
}

func (p parser) symbol(obj types.Object) Symbol {
	s := Symbol{Name: obj.Name(), Position: p.pass.Fset.Position(obj.Pos())}
	if obj.Pkg() != nil {
		s.PackagePath = obj.Pkg().Path()
	}
	return s
}

func (p parser) parseDeclaration(call *ast.CallExpr, owner *ast.Ident) (Enum, bool) {
	var declaration Enum
	index, ok := ast.Unparen(call.Fun).(*ast.IndexExpr)
	if !ok {
		p.report(call, "enum-shape", "Define requires one explicit enum type argument")
		return declaration, false
	}
	enumType, ok := types.Unalias(p.pass.TypesInfo.TypeOf(index.Index)).(*types.Named)
	if !ok || enumType.Obj().Pkg() != p.pass.Pkg || enumType.TypeParams().Len() != 0 {
		p.report(index.Index, "enum-type", "enum type must be a defined non-generic scalar in this package")
		return declaration, false
	}
	scalar, ok := enumType.Underlying().(*types.Basic)
	if !ok || scalar.Info()&(types.IsString|types.IsInteger) == 0 {
		p.report(index.Index, "enum-type", "enum underlying type must be string or integer")
		return declaration, false
	}
	membersLiteral, memberType, valid := p.parseMemberSlice(call, enumType)
	if !valid {
		return declaration, false
	}
	declaration = Enum{
		Declaration: p.symbol(p.pass.TypesInfo.ObjectOf(owner)),
		Type:        p.symbol(enumType.Obj()),
		Underlying:  scalar.Name(),
		Source:      p.span(call),
	}
	declaration.Members, valid = p.parseMembers(membersLiteral, enumType, memberType)
	return declaration, valid
}

func (p parser) parseMemberSlice(call *ast.CallExpr, enumType *types.Named) (*ast.CompositeLit, *types.Named, bool) {
	if len(call.Args) != 1 || call.Ellipsis.IsValid() {
		p.report(call, "enum-shape", "Define requires one literal member slice")
		return nil, nil, false
	}
	membersLiteral, ok := ast.Unparen(call.Args[0]).(*ast.CompositeLit)
	if !ok {
		p.report(call.Args[0], "enum-shape", "Define requires a literal member slice, not a variable or function call")
		return nil, nil, false
	}
	slice, ok := types.Unalias(p.pass.TypesInfo.TypeOf(membersLiteral)).(*types.Slice)
	if !ok {
		p.report(membersLiteral, "enum-shape", "expected []experimental.Member[Enum]")
		return nil, nil, false
	}
	memberType, ok := types.Unalias(slice.Elem()).(*types.Named)
	if !ok || !isSDKObject(memberType.Obj(), "Member") {
		p.report(membersLiteral, "enum-type", "member slice must use SDK Member with the same enum type")
		return nil, nil, false
	}
	arguments := memberType.TypeArgs()
	if arguments.Len() != 1 || !types.Identical(arguments.At(0), enumType) {
		p.report(membersLiteral, "enum-type", "member slice must use SDK Member with the same enum type")
		return nil, nil, false
	}
	return membersLiteral, memberType, true
}

func (p parser) parseMembers(membersLiteral *ast.CompositeLit, enumType, memberType *types.Named) ([]Member, bool) {
	var members []Member
	seen := map[string]Member{}
	valid := true
	for _, element := range membersLiteral.Elts {
		member, ok := p.parseMember(element, enumType, memberType)
		if !ok {
			valid = false
			continue
		}
		if prior, exists := seen[member.WireValue]; exists {
			message := fmt.Sprintf("duplicate enum wire value %q in %s; first selected as %s at %s",
				member.WireValue, member.Constant.Name, prior.Constant.Name, prior.Source.Start)
			p.report(element, "enum-duplicate-value", message)
			valid = false
			continue
		}
		seen[member.WireValue] = member
		members = append(members, member)
	}
	return members, valid
}

func (p parser) parseMember(expr ast.Expr, enumType, memberType *types.Named) (Member, bool) {
	var member Member
	literal, ok := ast.Unparen(expr).(*ast.CompositeLit)
	if !ok || !types.Identical(p.pass.TypesInfo.TypeOf(literal), memberType) {
		p.report(expr, "enum-shape", "each member must be a keyed SDK Member literal; indexed elements are unsupported")
		return member, false
	}
	member.Source = p.span(literal)
	for _, element := range literal.Elts {
		if !p.parseMemberField(element, enumType, &member) {
			return member, false
		}
	}
	if member.Constant.Name == "" {
		p.report(literal, "enum-value", "member requires an explicit Value constant")
		return member, false
	}
	return member, true
}

// parseMemberField validates the field shape and dispatches its value parser.
func (p parser) parseMemberField(element ast.Expr, enumType *types.Named, member *Member) bool {
	field, ok := element.(*ast.KeyValueExpr)
	if !ok {
		p.report(element, "enum-shape", "member fields must be keyed")
		return false
	}
	key, ok := field.Key.(*ast.Ident)
	if !ok {
		p.report(field, "enum-shape", "member field must be named")
		return false
	}

	switch key.Name {
	case "Value":
		return p.parseValueField(field.Value, enumType, member)
	case "Label", "Description":
		return p.parseMetadataField(field.Value, key.Name, member)
	default:
		p.report(field, "enum-shape", "unsupported member field "+key.Name)
		return false
	}
}

func (p parser) parseValueField(expr ast.Expr, enumType *types.Named, member *Member) bool {
	value, ok := referencedObject(p.pass.TypesInfo, expr).(*types.Const)
	if !ok || !types.Identical(value.Type(), enumType) {
		p.report(expr, "enum-value", "Value must reference a constant of the exact enum type")
		return false
	}

	member.Constant = p.symbol(value)
	member.ValueSource = p.span(expr)
	member.WireValue = constantWireValue(value.Val())
	return true
}

func constantWireValue(value constant.Value) string {
	if value.Kind() == constant.String {
		return constant.StringVal(value)
	}
	return value.ExactString()
}

func (p parser) parseMetadataField(expr ast.Expr, name string, member *Member) bool {
	value := p.pass.TypesInfo.Types[expr].Value
	if value == nil || value.Kind() != constant.String {
		p.report(expr, "enum-metadata", name+" must be a compile-time string constant")
		return false
	}

	if name == "Label" {
		member.Label = constant.StringVal(value)
		member.LabelSource = p.span(expr)
		return true
	}
	member.Description = constant.StringVal(value)
	member.DescriptionSource = p.span(expr)
	return true
}
