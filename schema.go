package rest

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/a-h/rest/enums"
	"github.com/a-h/rest/getcomments/parser"
	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/exp/constraints"
)

func newSpec(name string) *openapi3.T {
	return &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:      name,
			Version:    "0.0.0",
			Extensions: map[string]interface{}{},
		},
		Components: &openapi3.Components{
			Schemas:    make(openapi3.Schemas),
			Extensions: map[string]interface{}{},
		},
		Paths:      &openapi3.Paths{},
		Extensions: map[string]interface{}{},
	}
}

func getSortedKeys[V any](m map[string]V) (op []string) {
	for k := range m {
		op = append(op, k)
	}
	sort.Slice(op, func(i, j int) bool {
		return op[i] < op[j]
	})
	return op
}

func newPrimitiveSchema(paramType PrimitiveType) *openapi3.Schema {
	switch paramType {
	case PrimitiveTypeString:
		return openapi3.NewStringSchema()
	case PrimitiveTypeBool:
		return openapi3.NewBoolSchema()
	case PrimitiveTypeInteger:
		return openapi3.NewIntegerSchema()
	case PrimitiveTypeFloat64:
		return openapi3.NewFloat64Schema()
	case "":
		return openapi3.NewStringSchema()
	default:
		return &openapi3.Schema{
			Type: &openapi3.Types{string(paramType)},
		}
	}
}

func (api *API) createOpenAPI() (spec *openapi3.T, err error) {
	spec = newSpec(api.Name)
	// Add all the routes.
	for pattern, methodToRoute := range api.Routes {
		path := &openapi3.PathItem{}
		for method, route := range methodToRoute {
			op := &openapi3.Operation{}

			// Add the query params.
			for _, k := range getSortedKeys(route.Params.Query) {
				v := route.Params.Query[k]

				ps := newPrimitiveSchema(v.Type).
					WithPattern(v.Regexp)
				queryParam := openapi3.NewQueryParameter(k).
					WithDescription(v.Description).
					WithSchema(ps)
				queryParam.Required = v.Required
				queryParam.AllowEmptyValue = v.AllowEmpty

				// Apply schema customisation.
				if v.ApplyCustomSchema != nil {
					v.ApplyCustomSchema(queryParam)
				}

				op.AddParameter(queryParam)
			}

			// Add the route params.
			for _, k := range getSortedKeys(route.Params.Path) {
				v := route.Params.Path[k]

				ps := newPrimitiveSchema(v.Type).
					WithPattern(v.Regexp)
				pathParam := openapi3.NewPathParameter(k).
					WithDescription(v.Description).
					WithSchema(ps)

				// Apply schema customisation.
				if v.ApplyCustomSchema != nil {
					v.ApplyCustomSchema(pathParam)
				}

				op.AddParameter(pathParam)
			}

			// Handle request types.
			if route.Models.Request.Type != nil {
				name, schema, err := api.RegisterModel(route.Models.Request)
				if err != nil {
					return spec, err
				}
				op.RequestBody = &openapi3.RequestBodyRef{
					Value: openapi3.NewRequestBody().WithContent(map[string]*openapi3.MediaType{
						"application/json": {
							Schema: getSchemaReferenceOrValue(name, schema),
						},
					}),
				}
			}

			// Handle response types.
			for status, model := range route.Models.Responses {
				name, schema, err := api.RegisterModel(model)
				if err != nil {
					return spec, err
				}
				resp := openapi3.NewResponse().
					WithDescription("").
					WithContent(map[string]*openapi3.MediaType{
						"application/json": {
							Schema: getSchemaReferenceOrValue(name, schema),
						},
					})
				op.AddResponse(status, resp)
			}

			// Handle tags.
			op.Tags = append(op.Tags, route.Tags...)

			// Handle OperationID.
			op.OperationID = route.OperationID

			// Handle description.
			op.Description = route.Description

			// Register the method.
			path.SetOperation(string(method), op)
		}

		// Populate the OpenAPI schemas from the models.
		for name, schema := range api.models {
			spec.Components.Schemas[name] = openapi3.NewSchemaRef("", schema)
		}

		spec.Paths.Set(string(pattern), path)
	}

	loader := openapi3.NewLoader()
	if err = loader.ResolveRefsIn(spec, nil); err != nil {
		return spec, fmt.Errorf("failed to resolve, due to external references: %w", err)
	}
	if err = spec.Validate(loader.Context); err != nil {
		return spec, fmt.Errorf("failed validation: %w", err)
	}

	return spec, err
}

func (api *API) getModelName(t reflect.Type) string {
	pkgPath, typeName := t.PkgPath(), t.Name()
	if t.Kind() == reflect.Pointer {
		pkgPath = t.Elem().PkgPath()
		typeName = t.Elem().Name() + "Ptr"
	}
	if t.Kind() == reflect.Map {
		typeName = fmt.Sprintf("map[%s]%s", t.Key().Name(), t.Elem().Name())
	}
	schemaName := api.normalizeTypeName(pkgPath, typeName)
	if typeName == "" {
		schemaName = fmt.Sprintf("AnonymousType%d", len(api.models))
	}
	return schemaName
}

func getSchemaReferenceOrValue(name string, schema *openapi3.Schema) *openapi3.SchemaRef {
	if shouldBeReferenced(schema) {
		return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", name), nil)
	}
	return openapi3.NewSchemaRef("", schema)
}

// ModelOpts defines options that can be set when registering a model.
type ModelOpts func(s *openapi3.Schema)

// WithNullable sets the nullable field to true.
func WithNullable() ModelOpts {
	return func(s *openapi3.Schema) {
		s.Nullable = true
	}
}

// WithDescription sets the description field on the schema.
func WithDescription(desc string) ModelOpts {
	return func(s *openapi3.Schema) {
		s.Description = desc
	}
}

// WithEnumValues sets the property to be an enum value with the specific values.
func WithEnumValues[T ~string | constraints.Integer](values ...T) ModelOpts {
	return func(s *openapi3.Schema) {
		if len(values) == 0 {
			return
		}
		s.Type = &openapi3.Types{openapi3.TypeString}
		if reflect.TypeOf(values[0]).Kind() != reflect.String {
			s.Type = &openapi3.Types{openapi3.TypeInteger}
		}
		for _, v := range values {
			s.Enum = append(s.Enum, v)
		}
	}
}

// WithEnumConstants sets the property to be an enum containing the values of the type found in the package.
func WithEnumConstants[T ~string | constraints.Integer]() ModelOpts {
	return func(s *openapi3.Schema) {
		var t T
		ty := reflect.TypeOf(t)
		s.Type = &openapi3.Types{openapi3.TypeString}
		if ty.Kind() != reflect.String {
			s.Type = &openapi3.Types{openapi3.TypeInteger}
		}
		enum, err := enums.Get(ty)
		if err != nil {
			panic(err)
		}
		s.Enum = enum
	}
}

func isFieldRequired(isPointer, hasOmitEmpty bool) bool {
	return !(isPointer || hasOmitEmpty)
}

func isMarkedAsDeprecated(comment string) bool {
	// A field is only marked as deprecated if a paragraph (line) begins with Deprecated.
	// https://github.com/golang/go/wiki/Deprecated
	for _, line := range strings.Split(comment, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Deprecated:") {
			return true
		}
	}
	return false
}

// RegisterModel allows a model to be registered manually so that additional configuration can be applied.
// The schema returned can be modified as required.
func (api *API) RegisterModel(model Model, opts ...ModelOpts) (name string, schema *openapi3.Schema, err error) {
	// Get the name.
	t := model.Type
	name = api.getModelName(t)

	// If we've already got the schema, return it.
	var ok bool
	if schema, ok = api.models[name]; ok {
		return name, schema, nil
	}

	// It's known, but not in the schemaset yet.
	if knownSchema, ok := api.KnownTypes[t]; ok {
		// Objects, enums, need to be references, so add it into the
		// list.
		if shouldBeReferenced(&knownSchema) {
			api.models[name] = &knownSchema
		}
		return name, &knownSchema, nil
	}

	var elementName string
	var elementSchema *openapi3.Schema
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		elementName, elementSchema, err = api.RegisterModel(modelFromType(t.Elem()))
		if err != nil {
			return name, schema, fmt.Errorf("error getting schema of slice element %v: %w", t.Elem(), err)
		}
		schema = openapi3.NewArraySchema().WithNullable() // Arrays are always nilable in Go.
		schema.Items = getSchemaReferenceOrValue(elementName, elementSchema)
	case reflect.String:
		schema = openapi3.NewStringSchema()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		schema = openapi3.NewIntegerSchema()
	case reflect.Float64, reflect.Float32:
		schema = openapi3.NewFloat64Schema()
	case reflect.Bool:
		schema = openapi3.NewBoolSchema()
	case reflect.Pointer:
		name, schema, err = api.RegisterModel(modelFromType(t.Elem()), WithNullable())
	case reflect.Map:
		// Check that the key is a string.
		if t.Key().Kind() != reflect.String {
			return name, schema, fmt.Errorf("maps must have a string key, but this map is of type %q", t.Key().String())
		}

		// Get the element schema.
		elementName, elementSchema, err = api.RegisterModel(modelFromType(t.Elem()))
		if err != nil {
			return name, schema, fmt.Errorf("error getting schema of map value element %v: %w", t.Elem(), err)
		}
		schema = openapi3.NewObjectSchema().WithNullable()
		schema.AdditionalProperties.Schema = getSchemaReferenceOrValue(elementName, elementSchema)
	case reflect.Struct:
		schema = openapi3.NewObjectSchema()
		if schema.Description, schema.Deprecated, err = api.getTypeComment(t.PkgPath(), t.Name()); err != nil {
			return name, schema, fmt.Errorf("failed to get comments for type %q: %w", name, err)
		}
		schema.Properties = make(openapi3.Schemas)
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			// Get JSON fieldName.
			jsonTags := strings.Split(f.Tag.Get("json"), ",")
			fieldName := jsonTags[0]
			if fieldName == "" {
				fieldName = f.Name
			}
			// If the model doesn't exist.
			_, alreadyExists := api.models[api.getModelName(f.Type)]
			fieldSchemaName, fieldSchema, err := api.RegisterModel(modelFromType(f.Type))
			if err != nil {
				return name, schema, fmt.Errorf("error getting schema for type %q, field %q, failed to get schema for embedded type %q: %w", t, fieldName, f.Type, err)
			}
			if f.Anonymous {
				// It's an anonymous type, no need for a reference to it,
				// since we're copying the fields.
				if !alreadyExists {
					delete(api.models, fieldSchemaName)
				}
				// Add all embedded fields to this type.
				for name, ref := range fieldSchema.Properties {
					schema.Properties[name] = ref
				}
				schema.Required = append(schema.Required, fieldSchema.Required...)
				continue
			}
			ref := getSchemaReferenceOrValue(fieldSchemaName, fieldSchema)
			if ref.Value != nil {
				if ref.Value.Description, ref.Value.Deprecated, err = api.getTypeFieldComment(t.PkgPath(), t.Name(), f.Name); err != nil {
					return name, schema, fmt.Errorf("failed to get comments for field %q in type %q: %w", fieldName, name, err)
				}
			}
			schema.Properties[fieldName] = ref
			isPtr := f.Type.Kind() == reflect.Pointer
			hasOmitEmptySet := slices.Contains(jsonTags, "omitempty")
			if isFieldRequired(isPtr, hasOmitEmptySet) {
				schema.Required = append(schema.Required, fieldName)
			}
		}
	}

	if schema == nil {
		return name, schema, fmt.Errorf("unsupported type: %v/%v", t.PkgPath(), t.Name())
	}

	// Apply global customisation.
	if api.ApplyCustomSchemaToType != nil {
		api.ApplyCustomSchemaToType(t, schema)
	}

	// Customise the model using its ApplyCustomSchema method.
	// This allows any type to customise its schema.
	model.ApplyCustomSchema(schema)

	for _, opt := range opts {
		opt(schema)
	}

	// After all processing, register the type if required.
	if shouldBeReferenced(schema) {
		api.models[name] = schema
		return
	}

	return
}

func (api *API) getCommentsForPackage(pkg string) (pkgComments map[string]string, err error) {
	if pkgComments, loaded := api.comments[pkg]; loaded {
		return pkgComments, nil
	}
	pkgComments, err = parser.Get(pkg)
	if errors.Is(err, parser.ErrGoCommandRequired) {
		err = nil
	}
	if err != nil {
		return
	}
	api.comments[pkg] = pkgComments
	return
}

func (api *API) getTypeComment(pkg string, name string) (comment string, deprecated bool, err error) {
	pkgComments, err := api.getCommentsForPackage(pkg)
	if err != nil {
		return
	}
	comment = pkgComments[pkg+"."+name]
	deprecated = isMarkedAsDeprecated(comment)
	return
}

func (api *API) getTypeFieldComment(pkg string, name string, field string) (comment string, deprecated bool, err error) {
	pkgComments, err := api.getCommentsForPackage(pkg)
	if err != nil {
		return
	}
	comment = pkgComments[pkg+"."+name+"."+field]
	deprecated = isMarkedAsDeprecated(comment)
	return
}

func shouldBeReferenced(schema *openapi3.Schema) bool {
	if schema.Type.Is(openapi3.TypeObject) && schema.AdditionalProperties.Schema == nil {
		return true
	}
	if len(schema.Enum) > 0 {
		return true
	}
	return false
}

var normalizer = strings.NewReplacer("/", "_",
	".", "_",
	"[", "_",
	"]", "_")

func (api *API) normalizeTypeName(pkgPath, name string) string {
	var omitPackage bool
	for _, pkg := range api.StripPkgPaths {
		if strings.HasPrefix(pkgPath, pkg) {
			omitPackage = true
			break
		}
	}
	if omitPackage || pkgPath == "" {
		return normalizer.Replace(name)
	}
	return normalizer.Replace(pkgPath + "/" + name)
}
