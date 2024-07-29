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

	request, err := buildOpenAPIRequest(doc, endpoint)
	if err != nil {
		return
	}

	response, err := buildOpenAPIResponses(doc, endpoint)
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

type DTOType string

const (
	DTOTypeRequest  DTOType = "request"
	DTOTypeResponse DTOType = "response"
)

type schemaBuilderParams struct {
	doc           *v3.Document
	ty            types.Type
	dive          schemaBuilderFunc
	searchTagName string
	dtoType       DTOType
}

func (sb schemaBuilderParams) WithDocument(doc *v3.Document) *schemaBuilderParams {
	sb.doc = doc
	return &sb
}

func (sb schemaBuilderParams) WithType(ty types.Type) *schemaBuilderParams {
	sb.ty = ty
	return &sb
}

func (sb schemaBuilderParams) WithDive(dive schemaBuilderFunc) *schemaBuilderParams {
	sb.dive = dive
	return &sb
}

func (sb schemaBuilderParams) WithSearchTagName(searchTagName string) *schemaBuilderParams {
	sb.searchTagName = searchTagName
	return &sb
}

type schemaBuilderFunc func(
	params *schemaBuilderParams,
) (schemaProxy *base.SchemaProxy, err error)

type schemaBuilderChain []schemaBuilderFunc

type buildWithSchemaChainParams struct {
	doc           *v3.Document
	ty            types.Type
	chain         schemaBuilderChain
	searchTagName string
	dtoType       DTOType
}

func buildWithSchemaChain(
	params buildWithSchemaChainParams,
) (schemaProxy *base.SchemaProxy, err error) {
	diveFunc := createSchemaBuilderDiveFunc(params.chain)
	for _, builder := range params.chain {
		schemaProxy, err = builder(&schemaBuilderParams{
			doc:           params.doc,
			ty:            params.ty,
			dive:          diveFunc,
			searchTagName: params.searchTagName,
			dtoType:       params.dtoType,
		})

		// something bad happened
		if err != nil {
			return
		}

		// we found the schema, no need to continue
		if schemaProxy != nil {
			return
		}
	}

	// don't allow a schema to be null, fallback and add debugging context
	if schemaProxy == nil {
		err = errors.Join(errUnsupportedType, errors.New(params.ty.String()))
	}
	return
}

func createSchemaBuilderDiveFunc(chain schemaBuilderChain) schemaBuilderFunc {
	return func(params *schemaBuilderParams) (*base.SchemaProxy, error) {
		return buildWithSchemaChain(buildWithSchemaChainParams{
			doc:           params.doc,
			ty:            params.ty,
			chain:         chain,
			searchTagName: params.searchTagName,
		})
	}
}

func openApiSchemaDefaultChain() schemaBuilderChain {
	return schemaBuilderChain{
		schemaFromBasicType,
		schemaFromAny,
		// it is important to process more specific types fist
		schemaFromMapType,
		schemaFromTimeDotTime,
		schemaFromGoogleUUIDType,
		schemaFromGoogleNullUUIDType,
		schemaFromDecimalType,
		schemaFromNullDecimalType,
		// process more ambiguous types here
		schemaFromSliceType,
		schemaFromStructType,
		schemaFromPointer,
		// openApiSchemaFromAliasType,
		fallbackType,
	}
}

func schemaFromPointer(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	pointer, ok := params.ty.(*types.Pointer)
	if !ok {
		return
	}

	var schemaRef *base.SchemaProxy
	schemaRef, err = params.dive(params.WithType(pointer.Elem()))
	if err != nil {
		return
	}

	schemaProxy = base.CreateSchemaProxy(&base.Schema{
		AllOf:    []*base.SchemaProxy{schemaRef},
		Nullable: lo.ToPtr(true),
	})

	return
}

func fallbackType(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	return getCachedComponentRef(params.doc, "fallback", func() (*base.SchemaProxy, error) {
		return base.CreateSchemaProxy(&base.Schema{
			Description: fmt.Sprintf("FIXME: fallback for unsupported type %s", params.ty.String()),
			Type:        []string{"string"},
		}), nil
	})
}

func schemaFromMapType(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	if _, ok := params.ty.(*types.Map); !ok {
		return
	}

	return getCachedComponentRef(params.doc, "go.map", func() (*base.SchemaProxy, error) {
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"object"},
		}), nil
	})
}

func schemaFromGoogleUUIDType(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	if params.ty.String() != "github.com/google/uuid.UUID" {
		return
	}
	return getCachedComponentRef(params.doc, params.ty.String(), func() (*base.SchemaProxy, error) {
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"string"},
		}), nil
	})
}

func schemaFromGoogleNullUUIDType(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	if params.ty.String() != "github.com/google/uuid.NullUUID" {
		return
	}
	return getCachedComponentRef(params.doc, params.ty.String(), func() (*base.SchemaProxy, error) {
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"string"},
		}), nil
	})
}

func schemaFromDecimalType(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	if params.ty.String() != "github.com/shopspring/decimal.Decimal" {
		return
	}
	return getCachedComponentRef(params.doc, params.ty.String(), func() (*base.SchemaProxy, error) {
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"string"},
		}), nil
	})
}

func schemaFromNullDecimalType(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	if params.ty.String() != "github.com/shopspring/decimal.NullDecimal" {
		return
	}
	return getCachedComponentRef(params.doc, params.ty.String(), func() (*base.SchemaProxy, error) {
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"string"},
		}), nil
	})
}

func schemaFromTimeDotTime(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	if params.ty.String() != "time.Time" {
		return
	}
	schemaProxy, err = getCachedComponentRef(params.doc, "time.Time", func() (*base.SchemaProxy, error) {
		return base.CreateSchemaProxy(&base.Schema{
			Type:   []string{"string"},
			Format: "date-time",
		}), nil
	})
	return
}

func schemaFromAny(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	if params.ty.String() != "any" {
		return
	}
	schemaProxy, err = getCachedComponentRef(params.doc, "go.any", func() (*base.SchemaProxy, error) {
		return base.CreateSchemaProxy(&base.Schema{
			Type:       []string{"object"},
			Properties: make(map[string]*base.SchemaProxy),
			Nullable:   lo.ToPtr(true),
		}), nil
	})
	return
}

func schemaFromStructType(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	named, ok := params.ty.(*types.Named)
	if !ok {
		return
	}

	_, ok = named.Underlying().(*types.Struct)
	if !ok {
		return
	}

	schemaName := maxPathSuffix(named.String(), 2)
	schemaProxy, err = getCachedComponentRef(params.doc, schemaName, func() (*base.SchemaProxy, error) {
		schemaDefinition := &base.Schema{
			Title:      named.Obj().Name(),
			Type:       []string{"object"},
			Properties: make(map[string]*base.SchemaProxy),
		}

		for _, field := range structFields(params.ty) {
			searchTag, _ := field.Tags.Get(params.searchTagName)
			validateTag, _ := field.Tags.Get("validate")
			openApiTag, _ := field.Tags.Get("openapi")
			fieldName := useStructTagNameOrFieldName(searchTag, field.Var.Name())
			fieldType := field.Var.Type()

			if structTagUsesStandardIgnoreFlag(searchTag) {
				// skip fields that don't have an explicit JSON serialization tag
				continue
			}

			if shouldBeMarkedAsRequired(fieldType, validateTag, params.dtoType) {
				schemaDefinition.Required = append(schemaDefinition.Required, fieldName)
			}

			if openApiTag != nil {
				tagOptions := parseOptions(openApiTag.Options)
				dataType, _ := tagOptions.Get("type")
				dataFormat, _ := tagOptions.Get("format")
				schemaDefinition.Properties[fieldName] = base.CreateSchemaProxy(&base.Schema{
					Type:     []string{dataType.Value},
					Format:   dataFormat.Value,
					Nullable: lo.ToPtr(tagOptions.Has("nullable")),
				})
				continue
			}

			var fieldSchema *base.SchemaProxy
			fieldSchema, err = params.dive(params.WithType(fieldType))
			if err != nil {
				break
			}

			//FIXME
			//if shouldFlatteningEmbeddedStruct(searchTag, field) {
			//	// flatten embedded struct fields
			//	for k, v := range fieldSchema.Schema().Properties {
			//		schemaDefinition.Properties[k] = v
			//	}
			//	continue
			//}

			schemaDefinition.Properties[fieldName] = fieldSchema
		}

		return base.CreateSchemaProxy(schemaDefinition), err
	})

	return
}

func shouldBeMarkedAsRequired(fieldType types.Type, tag *structtag.Tag, dtoType DTOType) bool {
	if dtoType == DTOTypeResponse {
		// as long as the field is not a pointer, it is required on response objects
		return isNotPointerType(fieldType.Underlying())
	}

	// struct field tag has validate:"required"
	if validateTagHasRequiredAnnotation(tag) {
		return true
	}

	// pointers are optional on request objects (unless tagged as required)
	if isPointerType(fieldType.Underlying()) {
		return false
	}

	// google uuids on requests are nullable unless marked as required by annotation
	// this is because JSON unmarshalling of an empty string is not a valid UUID
	if isGoogleUUID(fieldType.Underlying()) || isGoogleNullUUID(fieldType.Underlying()) {
		return false
	}

	return false
}

func isGoogleUUID(fieldType types.Type) bool {
	return fieldType.String() == "github.com/google/uuid.UUID"
}

func isGoogleNullUUID(fieldType types.Type) bool {
	return fieldType.String() == "github.com/google/uuid.NullUUID"
}

func isNotGoogleUUID(fieldType types.Type) bool {
	return !isGoogleUUID(fieldType)
}

func isNotGoogleNullUUID(fieldType types.Type) bool {
	return !isGoogleNullUUID(fieldType)
}

func isNamedType(ty types.Type) bool {
	_, ok := ty.(*types.Named)
	return ok
}

func isStructType(ty types.Type) bool {
	_, ok := ty.Underlying().(*types.Struct)
	return ok
}

func isNotPointerType(fieldType types.Type) bool {
	return !isPointerType(fieldType)
}

func isPointerType(fieldType types.Type) bool {
	_, ok := fieldType.(*types.Pointer)
	return ok
}

func formatGoIDAsOpenApiSchemaRef(s string) string {
	return fmt.Sprintf("#/components/schemas/%s", formatGoIDAsOpenApiSchemaName(s))
}

func formatGoIDAsOpenApiSchemaName(s string) string {
	return strings.ReplaceAll(s, "/", ".")
}

func shouldFlatteningEmbeddedStruct(searchTag *structtag.Tag, field *structField) bool {
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

func schemaFromSliceType(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	sliceType, ok := params.ty.Underlying().(*types.Slice)
	if !ok {
		return
	}

	itemSchema, err := params.dive(params.WithType(
		sliceType.Elem(),
	))
	if err != nil {
		return
	}

	schemaProxy = base.CreateSchemaProxy(&base.Schema{
		Type:     []string{"array"},
		Nullable: lo.ToPtr(true),
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{
			A: itemSchema,
		},
	})

	return
}

var errUnsupportedType = errors.New("unsupported type, cannot map to openapi schema")

func schemaFromBasicType(params *schemaBuilderParams) (schemaProxy *base.SchemaProxy, err error) {
	t, ok := params.ty.(*types.Basic)
	if !ok {
		return
	}

	schema := &base.Schema{}
	schemaProxy = base.CreateSchemaProxy(schema)

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

func buildOpenAPIRequest(
	doc *v3.Document,
	endpoint *parser.Endpoint,
) (result *v3.RequestBody, err error) {
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

	schema, err := buildWithSchemaChain(buildWithSchemaChainParams{
		doc:           doc,
		ty:            endpoint.Request.Type(),
		chain:         openApiSchemaDefaultChain(),
		searchTagName: "json",
		dtoType:       DTOTypeRequest,
	})
	if err != nil {
		return
	}

	result = &v3.RequestBody{
		Content: map[string]*v3.MediaType{
			"application/json": {
				Schema: schema,
			},
		},
	}

	return
}

func buildOpenAPIResponses(
	doc *v3.Document,
	endpoint *parser.Endpoint,
) (result *v3.Responses, err error) {
	if endpoint.Response == nil {
		return
	}

	if endpoint.Response.Name() == "_" {
		return
	}

	schemaProxy, err := buildWithSchemaChain(buildWithSchemaChainParams{
		doc:           doc,
		ty:            endpoint.Response.Type(),
		chain:         openApiSchemaDefaultChain(),
		searchTagName: "json",
		dtoType:       DTOTypeResponse,
	})

	if err != nil {
		return
	}

	result = &v3.Responses{
		Codes: map[string]*v3.Response{
			"200": {
				Description: "",
				Content: map[string]*v3.MediaType{
					"application/json": {
						Schema: schemaProxy,
					},
				},
			},
		},
	}

	return
}

func getCachedComponentRef(doc *v3.Document, key string, build func() (*base.SchemaProxy, error)) (schemaProxy *base.SchemaProxy, err error) {
	schemaName := formatGoIDAsOpenApiSchemaName(key)
	schemaRef := formatGoIDAsOpenApiSchemaRef(key)
	schemaProxy = base.CreateSchemaProxyRef(schemaRef)
	_, cached := doc.Components.Schemas[schemaName]
	if cached {
		return
	}

	doc.Components.Schemas[schemaName], err = build()
	return
}

// maxPathSuffix returns the maximum path specificity started at the end.
//
//	maxPathSuffix("github.com/discernhq/devx/src/backend/systems/foo/bar.Request", 2) => "foo/bar.Request"
func maxPathSuffix(name string, segments int) string {
	// If segments is -1, return the original name
	if segments == -1 {
		return name
	}

	// Split the name by the '/' character
	parts := strings.Split(name, "/")

	// If there are fewer segments than available, return the original name
	if len(parts) < segments {
		return name
	}

	// Take the last 'segments' parts and join them back
	return strings.Join(parts[len(parts)-segments:], "/")
}

type TagOption struct {
	Name  string
	Value string
}

type TagOptions []TagOption

func (t TagOptions) Has(name string) bool {
	for _, option := range t {
		if option.Name == name {
			return true
		}
	}
	return false
}

func (t TagOptions) Get(name string) (to TagOption, ok bool) {
	for _, option := range t {
		if option.Name == name {
			return option, true
		}
	}
	return
}

func (t TagOptions) GetErr(name string) (TagOption, error) {
	if to, ok := t.Get(name); ok {
		return to, nil
	}
	return TagOption{}, fmt.Errorf("tag option %s not found", name)
}

func parseOptions(options []string) TagOptions {
	var tagOptions TagOptions
	for _, option := range options {
		parts := strings.Split(option, ":")
		switch len(parts) {
		case 1:
			tagOptions = append(tagOptions, TagOption{
				Name: parts[0],
			})
		case 2:
			tagOptions = append(tagOptions, TagOption{
				Name:  parts[0],
				Value: parts[1],
			})
		}
	}
	return tagOptions
}
