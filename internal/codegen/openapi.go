package codegen

import (
	"errors"
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

	request, err := buildOpenAPIRequest(endpoint)
	if err != nil {
		return
	}

	response, err := buildOpenAPIResponses(endpoint)
	if err != nil {
		return
	}

	operation := &v3.Operation{
		Tags:        []string{endpoint.Package.Name},
		Summary:     "TODO",
		Description: "TODO",
		OperationId: endpoint.Name,
		Parameters:  parameters,
		RequestBody: request,
		Responses:   response,
	}
	// create a new path item for this endpoint.
	pathItem := &v3.PathItem{}
	doc.Paths.PathItems[endpoint.Path] = pathItem

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

	for _, v := range endpoint.Request.Fields() {
		queryTag, _ := v.StructTags.Get("query")
		if queryTag == nil {
			continue
		}
		validateTag, _ := v.StructTags.Get("validate")
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

func validateTagHasRequiredAnnotation(validateTag *structtag.Tag) bool {
	return validateTag != nil && (validateTag.Name == "required" || validateTag.HasOption("required"))
}

func buildOpenAPIRequest(endpoint *parser.Endpoint) (result *v3.RequestBody, err error) {
	if endpoint.Request == nil {
		return
	}

	if !lo.ContainsBy(endpoint.Methods, hasHTTPBodyByMethod) {
		return
	}

	result = &v3.RequestBody{}

	parentSchema := &base.Schema{
		Title:       "",
		Description: "",
		Type:        []string{"object"},
		Properties:  make(map[string]*base.SchemaProxy),
	}

	if err = recursivelyBuildSchema(parentSchema, endpoint.Request, "json"); err != nil {
		return
	}

	result.Content = map[string]*v3.MediaType{
		"application/json": {
			Schema: base.CreateSchemaProxy(parentSchema),
		},
	}
	return
}

func recursivelyBuildSchema(parentSchema *base.Schema, baseVar *parser.Var, searchTagName string) (err error) {
	for _, v := range baseVar.Fields() {
		searchTag, _ := v.StructTags.Get(searchTagName)
		if searchTag == nil {
			continue
		}
		validateTag, _ := v.StructTags.Get("validate")
		if validateTagHasRequiredAnnotation(validateTag) {
			parentSchema.Required = append(parentSchema.Required, searchTag.Name)
		}

		var recursive bool
		var schema *base.Schema
		schema, recursive, err = buildOpenApiSchemaWithDefaultChain(v)
		if err != nil {
			return
		}

		if recursive {
			if err = recursivelyBuildSchema(schema, v, searchTagName); err != nil {
				return
			}
		}

		parentSchema.Properties[searchTag.Name] = base.CreateSchemaProxy(schema)
	}
	return
}

type openApiSchemaBuilderFunc func(baseVar *parser.Var) (schema *base.Schema, recursive bool, err error)
type openApiSchemaBuilderChain []openApiSchemaBuilderFunc

func buildOpenApiSchemaChain(baseVar *parser.Var, chain openApiSchemaBuilderChain) (schema *base.Schema, recursive bool, err error) {
	for _, builder := range chain {
		schema, recursive, err = builder(baseVar)
		if err != nil {
			return
		}
		if schema != nil {
			return
		}
	}

	if schema == nil {
		err = errors.Join(errUnsupportedType, errors.New(baseVar.Type().String()))
	}
	return
}

func buildOpenApiSchemaWithDefaultChain(baseVar *parser.Var) (schema *base.Schema, recursive bool, err error) {
	return buildOpenApiSchemaChain(baseVar, openApiSchemaDefaultChain())
}

func openApiSchemaDefaultChain() openApiSchemaBuilderChain {
	return openApiSchemaBuilderChain{
		openApiSchemaFromBasicType,
		openApiSchemaFromStructType,
		openApiSchemaFromGoogleUUIDType,
		openApiSchemaFromMapType,
		// openApiSchemaFromAliasType,
		fallbackType,
	}
}

func fallbackType(baseVar *parser.Var) (schema *base.Schema, recursive bool, err error) {
	schema = &base.Schema{
		Type: []string{"string"},
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
// 		schema, recursive, err = buildOpenApiSchemaWithDefaultChain(&parser.Var{
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

func openApiSchemaFromMapType(baseVar *parser.Var) (schema *base.Schema, recursive bool, err error) {
	_, ok := baseVar.Type().(*types.Map)
	if !ok {
		return
	}

	schema = &base.Schema{
		Type: []string{"object"},
	}
	return
}

func openApiSchemaFromGoogleUUIDType(baseVar *parser.Var) (schema *base.Schema, recursive bool, err error) {
	if baseVar.Type().String() != "github.com/google/uuid.UUID" {
		return
	}

	schema = &base.Schema{
		Type: []string{"uuid"},
	}
	return
}

func openApiSchemaFromStructType(baseVar *parser.Var) (schema *base.Schema, recursive bool, err error) {
	_, ok := baseVar.Type().Underlying().(*types.Struct)
	if !ok {
		return
	}

	schema = &base.Schema{
		Type:       []string{"object"},
		Properties: make(map[string]*base.SchemaProxy),
	}

	recursive = true

	return
}

var errUnsupportedType = errors.New("unsupported type, cannot map to openapi schema")

func openApiSchemaFromBasicType(baseVar *parser.Var) (schema *base.Schema, recursive bool, err error) {
	t, ok := baseVar.Type().(*types.Basic)
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
	default:
		schema = nil
		err = errors.Join(errUnsupportedType, errors.New(t.String()))
		return
	}
	return
}

func hasHTTPBodyByMethod(item string) bool {
	return item == http.MethodPost || item == http.MethodPut || item == http.MethodPatch
}

func buildOpenAPIResponses(endpoint *parser.Endpoint) (result *v3.Responses, err error) {
	if endpoint.Response == nil {
		return
	}

	parentSchema := &base.Schema{
		Title:       "",
		Description: "",
		Type:        []string{"object"},
		Properties:  make(map[string]*base.SchemaProxy),
	}

	if err = recursivelyBuildSchema(parentSchema, endpoint.Response, "json"); err != nil {
		return
	}

	result = &v3.Responses{
		Codes: map[string]*v3.Response{
			"200": {
				Description: "",
				Content: map[string]*v3.MediaType{
					"application/json": {
						Schema: base.CreateSchemaProxy(parentSchema),
					},
				},
			},
		},
	}
	return
}
