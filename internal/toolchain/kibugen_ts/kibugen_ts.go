package kibugen_ts

import (
	"embed"
	"fmt"
	"go/ast"
	"go/types"
	"net/url"
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

// normalizePathParams holds the parameters for normalizePathForOutput
type normalizePathParams struct {
	relPath string
	pkgName string
}

// normalizePathForOutput converts a relative path to a normalized output name.
// Falls back to pkgName when the path is empty.
func normalizePathForOutput(params normalizePathParams) string {
	normalized := strings.ReplaceAll(params.relPath, "/", "_")
	if normalized != "" {
		return normalized
	}
	return params.pkgName
}

func (a *artifact) OutputPath() string {
	relPath := modspecv2.RelPathFromPass(a.pass)
	relPath = strings.TrimPrefix(relPath, "/")
	normalizedPath := normalizePathForOutput(normalizePathParams{
		relPath: relPath,
		pkgName: a.pass.Pkg.Name(),
	})
	return fmt.Sprintf("%s.gen.ts", normalizedPath)
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

		if !field.Embedded() {
			result = append(result, promotedField{
				field: field,
				tag:   tag,
			})
			continue
		}

		// This is an embedded field - promote its fields to the parent
		promoted := promoteEmbeddedField(field)
		result = append(result, promoted...)
	}

	return result
}

// promoteEmbeddedField extracts promoted fields from an embedded struct field.
// Returns nil if the embedded type is not a struct.
func promoteEmbeddedField(field *types.Var) []promotedField {
	embeddedType := field.Type()

	// Unwrap pointer if the embedded field is a pointer type
	if ptr, ok := embeddedType.(*types.Pointer); ok {
		embeddedType = ptr.Elem()
	}

	// Get the underlying type (handles type aliases)
	embeddedType = embeddedType.Underlying()

	embeddedStruct, ok := embeddedType.(*types.Struct)
	if !ok {
		return nil
	}

	return getPromotedStructFields(embeddedStruct)
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
		writeFieldDefinition(sb, ctx, pf)
	}

	sb.WriteString("}\n\n")
}

// fieldTypeResult holds the result of resolving a field's TypeScript type
type fieldTypeResult struct {
	tsType     string
	isOptional bool
}

// resolveFieldType resolves a Go type to its TypeScript type string and optionality
func resolveFieldType(ctx *genContext, ty types.Type) fieldTypeResult {
	tsType, isOptional, err := buildWithTypeChain(buildWithTypeChainParams{
		ctx:   ctx,
		ty:    ty,
		chain: defaultTypeChain(),
	})
	if err != nil {
		tsType = "any"
	}
	return fieldTypeResult{tsType: tsType, isOptional: isOptional}
}

// writeFieldDefinition writes a single TypeScript field definition to the builder
func writeFieldDefinition(sb *strings.Builder, ctx *genContext, pf promotedField) {
	jsonName := lookupJSONFieldName(jsonFieldNameParams{
		fieldName: pf.field.Name(),
		tag:       pf.tag,
	})
	if jsonName == "-" {
		return
	}

	sb.WriteString("  ")
	sb.WriteString(jsonName)

	result := resolveFieldType(ctx, pf.field.Type())

	if result.isOptional {
		sb.WriteString("?")
	}

	sb.WriteString(": ")
	sb.WriteString(result.tsType)

	if result.isOptional {
		sb.WriteString(" | null")
	}

	sb.WriteString("\n")
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

	serviceFnName := fmt.Sprintf("create%s", svc.Name)
	ctx.serviceFns = append(ctx.serviceFns, serviceFnName)

	sb.WriteString("function ")
	sb.WriteString(serviceFnName)
	sb.WriteString("(c: HTTPClient) {\n")
	sb.WriteString("  return {\n")

	for _, op := range svc.Operations {
		writeServiceOperation(sb, ctx, serviceOperationParams{
			svc: svc,
			op:  op,
		})
	}

	sb.WriteString("  }\n")
	sb.WriteString("}\n\n")
}

// serviceOperationParams holds the parameters for writing a service operation
type serviceOperationParams struct {
	svc *modspecv2.Service
	op  *modspecv2.Operation
}

func writeServiceOperation(sb *strings.Builder, ctx *genContext, params serviceOperationParams) {
	op := params.op
	svc := params.svc

	if len(op.Params) != 2 || len(op.Results) != 2 {
		return
	}

	req := op.Params[1]
	res := op.Results[0]

	methodDecorator, _ := op.Decorators.Find(decorators.HasPrefix("kibu:service:method"))

	// Skip raw endpoints - they're not exposed in TypeScript clients
	if methodDecorator.Options.Lookup("mode").Or("") == "raw" {
		return
	}

	httpMethod := methodDecorator.Options.Lookup("method").Or("POST")
	path := resolveOperationPath(resolveOperationPathParams{
		pkgName:         ctx.pkg.Name,
		svcName:         svc.Name,
		opName:          op.Name,
		methodDecorator: methodDecorator,
	})

	funcName := toLowerCamelCase(op.Name)

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

// resolveOperationPathParams holds the parameters for resolveOperationPath
type resolveOperationPathParams struct {
	pkgName         string
	svcName         string
	opName          string
	methodDecorator decorators.Line
}

// resolveOperationPath determines the URL path for a service operation
func resolveOperationPath(params resolveOperationPathParams) string {
	lookup := params.methodDecorator.Options.Lookup("path")
	if lookup.Found() {
		return lookup.Or("")
	}
	defaultPath, _ := url.JoinPath("/", params.pkgName, params.svcName, params.opName)
	return defaultPath
}

// toLowerCamelCase converts a PascalCase name to lowerCamelCase
func toLowerCamelCase(name string) string {
	if len(name) == 0 {
		return name
	}
	return fmt.Sprintf("%s%s", strings.ToLower(string(name[0])), name[1:])
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
		writeFieldDefinition(sb, ctx, pf)
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
		generateNestedNamedType(sb, ctx, t)
	}
}

// generateNestedNamedType handles generation of nested named struct types
func generateNestedNamedType(sb *strings.Builder, ctx *genContext, t *types.Named) {
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
		writeFieldDefinition(sb, ctx, pf)
	}

	sb.WriteString("}\n\n")
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

// jsonFieldNameParams holds the parameters for lookupJSONFieldName
type jsonFieldNameParams struct {
	fieldName string
	tag       string
}

// lookupJSONFieldName extracts the JSON field name from a struct tag.
// Returns the fieldName if no json tag is found.
func lookupJSONFieldName(params jsonFieldNameParams) string {
	if params.tag == "" {
		return params.fieldName
	}

	tagParts := strings.Split(params.tag, " ")
	for _, part := range tagParts {
		name := extractJSONName(part)
		if name != "" {
			return name
		}
	}

	return params.fieldName
}

// extractJSONName extracts the JSON field name from a single struct tag part.
// Returns empty string if the part is not a json tag or has no usable name.
func extractJSONName(tagPart string) string {
	if !strings.HasPrefix(tagPart, "json:") {
		return ""
	}
	jsonTag := strings.Trim(strings.TrimPrefix(tagPart, "json:"), `"`)
	parts := strings.Split(jsonTag, ",")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return parts[0]
}

func writeNamespaceExport(sb *strings.Builder, ctx *genContext, pkg *modspecv2.Package) {
	if len(ctx.serviceFns) == 0 {
		return
	}

	namespaceName := toLowerCamelCase(pkg.Name)

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
