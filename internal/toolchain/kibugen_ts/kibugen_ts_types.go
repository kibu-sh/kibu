package kibugen_ts

import (
	"fmt"
	"go/types"
	"strings"
)

// typeBuilderParams holds the context and information needed to build TypeScript types
type typeBuilderParams struct {
	ctx  *genContext
	ty   types.Type
	dive typeBuilderFunc
}

// typeBuilderFunc is a function that attempts to convert a Go type to a TypeScript type string
// Returns empty string if it cannot handle the type (passes to next handler)
type typeBuilderFunc func(params *typeBuilderParams) (tsType string, isOptional bool, err error)

// typeBuilderChain is a sequence of type builder functions that are tried in order
type typeBuilderChain []typeBuilderFunc

// buildWithTypeChainParams contains parameters for building types with a chain
type buildWithTypeChainParams struct {
	ctx   *genContext
	ty    types.Type
	chain typeBuilderChain
}

// buildWithTypeChain executes a chain of type builders until one succeeds
func buildWithTypeChain(params buildWithTypeChainParams) (tsType string, isOptional bool, err error) {
	diveFunc := createTypeBuilderDiveFunc(params.chain)
	for _, builder := range params.chain {
		tsType, isOptional, err = builder(&typeBuilderParams{
			ctx:  params.ctx,
			ty:   params.ty,
			dive: diveFunc,
		})

		// something bad happened
		if err != nil {
			return
		}

		// we found the type, no need to continue
		if tsType != "" {
			return
		}
	}

	// fallback - should not reach here if fallback handler is in chain
	tsType = "any"
	return
}

// createTypeBuilderDiveFunc creates a dive function for recursive type resolution
func createTypeBuilderDiveFunc(chain typeBuilderChain) typeBuilderFunc {
	return func(params *typeBuilderParams) (string, bool, error) {
		return buildWithTypeChain(buildWithTypeChainParams{
			ctx:   params.ctx,
			ty:    params.ty,
			chain: chain,
		})
	}
}

// defaultTypeChain returns the default chain of type builders in priority order
func defaultTypeChain() typeBuilderChain {
	return typeBuilderChain{
		typeFromBasicType,
		typeFromUUID,
		typeFromNullUUID,
		typeFromTime,
		typeFromPointer,
		typeFromSlice,
		typeFromArray,
		typeFromMap,
		typeFromNamedStruct,
		typeFallback,
	}
}

// typeFromBasicType handles Go basic types (bool, int, string, etc.)
func typeFromBasicType(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	basic, ok := params.ty.(*types.Basic)
	if !ok {
		return
	}

	switch basic.Kind() {
	case types.Bool:
		tsType = "boolean"
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64,
		types.Float32, types.Float64:
		tsType = "number"
	case types.String:
		tsType = "string"
	default:
		tsType = "any"
	}
	return
}

// typeFromPointer handles pointer types by unwrapping and marking as optional
func typeFromPointer(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	pointer, ok := params.ty.(*types.Pointer)
	if !ok {
		return
	}

	// Unwrap the pointer and process the underlying type
	tsType, _, err = params.dive(&typeBuilderParams{
		ctx:  params.ctx,
		ty:   pointer.Elem(),
		dive: params.dive,
	})
	isOptional = true
	return
}

// typeFromSlice handles slice types
func typeFromSlice(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	slice, ok := params.ty.(*types.Slice)
	if !ok {
		return
	}

	var elemType string
	elemType, _, err = params.dive(&typeBuilderParams{
		ctx:  params.ctx,
		ty:   slice.Elem(),
		dive: params.dive,
	})
	if err != nil {
		return
	}

	tsType = elemType + "[]"
	return
}

// typeFromArray handles array types
func typeFromArray(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	array, ok := params.ty.(*types.Array)
	if !ok {
		return
	}

	var elemType string
	elemType, _, err = params.dive(&typeBuilderParams{
		ctx:  params.ctx,
		ty:   array.Elem(),
		dive: params.dive,
	})
	if err != nil {
		return
	}

	tsType = elemType + "[]"
	return
}

// typeFromMap handles map types
func typeFromMap(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	mapType, ok := params.ty.(*types.Map)
	if !ok {
		return
	}

	var keyType, valueType string
	keyType, _, err = params.dive(&typeBuilderParams{
		ctx:  params.ctx,
		ty:   mapType.Key(),
		dive: params.dive,
	})
	if err != nil {
		return
	}

	valueType, _, err = params.dive(&typeBuilderParams{
		ctx:  params.ctx,
		ty:   mapType.Elem(),
		dive: params.dive,
	})
	if err != nil {
		return
	}

	tsType = fmt.Sprintf("Record<%s, %s>", keyType, valueType)
	return
}

// typeFromUUID handles github.com/google/uuid.UUID
func typeFromUUID(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	named, ok := params.ty.(*types.Named)
	if !ok {
		return
	}

	obj := named.Obj()
	if obj.Pkg() == nil || obj.Pkg().Path() != "github.com/google/uuid" || obj.Name() != "UUID" {
		return
	}

	tsType = "string"
	return
}

// typeFromNullUUID handles github.com/google/uuid.NullUUID
func typeFromNullUUID(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	named, ok := params.ty.(*types.Named)
	if !ok {
		return
	}

	obj := named.Obj()
	if obj.Pkg() == nil || obj.Pkg().Path() != "github.com/google/uuid" || obj.Name() != "NullUUID" {
		return
	}

	tsType = "string"
	isOptional = true
	return
}

// typeFromTime handles time.Time
func typeFromTime(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	named, ok := params.ty.(*types.Named)
	if !ok {
		return
	}

	obj := named.Obj()
	if obj.Pkg() == nil || obj.Pkg().Path() != "time" || obj.Name() != "Time" {
		return
	}

	tsType = "string"
	return
}

// typeFromNamedStruct handles named struct types and generates TypeScript type definitions
func typeFromNamedStruct(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	named, ok := params.ty.(*types.Named)
	if !ok {
		return
	}

	obj := named.Obj()
	if obj.Pkg() == nil {
		return
	}

	// Check if it's a struct type
	_, ok = named.Underlying().(*types.Struct)
	if !ok {
		return
	}

	// Generate fully qualified type name
	fullyQualifiedName := makeFullyQualifiedTypeName(named)
	tsType = fullyQualifiedName

	// Check if we've already generated or queued this type
	if params.ctx.generatedTypes[fullyQualifiedName] {
		return
	}

	// Mark as generated to prevent duplicates
	params.ctx.generatedTypes[fullyQualifiedName] = true

	// Queue the type for generation
	params.ctx.typeQueue = append(params.ctx.typeQueue, &queuedType{
		name:  fullyQualifiedName,
		named: named,
	})

	return
}

// typeFallback is the last resort handler that returns "any"
func typeFallback(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	tsType = "any"
	return
}

// makeFullyQualifiedTypeName creates a TypeScript-safe fully qualified type name
// Example: backend/typesv1.User -> backend_typesv1_User
func makeFullyQualifiedTypeName(named *types.Named) string {
	obj := named.Obj()
	if obj.Pkg() == nil {
		return obj.Name()
	}

	// Get the package path and replace slashes with underscores
	pkgPath := obj.Pkg().Path()
	typeName := obj.Name()

	// Find the last segment after the module path
	// For example: github.com/example/module/backend/typesv1 -> backend/typesv1
	parts := strings.Split(pkgPath, "/")
	var relevantParts []string

	// Try to find where the actual project structure starts
	// This is a heuristic - we look for common patterns
	foundModule := false
	for i, part := range parts {
		// Skip until we find something that looks like it's past the module root
		// Common patterns: "internal", "pkg", "cmd", or directories like "backend", "frontend", etc.
		if !foundModule && (part == "internal" || part == "pkg" || part == "cmd" || i >= len(parts)-2) {
			foundModule = true
		}
		if foundModule {
			relevantParts = append(relevantParts, part)
		}
	}

	// If we didn't find a good split point, just use the last 2 segments
	if len(relevantParts) == 0 && len(parts) >= 2 {
		relevantParts = parts[len(parts)-2:]
	} else if len(relevantParts) == 0 {
		// Last resort: use the package name
		return obj.Pkg().Name() + "_" + typeName
	}

	// Join with underscores to create a TS-safe name
	prefix := strings.Join(relevantParts, "_")
	return prefix + "_" + typeName
}

// queuedType represents a type that needs to be generated
type queuedType struct {
	name  string
	named *types.Named
}
