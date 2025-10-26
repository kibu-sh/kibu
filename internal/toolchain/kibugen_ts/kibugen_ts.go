package kibugen_ts

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/analysis"
)

var resultType = reflect.TypeOf((*artifact)(nil))

type Artifact interface {
	Contents() string
	OutputPath() string
}

var _ Artifact = (*artifact)(nil)

var missingPackageError = errors.New("missing result of kibugen_ts analyzer")

var Analyzer = &analysis.Analyzer{
	Name:             "kibugen_ts",
	Doc:              "Analyzes go source code for kibu services and generates typescript",
	Requires:         []*analysis.Analyzer{kibumod.Analyzer},
	ResultType:       resultType,
	RunDespiteErrors: true,
	Run:              run,
}

type artifact struct {
	contents string
	pass     *analysis.Pass
}

func (a *artifact) Contents() string {
	return a.contents
}

func (a *artifact) OutputPath() string {
	relPath := modspecv2.RelPathFromPass(a.pass)
	return filepath.Join(relPath, a.pass.Pkg.Name()+".gen.ts")
}

func FromPass(pass *analysis.Pass) (Artifact, bool) {
	result, ok := pass.ResultOf[Analyzer].(*artifact)
	return result, ok
}

func run(pass *analysis.Pass) (any, error) {
	pkg, ok := kibumod.FromPass(pass)
	if !ok {
		return nil, missingPackageError
	}

	if len(pkg.Services) == 0 {
		return nil, nil
	}

	contents := GenerateTypeScript(pkg)

	return &artifact{
		contents: contents,
		pass:     pass,
	}, nil
}

func GenerateTypeScript(pkg *modspecv2.Package) string {
	var sb strings.Builder
	ctx := &genContext{
		pkg:            pkg,
		generatedTypes: make(map[string]bool),
	}

	sb.WriteString("import type { HTTPClient } from './client'\n\n")

	for _, svc := range pkg.Services {
		writeService(&sb, ctx, svc)
	}

	return sb.String()
}

type genContext struct {
	pkg            *modspecv2.Package
	generatedTypes map[string]bool
}

func writeService(sb *strings.Builder, ctx *genContext, svc *modspecv2.Service) {
	for _, op := range svc.Operations {
		if len(op.Params) >= 2 && len(op.Results) >= 1 {
			writeTypeDefinition(sb, ctx, op.Params[1])
			writeTypeDefinition(sb, ctx, op.Results[0])
		}
	}

	serviceFnName := "create" + svc.Name
	sb.WriteString("export function ")
	sb.WriteString(serviceFnName)
	sb.WriteString("(c: HTTPClient) {\n")
	sb.WriteString("  return {\n")

	for _, op := range svc.Operations {
		writeServiceOperation(sb, ctx, svc, op)
	}

	sb.WriteString("  }\n")
	sb.WriteString("}\n")
}

func writeServiceOperation(sb *strings.Builder, ctx *genContext, svc *modspecv2.Service, op *modspecv2.Operation) {
	if len(op.Params) != 2 || len(op.Results) != 2 {
		return
	}

	req := op.Params[1]
	res := op.Results[0]

	methodDecorator, _ := op.Decorators.Find(decorators.HasPrefix("kibu:service:method"))
	httpMethod, _ := methodDecorator.Options.GetOne("method", "POST")
	path, _ := methodDecorator.Options.GetOne("path", fmt.Sprintf("/%s/%s", strings.ToLower(svc.Name), op.Name))

	funcName := op.Name
	if len(funcName) > 0 {
		funcName = strings.ToLower(string(funcName[0])) + funcName[1:]
	}

	sb.WriteString("    async ")
	sb.WriteString(funcName)
	sb.WriteString("(req: ")
	writeTypeName(sb, ctx, req)
	sb.WriteString("): Promise<")
	writeTypeName(sb, ctx, res)
	sb.WriteString("> {\n")
	sb.WriteString("      return c.request({\n")
	sb.WriteString("        method: '")
	sb.WriteString(httpMethod)
	sb.WriteString("',\n")
	sb.WriteString("        pathname: '")
	sb.WriteString(path)
	sb.WriteString("',\n")
	sb.WriteString("        data: req,\n")
	sb.WriteString("      })\n")
	sb.WriteString("    },\n")
}

func writeTypeDefinition(sb *strings.Builder, ctx *genContext, typ modspecv2.Type) {
	typeName := getTypeNameFromExpr(typ.Field.Type)
	if typeName == "" || ctx.generatedTypes[typeName] {
		return
	}
	ctx.generatedTypes[typeName] = true

	typeInfo := ctx.pkg.GoPkg.Scope().Lookup(typeName)
	if typeInfo == nil {
		return
	}

	named, ok := typeInfo.Type().(*types.Named)
	if !ok {
		return
	}

	underlying, ok := named.Underlying().(*types.Struct)
	if !ok {
		return
	}

	for i := 0; i < underlying.NumFields(); i++ {
		field := underlying.Field(i)
		generateNestedTypes(sb, ctx, field.Type())
	}

	sb.WriteString("export type ")
	sb.WriteString(typeName)
	sb.WriteString(" = {\n")

	for i := 0; i < underlying.NumFields(); i++ {
		field := underlying.Field(i)
		tag := underlying.Tag(i)

		jsonName := getJSONFieldName(field.Name(), tag)
		if jsonName == "-" {
			continue
		}

		sb.WriteString("  ")
		sb.WriteString(jsonName)

		isPointer := false
		if _, ok := field.Type().(*types.Pointer); ok {
			isPointer = true
			sb.WriteString("?")
		}

		sb.WriteString(": ")
		writeGoTypeAsTS(sb, ctx, field.Type())

		if isPointer {
			sb.WriteString(" | null")
		}

		sb.WriteString("\n")
	}

	sb.WriteString("}\n\n")
}

func generateNestedTypes(sb *strings.Builder, ctx *genContext, typ types.Type) {
	switch t := typ.(type) {
	case *types.Pointer:
		generateNestedTypes(sb, ctx, t.Elem())
	case *types.Slice:
		generateNestedTypes(sb, ctx, t.Elem())
	case *types.Array:
		generateNestedTypes(sb, ctx, t.Elem())
	case *types.Map:
		generateNestedTypes(sb, ctx, t.Key())
		generateNestedTypes(sb, ctx, t.Elem())
	case *types.Named:
		obj := t.Obj()
		if obj.Pkg() == nil || obj.Pkg() != ctx.pkg.GoPkg {
			return
		}

		typeName := obj.Name()
		if ctx.generatedTypes[typeName] {
			return
		}

		underlying, ok := t.Underlying().(*types.Struct)
		if !ok {
			return
		}

		ctx.generatedTypes[typeName] = true

		for i := 0; i < underlying.NumFields(); i++ {
			field := underlying.Field(i)
			generateNestedTypes(sb, ctx, field.Type())
		}

		sb.WriteString("export type ")
		sb.WriteString(typeName)
		sb.WriteString(" = {\n")

		for i := 0; i < underlying.NumFields(); i++ {
			field := underlying.Field(i)
			tag := underlying.Tag(i)

			jsonName := getJSONFieldName(field.Name(), tag)
			if jsonName == "-" {
				continue
			}

			sb.WriteString("  ")
			sb.WriteString(jsonName)

			isPointer := false
			if _, ok := field.Type().(*types.Pointer); ok {
				isPointer = true
				sb.WriteString("?")
			}

			sb.WriteString(": ")
			writeGoTypeAsTS(sb, ctx, field.Type())

			if isPointer {
				sb.WriteString(" | null")
			}

			sb.WriteString("\n")
		}

		sb.WriteString("}\n\n")
	}
}

func writeTypeName(sb *strings.Builder, ctx *genContext, typ modspecv2.Type) {
	writeTypeExprAsTS(sb, ctx, typ.Field.Type)
}

func writeTypeExprAsTS(sb *strings.Builder, ctx *genContext, expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		sb.WriteString(e.Name)
	case *ast.StarExpr:
		writeTypeExprAsTS(sb, ctx, e.X)
	case *ast.SelectorExpr:
		if x, ok := e.X.(*ast.Ident); ok {
			sb.WriteString(x.Name)
			sb.WriteString(".")
		}
		sb.WriteString(e.Sel.Name)
	case *ast.ArrayType:
		writeTypeExprAsTS(sb, ctx, e.Elt)
		sb.WriteString("[]")
	case *ast.MapType:
		sb.WriteString("Record<")
		writeTypeExprAsTS(sb, ctx, e.Key)
		sb.WriteString(", ")
		writeTypeExprAsTS(sb, ctx, e.Value)
		sb.WriteString(">")
	default:
		sb.WriteString("any")
	}
}

func writeGoTypeAsTS(sb *strings.Builder, ctx *genContext, typ types.Type) {
	switch t := typ.(type) {
	case *types.Basic:
		sb.WriteString(goBasicToTS(t))
	case *types.Pointer:
		writeGoTypeAsTS(sb, ctx, t.Elem())
	case *types.Slice:
		writeGoTypeAsTS(sb, ctx, t.Elem())
		sb.WriteString("[]")
	case *types.Array:
		writeGoTypeAsTS(sb, ctx, t.Elem())
		sb.WriteString("[]")
	case *types.Map:
		sb.WriteString("Record<")
		writeGoTypeAsTS(sb, ctx, t.Key())
		sb.WriteString(", ")
		writeGoTypeAsTS(sb, ctx, t.Elem())
		sb.WriteString(">")
	case *types.Named:
		obj := t.Obj()
		if obj.Pkg() != nil && obj.Pkg() != ctx.pkg.GoPkg {
			sb.WriteString(obj.Pkg().Name())
			sb.WriteString(".")
		}
		sb.WriteString(obj.Name())
	default:
		sb.WriteString("any")
	}
}

func goBasicToTS(basic *types.Basic) string {
	switch basic.Kind() {
	case types.Bool:
		return "boolean"
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64,
		types.Float32, types.Float64:
		return "number"
	case types.String:
		return "string"
	default:
		return "any"
	}
}

func getTypeNameFromExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return getTypeNameFromExpr(e.X)
	case *ast.SelectorExpr:
		return e.Sel.Name
	default:
		return ""
	}
}

func getJSONFieldName(fieldName, tag string) string {
	if tag == "" {
		return fieldName
	}

	tagParts := strings.Split(tag, " ")
	for _, part := range tagParts {
		if strings.HasPrefix(part, "json:") {
			jsonTag := strings.Trim(strings.TrimPrefix(part, "json:"), `"`)
			parts := strings.Split(jsonTag, ",")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
	}

	return fieldName
}
