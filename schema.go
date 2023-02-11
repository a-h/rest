package rest

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

var allMethods = []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

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
		Paths:      openapi3.Paths{},
		Extensions: map[string]interface{}{},
	}
}

func (api *API) createOpenAPI() (spec *openapi3.T, err error) {
	spec = newSpec(api.Name)
	// Add all the routes.
	for pattern, methodToRoute := range api.Routes {
		path := &openapi3.PathItem{}
		methodToOperation := make(map[string]*openapi3.Operation)
		for _, method := range allMethods {
			route, hasMethod := methodToRoute[Method(method)]
			if !hasMethod {
				continue
			}
			op := &openapi3.Operation{}

			// Add the route params.
			for k, v := range route.Params.Path {
				ps := openapi3.NewStringSchema()
				if v.Regexp != "" {
					ps.WithPattern(v.Regexp)
				}
				param := openapi3.NewPathParameter(k).
					WithDescription(v.Description).
					WithSchema(ps)
				op.AddParameter(param)
			}

			// Handle request types.
			if route.Models.Request.Type != nil {
				name, _, err := api.RegisterModel(route.Models.Request)
				if err != nil {
					return spec, err
				}
				op.RequestBody = &openapi3.RequestBodyRef{
					Value: &openapi3.RequestBody{
						Description: "",
						Content: map[string]*openapi3.MediaType{
							"application/json": {
								Schema: getSchemaReference(name),
							},
						},
					},
				}
			}

			// Handle response types.
			for status, model := range route.Models.Responses {
				name, _, err := api.RegisterModel(model)
				if err != nil {
					return spec, err
				}
				op.AddResponse(status, &openapi3.Response{
					Description: pointerTo(""),
					Content: map[string]*openapi3.MediaType{
						"application/json": {
							Schema: getSchemaReference(name),
						},
					},
				})
			}

			// Register the method.
			methodToOperation[method] = op
		}

		// Populate the OpenAPI schemas from the models.
		for name, schema := range api.models {
			spec.Components.Schemas[name] = openapi3.NewSchemaRef("", schema)
		}

		// Register the routes.
		for method, operation := range methodToOperation {
			path.SetOperation(method, operation)
		}
		spec.Paths[string(pattern)] = path
	}

	data, err := spec.MarshalJSON()
	if err != nil {
		return spec, fmt.Errorf("failed to marshal spec to/from JSON: %w", err)
	}
	spec, err = openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		return spec, fmt.Errorf("failed to load spec to/from JSON: %w", err)
	}
	if err = spec.Validate(context.Background()); err != nil {
		return spec, fmt.Errorf("failed validation: %w", err)
	}

	return spec, err
}

func pointerTo[T any](v T) *T {
	return &v
}

func (api *API) getModelName(t reflect.Type) string {
	pkgPath, typeName := t.PkgPath(), t.Name()
	if t.Kind() == reflect.Pointer {
		pkgPath = t.Elem().PkgPath()
		typeName = t.Elem().Name() + "Ptr"
	}
	schemaName := api.normalizeTypeName(pkgPath, typeName)
	if typeName == "" {
		schemaName = fmt.Sprintf("AnonymousType%d", len(api.models))
	}
	return schemaName
}

func getSchemaReference(name string) *openapi3.SchemaRef {
	return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", name), nil)
}

type ModelOpts func(s *openapi3.Schema)

func WithNullable() ModelOpts {
	return func(s *openapi3.Schema) {
		s.Nullable = true
	}
}

func WithDescription(desc string) ModelOpts {
	return func(s *openapi3.Schema) {
		s.Description = desc
	}
}

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
	if schema, ok = api.KnownTypes[t]; ok {
		// Only objects need to be references.
		if schema.Type == openapi3.TypeObject {
			api.models[name] = schema
		}
		return name, schema, nil
	}

	var elementName string
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		elementName, _, err = api.RegisterModel(ModelFromType(t.Elem()))
		if err != nil {
			return name, schema, fmt.Errorf("error getting schema of slice element %v: %w", t.Elem(), err)
		}
		schema = openapi3.NewArraySchema().WithNullable() // Arrays are always nilable in Go.
		schema.Items = getSchemaReference(elementName)
	case reflect.String:
		schema = openapi3.NewStringSchema()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		schema = openapi3.NewIntegerSchema()
	case reflect.Float64, reflect.Float32:
		schema = openapi3.NewFloat64Schema()
	case reflect.Bool:
		schema = openapi3.NewBoolSchema()
	case reflect.Pointer:
		name, schema, err = api.RegisterModel(ModelFromType(t.Elem()), WithNullable())
	case reflect.Struct:
		schema = openapi3.NewObjectSchema()
		schema.Properties = make(openapi3.Schemas)
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			// Get JSON fieldName.
			fieldName := strings.Split(f.Tag.Get("json"), ",")[0]
			if fieldName == "" {
				fieldName = f.Name
			}
			// If the model doesn't exist.
			_, alreadyExists := api.models[api.getModelName(f.Type)]
			fieldSchemaName, fieldSchema, err := api.RegisterModel(ModelFromType(f.Type))
			if err != nil {
				return fieldName, schema, fmt.Errorf("error getting schema for type %q, field %q, failed to get schema for embedded type %q: %w", t, fieldName, f.Type, err)
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
				continue
			}
			if fieldSchema.Type == openapi3.TypeObject {
				schema.Properties[fieldName] = getSchemaReference(fieldSchemaName)
				continue
			}
			schema.Properties[fieldName] = openapi3.NewSchemaRef("", fieldSchema)
		}
		api.models[name] = schema
	}

	if schema == nil {
		return name, schema, fmt.Errorf("unsupported type: %v/%v", t.PkgPath(), t.Name())
	}

	for _, opt := range opts {
		opt(schema)
	}

	return
}

var normalizer = strings.NewReplacer("/", "_", ".", "_")

func (api *API) normalizeTypeName(pkgPath, name string) string {
	var omitPackage bool
	for _, pkg := range api.StripPkgPaths {
		if strings.HasPrefix(pkgPath, pkg) {
			omitPackage = true
			break
		}
	}
	if omitPackage {
		return normalizer.Replace(name)
	}
	return normalizer.Replace(pkgPath + "/" + name)
}
