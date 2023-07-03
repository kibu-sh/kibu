package codegen

import (
	"errors"
	"fmt"
	"github.com/discernhq/devx/internal/parser"
	"github.com/fatih/structtag"
	base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/samber/lo"
	"go/types"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func BuildOpenAPISpec(opts *PipelineOptions) (err error) {
	doc := &v3.Document{
		Version: "3.1.0",
		Info: &base.Info{
			Title:   "Discern API",
			Version: "0.0.1",
			Contact: &base.Contact{
				Name:  "MN3, Inc.",
				Email: "support@discern.com",
			},
		},
		Paths: &v3.Paths{
			PathItems: make(map[string]*v3.PathItem),
		},
		Components: &v3.Components{
			Schemas: make(map[string]*base.SchemaProxy),
		},
	}

	for _, service := range opts.Services {
		endpoints := make([]*parser.Endpoint, 0, len(service.Endpoints))
		for _, endpoint := range service.Endpoints {
			endpoints = append(endpoints, endpoint)
		}

		sort.Slice(endpoints, sortByID(endpoints))

		for _, endpoint := range endpoints {
			if err = buildOpenAPIEndpoint(endpoint, doc); err != nil {
				return
			}
		}
	}

	// create a sample schema that we can use in our document
	// sampleSchema := base.CreateSchemaProxy(&base.Schema{
	// 	Type: []string{"object"},
	// 	Properties: map[string]*base.SchemaProxy{
	// 		"nothing": base.CreateSchemaProxy(&base.Schema{
	// 			Type:    []string{"string"},
	// 			Example: "nothing",
	// 		}),
	// 	},
	// })

	// create a new OpenAPI document.

	// render the document to YAML.
	yamlBytes, err := doc.Render()
	if err != nil {
		return
	}

	err = os.WriteFile(filepath.Join(opts.GenerateParams.OutputDir, "openapi.yaml"), yamlBytes, 0644)
	return
}

func buildOpenAPIEndpoint(endpoint *parser.Endpoint, doc *v3.Document) (err error) {
	parameters, err := buildOpenApiParameters(endpoint)
	if err != nil {
		return
	}

	request, err := buildOpenAPIRequest(endpoint, doc)
	if err != nil {
		return
	}

	response, err := buildOpenAPIResponses(endpoint, doc)
	if err != nil {
		return
	}

	operation := &v3.Operation{
		Tags:        []string{endpoint.Package.Name},
		Summary:     "TODO: this needs to be implemented, where should we pull this from in Go? A the comment above the function?",
		Description: "TODO: this needs to be implemented, where should we pull this from in Go? A the comment above the function?",
		OperationId: endpoint.Name,
		Parameters:  parameters,
		RequestBody: request,
		Responses:   response,
	}

	// create a new path item for this endpoint.
	// reuse existing path item (allows for GET/POST) on the same path
	pathItem, hasPathItem := doc.Paths.PathItems[endpoint.Path]
	if !hasPathItem {
		pathItem = &v3.PathItem{}
		doc.Paths.PathItems[endpoint.Path] = pathItem
	}

	for _, method := range endpoint.Methods {
		switch method {
		case http.MethodGet:
			pathItem.Get = operation
		case http.MethodPost:
			pathItem.Post = operation
		case http.MethodPut:
			pathItem.Put = operation
		case http.MethodPatch:
			pathItem.Patch = operation
		case http.MethodDelete:
			pathItem.Delete = operation
		case http.MethodHead:
			pathItem.Head = operation
		case http.MethodOptions:
			pathItem.Options = operation
		}
	}
	return
}

func buildOpenApiParameters(endpoint *parser.Endpoint) (result []*v3.Parameter, err error) {
	if endpoint.Request == nil {
		return
	}

	for _, field := range structFields(endpoint.Request.Type()) {
		queryTag, _ := field.Tags.Get("query")
		if queryTag == nil {
			return
		}
		validateTag, _ := field.Tags.Get("validate")
		required := validateTagHasRequiredAnnotation(validateTag)
		result = append(result, &v3.Parameter{
			Name:     queryTag.Name,
			In:       "query",
			Required: required,
			Style:    "form", // TODO: think about mapping struct tag options
			Schema: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"string"},
			}),
		})
	}

	return
}

type structField struct {
	Var  *types.Var
	Tags *structtag.Tags
}

func structFields(ty types.Type) (result []*structField) {
	underlying := ty.Underlying()
	if underlying == nil {
		return
	}

	structType, ok := underlying.(*types.Struct)
	if !ok {
		return
	}

	for i := 0; i < structType.NumFields(); i++ {
		tags, _ := structtag.Parse(structType.Tag(i))
		result = append(result, &structField{
			Var:  structType.Field(i),
			Tags: tags,
		})
	}
	return
}

func validateTagHasRequiredAnnotation(validateTag *structtag.Tag) bool {
	return validateTag != nil && (validateTag.Name == "required" || validateTag.HasOption("required"))
}

func parserVarToSchemaRef(v *parser.Var) string {
	return fmt.Sprintf("#/components/schemas/%s", parserVarToSchemaName(v))
}

func parserVarToSchemaName(v *parser.Var) string {
	return strings.Join([]string{
		v.Pkg().Name(), v.TypeName(),
	}, "_")
}

func isStrut(ty types.Type) bool {
	_, ok := ty.Underlying().(*types.Struct)
	return ok
}

func isSlice(ty types.Type) bool {
	_, ok := ty.Underlying().(*types.Slice)
	return ok
}

type schemaBuilderFunc func(ty types.Type, dive schemaBuilderFunc, searchTagName string) (schema *base.Schema, err error)

type schemaBuilderChain []schemaBuilderFunc

func buildWithSchemaChain(ty types.Type, chain schemaBuilderChain, searchTagName string) (schema *base.Schema, err error) {
	diveFunc := createSchemaBuilderDiveFunc(chain)
	for _, builder := range chain {
		schema, err = builder(ty, diveFunc, searchTagName)

		// something bad happened
		if err != nil {
			return
		}

		// we found the schema, no need to continue
		if schema != nil {
			return
		}
	}

	// don't allow a schema to be null, fallback and add debugging context
	if schema == nil {
		err = errors.Join(errUnsupportedType, errors.New(ty.String()))
	}
	return
}

func createSchemaBuilderDiveFunc(chain schemaBuilderChain) schemaBuilderFunc {
	return func(ty types.Type, dive schemaBuilderFunc, searchTagName string) (*base.Schema, error) {
		return buildWithSchemaChain(ty, chain, searchTagName)
	}
}

func buildWithDefaultChain(ty types.Type, searchTagName string) (schema *base.Schema, err error) {
	return buildWithSchemaChain(ty, openApiSchemaDefaultChain(), searchTagName)
}

func openApiSchemaDefaultChain() schemaBuilderChain {
	return schemaBuilderChain{
		schemaFromBasicType,
		schemaFromAny,
		// it is important to process more specific types fist
		schemaFromMapType,
		schemaFromTimeDotTime,
		schemaFromGoogleUUIDType,
		// process more ambiguous types here
		schemaFromSliceType,
		schemaFromStructType,
		schemaFromPointer,
		// openApiSchemaFromAliasType,
		fallbackType,
	}
}

func schemaFromPointer(ty types.Type, dive schemaBuilderFunc, name string) (schema *base.Schema, err error) {
	pointer, ok := ty.(*types.Pointer)
	if !ok {
		return
	}

	schema, err = dive(pointer.Elem(), dive, name)
	if err != nil {
		return
	}

	schema.Nullable = lo.ToPtr(true)
	return
}

func fallbackType(ty types.Type, dive schemaBuilderFunc, _ string) (schema *base.Schema, err error) {
	schema = &base.Schema{
		Description: fmt.Sprintf("FIXME: fallback for unsupported type %s", ty.String()),
		Type:        []string{"string"},
	}
	return
}

func schemaFromMapType(ty types.Type, dive schemaBuilderFunc, _ string) (schema *base.Schema, err error) {
	if _, ok := ty.(*types.Map); !ok {
		return
	}

	schema = &base.Schema{
		Type: []string{"object"},
	}
	return
}

// func openApiSchemaFromAliasType(baseVar *parser.Var) (schema *base.Schema, recursive bool, err error) {
// 	named, ok := baseVar.Type().(*types.TypeName)
// 	if !ok {
// 		return
// 	}
//
// 	if named.IsAlias() {
// 		alias := named.Underlying().(*types.Var)
// 		schema, recursive, err = buildWithDefaultChain(&parser.Var{
// 			Type: alias,
// 		})
// 	}
// 	baseVar.
//
// 	aliasName := alias.Obj().Name()
// 	switch aliasName {
// 	case "Time":
// 		schema = &base.Schema{
// 			Type: []string{"string"},
// 		}
// 		return
// 	}
//
// 	return
// }

func schemaFromGoogleUUIDType(ty types.Type, _ schemaBuilderFunc, _ string) (schema *base.Schema, err error) {
	if ty.String() != "github.com/google/uuid.UUID" {
		return
	}

	schema = &base.Schema{
		Type: []string{"string"},
	}
	return
}

func schemaFromTimeDotTime(ty types.Type, _ schemaBuilderFunc, _ string) (schema *base.Schema, err error) {
	if ty.String() != "time.Time" {
		return
	}

	schema = &base.Schema{
		Type:   []string{"string"},
		Format: "date-time",
	}
	return
}

func schemaFromAny(ty types.Type, _ schemaBuilderFunc, _ string) (schema *base.Schema, err error) {
	if ty.String() != "any" {
		return
	}

	schema = &base.Schema{
		Type:       []string{"object"},
		Properties: make(map[string]*base.SchemaProxy),
		Nullable:   lo.ToPtr(true),
	}
	return
}

func schemaFromStructType(ty types.Type, dive schemaBuilderFunc, searchTagName string) (schema *base.Schema, err error) {
	switch ty.Underlying().(type) {
	case *types.Struct:
		break
	default:
		return
	}

	schema = &base.Schema{
		Type:       []string{"object"},
		Properties: make(map[string]*base.SchemaProxy),
	}

	for _, field := range structFields(ty) {
		searchTag, _ := field.Tags.Get(searchTagName)
		validateTag, _ := field.Tags.Get("validate")
		fieldName := useStructTagNameOrFieldName(searchTag, field.Var.Name())

		// skip fields that don't have an explicit JSON serialization tag
		if structTagUsesStandardIgnoreFlag(searchTag) {
			// TODO: we should log a warning
			continue
		}

		if validateTagHasRequiredAnnotation(validateTag) {
			schema.Required = append(schema.Required, fieldName)
		}

		var fieldSchema *base.Schema
		fieldSchema, err = dive(field.Var.Type(), dive, searchTagName)
		if err != nil {
			return
		}

		if requiresFlatteningEmbeddedStruct(searchTag, field) {
			// flatten embedded struct fields
			for k, v := range fieldSchema.Properties {
				schema.Properties[k] = v
			}
			continue
		}

		schema.Properties[fieldName] = base.CreateSchemaProxy(fieldSchema)
	}

	return
}

func requiresFlatteningEmbeddedStruct(searchTag *structtag.Tag, field *structField) bool {
	return searchTag == nil && field.Var.Embedded()
}

func useStructTagNameOrFieldName(tag *structtag.Tag, name string) string {
	if tag != nil && tag.Name != "" {
		return tag.Name
	}
	return name
}

func structTagUsesStandardIgnoreFlag(searchTag *structtag.Tag) bool {
	return searchTag != nil && (searchTag.HasOption("-") || searchTag.Name == "-")
}

func schemaFromSliceType(ty types.Type, dive schemaBuilderFunc, searchTagName string) (schema *base.Schema, err error) {
	sliceType, ok := ty.Underlying().(*types.Slice)
	if !ok {
		return
	}

	itemSchema, err := dive(sliceType.Elem().Underlying(), dive, searchTagName)
	if err != nil {
		return
	}

	schema = &base.Schema{
		Type: []string{"array"},
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{
			A: base.CreateSchemaProxy(itemSchema),
		},
	}

	return
}

var errUnsupportedType = errors.New("unsupported type, cannot map to openapi schema")

func schemaFromBasicType(ty types.Type, _ schemaBuilderFunc, _ string) (schema *base.Schema, err error) {
	t, ok := ty.(*types.Basic)
	if !ok {
		return
	}

	schema = &base.Schema{}

	switch t.Kind() {
	case types.Bool:
		schema.Type = []string{"boolean"}
	case types.Int, types.Int8, types.Int16, types.Int32:
		schema.Type = []string{"integer"}
	case types.Int64, types.Uint64:
		schema.Type = []string{"long"}
	case types.Float32:
		schema.Type = []string{"float"}
	case types.Float64:
		schema.Type = []string{"double"}
	case types.String:
		schema.Type = []string{"string"}
	case types.Byte:
		schema.Type = []string{"string"}
		schema.Format = "binary"
	default:
		schema.Type = []string{"string"}
		schema.Description = fmt.Sprintf("FIXME: unsupported basic type %s %s", t.String(), t.Name())
		return
	}
	return
}

func hasHTTPBodyByMethod(item string) bool {
	return item == http.MethodPost || item == http.MethodPut || item == http.MethodPatch
}

func buildOpenAPIRequest(endpoint *parser.Endpoint, doc *v3.Document) (result *v3.RequestBody, err error) {
	if endpoint.Request == nil {
		return
	}

	// ignored in go
	if endpoint.Request.Name() == "_" {
		return
	}

	if !lo.ContainsBy(endpoint.Methods, hasHTTPBodyByMethod) {
		return
	}

	schemaRef := parserVarToSchemaRef(endpoint.Request)
	schemaName := parserVarToSchemaName(endpoint.Request)

	result = &v3.RequestBody{
		Content: map[string]*v3.MediaType{
			"application/json": {
				Schema: base.CreateSchemaProxyRef(schemaRef),
			},
		},
	}

	if _, exists := doc.Components.Schemas[schemaName]; exists {
		// TODO: we may need to handle conflicts
		// err = fmt.Errorf("schema name %s already exists", schemaName)
		return
	}

	schema, err := buildWithDefaultChain(endpoint.Request.Type(), "json")
	if err != nil {
		return
	}

	schema.Title = schemaName
	doc.Components.Schemas[schemaName] = base.CreateSchemaProxy(schema)

	return
}

func buildOpenAPIResponses(endpoint *parser.Endpoint, doc *v3.Document) (result *v3.Responses, err error) {
	if endpoint.Response == nil {
		return
	}

	// FIXME: ignored in go
	if endpoint.Response.Name() == "_" {
		return
	}

	schemaRef := parserVarToSchemaRef(endpoint.Response)
	schemaName := parserVarToSchemaName(endpoint.Response)
	result = &v3.Responses{
		Codes: map[string]*v3.Response{
			"200": {
				Description: "",
				Content: map[string]*v3.MediaType{
					"application/json": {
						Schema: base.CreateSchemaProxyRef(schemaRef),
					},
				},
			},
		},
	}

	if _, exists := doc.Components.Schemas[schemaName]; exists {
		// TODO: we may need to handle conflicts
		// err = fmt.Errorf("schema name %s already exists", schemaName)
		return
	}

	schema, err := buildWithDefaultChain(endpoint.Response.Type(), "json")
	if err != nil {
		return
	}

	schema.Title = schemaName
	recursivelyMarkResponseSchemaFieldsAsRequired(schema)
	doc.Components.Schemas[schemaName] = base.CreateSchemaProxy(schema)

	return
}

func recursivelyMarkResponseSchemaFieldsAsRequired(schema *base.Schema) {
	// dive into typed arrays
	if schema.Items != nil {
		// it's always A for now
		recursivelyMarkResponseSchemaFieldsAsRequired(schema.Items.A.Schema())
		return
	}

	// dive into typed objects
	for name, field := range schema.Properties {
		fieldSchema := field.Schema()
		recursivelyMarkResponseSchemaFieldsAsRequired(fieldSchema)
		// non-nullable schemas in a response are required
		// this means the server will at least send the zero value of this type
		if !schemaIsNullable(fieldSchema) {
			schema.Required = append(schema.Required, name)
		}
	}
}

func schemaIsNullable(schema *base.Schema) bool {
	return lo.FromPtr(schema.Nullable)
}
