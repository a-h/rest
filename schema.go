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

func createOpenAPI(api *API) (spec *openapi3.T, err error) {
	spec = newSpec(api.Name)
	// Add all the routes.
	for _, r := range api.Routes {
		path := &openapi3.PathItem{}
		methodToOperation := make(map[string]*openapi3.Operation)
		for _, method := range allMethods {
			if models, hasMethod := r.MethodToModels[method]; hasMethod {
				op := &openapi3.Operation{}

				// Handle request types.
				if models.Request.Type != nil {
					ref, err := upsertSchema(spec.Components.Schemas, models.Request.Type)
					if err != nil {
						return spec, err
					}
					op.RequestBody = &openapi3.RequestBodyRef{
						Value: &openapi3.RequestBody{
							Description: "",
							Content: map[string]*openapi3.MediaType{
								"application/json": {
									Schema: ref,
								},
							},
						},
					}
				}

				// Handle response types.
				for status, model := range models.Responses {
					ref, err := upsertSchema(spec.Components.Schemas, model.Type)
					if err != nil {
						return spec, err
					}
					op.AddResponse(status, &openapi3.Response{
						Description: pointerTo(""),
						Content: map[string]*openapi3.MediaType{
							"application/json": {
								Schema: ref,
							},
						},
					})
				}

				// Register the method.
				methodToOperation[method] = op
			}
		}

		// Register the routes.
		for method, operation := range methodToOperation {
			switch method {
			case http.MethodGet:
				path.Get = operation
			case http.MethodHead:
				path.Head = operation
			case http.MethodPost:
				path.Post = operation
			case http.MethodPut:
				path.Put = operation
			case http.MethodPatch:
				path.Patch = operation
			case http.MethodDelete:
				path.Delete = operation
			case http.MethodConnect:
				path.Connect = operation
			case http.MethodOptions:
				path.Options = operation
			case http.MethodTrace:
				path.Trace = operation
			default:
				//TODO: Consider error instead?
				panic("uknown verb")
			}
		}
		spec.Paths[r.Path] = path
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

func upsertSchema(schemas openapi3.Schemas, t reflect.Type) (s *openapi3.SchemaRef, err error) {
	// If it already exists in the collection, return a reference to it.
	if _, hasExisting := schemas[t.Name()]; hasExisting {
		return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", t.Name()), nil), nil
	}

	schema := openapi3.NewSchema()
	schema.Properties = make(openapi3.Schemas)
	schema.Type = "object"

	//TODO: Add the fields using reflection.
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		// Get JSON name.
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if name == "" {
			name = f.Name
		}

		//TODO: Read the struct tags for documentation too.
		//TODO: Set the required fields to be the ones that aren't pointers.

		// Handle basic types.
		switch f.Type.Kind() {
		case reflect.Slice, reflect.Array:
			arraySchema := openapi3.NewArraySchema()
			arraySchema.Items, err = upsertSchema(schemas, f.Type.Elem())
			if err != nil {
				return
			}
			schema.Properties[name] = openapi3.NewSchemaRef("", arraySchema)
			continue
		case reflect.String:
			schema.Properties[name] = openapi3.NewSchemaRef("", openapi3.NewStringSchema())
			continue
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			schema.Properties[name] = openapi3.NewSchemaRef("", openapi3.NewIntegerSchema())
			continue
		case reflect.Float64, reflect.Float32:
			schema.Properties[name] = openapi3.NewSchemaRef("", openapi3.NewFloat64Schema())
			continue
		case reflect.Bool:
			schema.Properties[name] = openapi3.NewSchemaRef("", openapi3.NewBoolSchema())
			continue
		case reflect.Complex64, reflect.Complex128:
			return s, fmt.Errorf("error: complex types are not supported in JSON")
		}

		// Deal with embedded types.
		if f.Anonymous {
			// Add all the fields to this type.
			upsertSchema(schemas, f.Type)
			embedded := schemas[f.Type.Name()]
			for name, ref := range embedded.Value.Properties {
				schema.Properties[name] = ref
			}
			continue
		}

		// If it's not a basic type, it must be complex.
		schema.Properties[name], err = upsertSchema(schemas, f.Type)
		if err != nil {
			return
		}
	}

	ref := openapi3.NewSchemaRef("", schema)
	schemas[t.Name()] = ref

	// Return a reference.
	return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", t.Name()), nil), nil
}
