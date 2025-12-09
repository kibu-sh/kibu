package kibugen_ts

import (
	"embed"
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/analysis"
)

//go:embed embedded/*
var EmbeddedFiles embed.FS

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
	relPath = strings.TrimPrefix(relPath, "/")
	normalizedPath := strings.ReplaceAll(relPath, "/", "_")
	if normalizedPath == "" {
		normalizedPath = a.pass.Pkg.Name()
	}
	return normalizedPath + ".gen.ts"
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
		serviceFns:     make([]string, 0),
		typeQueue:      make([]*queuedType, 0),
	}

	sb.WriteString("import type { HTTPClient } from './client'\n\n")

	for _, svc := range pkg.Services {
		writeService(&sb, ctx, svc)
	}

	// Process any queued types from cross-package references
	processTypeQueue(&sb, ctx)

	writeNamespaceExport(&sb, ctx, pkg)

	return sb.String()
}

type genContext struct {
	pkg            *modspecv2.Package
	generatedTypes map[string]bool
	serviceFns     []string
	typeQueue      []*queuedType
}

// promotedField represents a struct field that may have been promoted from an embedded struct
type promotedField struct {
	field *types.Var
	tag   string
}

// getPromotedStructFields returns all fields from a struct, with embedded struct fields promoted
// This flattens embedded structs so their fields appear directly in the parent type
func getPromotedStructFields(underlying *types.Struct) []promotedField {
	var result []promotedField

	for i := 0; i < underlying.NumFields(); i++ {
		field := underlying.Field(i)
		tag := underlying.Tag(i)

		if field.Embedded() {
			// This is an embedded field - promote its fields to the parent
			embeddedType := field.Type()

			// Unwrap pointer if the embedded field is a pointer type
			if ptr, ok := embeddedType.(*types.Pointer); ok {
				embeddedType = ptr.Elem()
			}

			// Get the underlying type (handles type aliases)
			embeddedType = embeddedType.Underlying()

			// If it's a struct, recursively get its fields
			if embeddedStruct, ok := embeddedType.(*types.Struct); ok {
				// Recursively promote fields from the embedded struct
				embeddedFields := getPromotedStructFields(embeddedStruct)
				result = append(result, embeddedFields...)
			}
			// Skip adding the embedded field itself - we only want its promoted fields
		} else {
			// Regular field - add it directly
			result = append(result, promotedField{
				field: field,
				tag:   tag,
			})
		}
	}

	return result
}

// processTypeQueue generates TypeScript type definitions for all queued types
func processTypeQueue(sb *strings.Builder, ctx *genContext) {
	// Process types until queue is empty
	// Note: new types may be added to the queue as we process (for nested types)
	for len(ctx.typeQueue) > 0 {
		// Pop the first type from the queue
		queuedType := ctx.typeQueue[0]
		ctx.typeQueue = ctx.typeQueue[1:]

		// Generate the TypeScript type definition
		generateQueuedTypeDefinition(sb, ctx, queuedType)
	}
}

// generateQueuedTypeDefinition generates a TypeScript type definition for a queued type
func generateQueuedTypeDefinition(sb *strings.Builder, ctx *genContext, qt *queuedType) {
	underlying, ok := qt.named.Underlying().(*types.Struct)
	if !ok {
		return
	}

	// Get all fields including promoted fields from embedded structs
	promotedFields := getPromotedStructFields(underlying)

	// First, queue any nested types for all fields
	for _, pf := range promotedFields {
		generateNestedTypes(sb, ctx, pf.field.Type())
	}

	// Generate the type definition
	sb.WriteString("export type ")
	sb.WriteString(qt.name)
	sb.WriteString(" = {\n")

	// Write all promoted fields
	for _, pf := range promotedFields {
		jsonName := getJSONFieldName(pf.field.Name(), pf.tag)
		if jsonName == "-" {
			continue
		}

		sb.WriteString("  ")
		sb.WriteString(jsonName)

		// Check if the field is optional (pointer or NullUUID)
		tsType, isOptional, err := buildWithTypeChain(buildWithTypeChainParams{
			ctx:   ctx,
			ty:    pf.field.Type(),
			chain: defaultTypeChain(),
		})
		if err != nil {
			tsType = "any"
		}

		if isOptional {
			sb.WriteString("?")
		}

		sb.WriteString(": ")
		sb.WriteString(tsType)

		if isOptional {
			sb.WriteString(" | null")
		}

		sb.WriteString("\n")
	}

	sb.WriteString("}\n\n")
}

func writeService(sb *strings.Builder, ctx *genContext, svc *modspecv2.Service) {
	if !svc.Decorators.Some(decorators.HasPrefix("kibu:service")) {
		return
	}

	for _, op := range svc.Operations {
		if len(op.Params) >= 2 && len(op.Results) >= 1 {
			writeTypeDefinition(sb, ctx, op.Params[1])
			writeTypeDefinition(sb, ctx, op.Results[0])
		}
	}

	serviceFnName := "create" + svc.Name
	ctx.serviceFns = append(ctx.serviceFns, serviceFnName)

	sb.WriteString("function ")
	sb.WriteString(serviceFnName)
	sb.WriteString("(c: HTTPClient) {\n")
	sb.WriteString("  return {\n")

	for _, op := range svc.Operations {
		writeServiceOperation(sb, ctx, svc, op)
	}

	sb.WriteString("  }\n")
	sb.WriteString("}\n\n")
}

func writeServiceOperation(sb *strings.Builder, ctx *genContext, svc *modspecv2.Service, op *modspecv2.Operation) {
	if len(op.Params) != 2 || len(op.Results) != 2 {
		return
	}

	req := op.Params[1]
	res := op.Results[0]

	methodDecorator, _ := op.Decorators.Find(decorators.HasPrefix("kibu:service:method"))

	// Skip raw endpoints - they're not exposed in TypeScript clients
	if mode, _ := methodDecorator.Options.GetOne("mode", ""); mode == "raw" {
		return
	}

	httpMethod, _ := methodDecorator.Options.GetOne("method", "POST")
	path, _ := methodDecorator.Options.GetOne("path", fmt.Sprintf("/%s/%s/%s", ctx.pkg.Name, svc.Name, op.Name))

	funcName := op.Name
	if len(funcName) > 0 {
		funcName = strings.ToLower(string(funcName[0])) + funcName[1:]
	}

	// GET and HEAD methods cannot have request bodies per HTTP spec
	cannotHaveBody := httpMethod == "GET" || httpMethod == "HEAD"

	sb.WriteString("    async ")
	sb.WriteString(funcName)
	sb.WriteString("(")
	if !cannotHaveBody {
		sb.WriteString("req: ")
		writeTypeName(sb, ctx, req)
	}
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
	if !cannotHaveBody {
		sb.WriteString("        data: req,\n")
	}
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

	// Get all fields including promoted fields from embedded structs
	promotedFields := getPromotedStructFields(underlying)

	// First, generate nested types for all fields
	for _, pf := range promotedFields {
		generateNestedTypes(sb, ctx, pf.field.Type())
	}

	sb.WriteString("type ")
	sb.WriteString(typeName)
	sb.WriteString(" = {\n")

	// Write all promoted fields
	for _, pf := range promotedFields {
		jsonName := getJSONFieldName(pf.field.Name(), pf.tag)
		if jsonName == "-" {
			continue
		}

		sb.WriteString("  ")
		sb.WriteString(jsonName)

		// Check if the field is optional (pointer or NullUUID)
		tsType, isOptional, err := buildWithTypeChain(buildWithTypeChainParams{
			ctx:   ctx,
			ty:    pf.field.Type(),
			chain: defaultTypeChain(),
		})
		if err != nil {
			tsType = "any"
		}

		if isOptional {
			sb.WriteString("?")
		}

		sb.WriteString(": ")
		sb.WriteString(tsType)

		if isOptional {
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

		// Get all fields including promoted fields from embedded structs
		promotedFields := getPromotedStructFields(underlying)

		// First, generate nested types for all fields
		for _, pf := range promotedFields {
			generateNestedTypes(sb, ctx, pf.field.Type())
		}

		sb.WriteString("export type ")
		sb.WriteString(typeName)
		sb.WriteString(" = {\n")

		// Write all promoted fields
		for _, pf := range promotedFields {
			jsonName := getJSONFieldName(pf.field.Name(), pf.tag)
			if jsonName == "-" {
				continue
			}

			sb.WriteString("  ")
			sb.WriteString(jsonName)

			// Check if the field is optional (pointer or NullUUID)
			tsType, isOptional, err := buildWithTypeChain(buildWithTypeChainParams{
				ctx:   ctx,
				ty:    pf.field.Type(),
				chain: defaultTypeChain(),
			})
			if err != nil {
				tsType = "any"
			}

			if isOptional {
				sb.WriteString("?")
			}

			sb.WriteString(": ")
			sb.WriteString(tsType)

			if isOptional {
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
	tsType, isOptional, err := buildWithTypeChain(buildWithTypeChainParams{
		ctx:   ctx,
		ty:    typ,
		chain: defaultTypeChain(),
	})
	if err != nil {
		sb.WriteString("any")
		return
	}

	sb.WriteString(tsType)

	// Note: isOptional is handled by the caller (writeTypeDefinition and generateNestedTypes)
	// They check for pointer types separately and add the "?" and "| null" suffix
	_ = isOptional
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

func writeNamespaceExport(sb *strings.Builder, ctx *genContext, pkg *modspecv2.Package) {
	if len(ctx.serviceFns) == 0 {
		return
	}

	namespaceName := pkg.Name
	if len(namespaceName) > 0 {
		namespaceName = strings.ToLower(string(namespaceName[0])) + namespaceName[1:]
	}

	sb.WriteString("export const ")
	sb.WriteString(namespaceName)
	sb.WriteString(" = {\n")

	for _, fnName := range ctx.serviceFns {
		sb.WriteString("  ")
		sb.WriteString(fnName)
		sb.WriteString(",\n")
	}

	sb.WriteString("}\n")
}
