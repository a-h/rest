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
					ref, err := getSchema(spec.Components.Schemas, models.Request.Type, getSchemaOpts{})
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
					ref, err := getSchema(spec.Components.Schemas, model.Type, getSchemaOpts{})
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
				return spec, fmt.Errorf("unknown HTTP method: %v", method)
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

type getSchemaOpts struct {
	IsPointer  bool
	IsEmbedded bool
}

func getSchema(schemas openapi3.Schemas, t reflect.Type, opts getSchemaOpts) (s *openapi3.SchemaRef, err error) {
	if _, hasExisting := schemas[t.Name()]; hasExisting {
		return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", t.Name()), nil), nil
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		arraySchema := openapi3.NewArraySchema()
		arraySchema.Nullable = true // Arrays are always nilable in Go.
		arraySchema.Items, err = getSchema(schemas, t.Elem(), getSchemaOpts{})
		if err != nil {
			return
		}
		return openapi3.NewSchemaRef("", arraySchema), nil
	case reflect.String:
		return openapi3.NewSchemaRef("", &openapi3.Schema{
			Type:     openapi3.TypeString,
			Nullable: opts.IsPointer,
		}), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return openapi3.NewSchemaRef("", &openapi3.Schema{
			Type:     openapi3.TypeInteger,
			Nullable: opts.IsPointer,
		}), nil
	case reflect.Float64, reflect.Float32:
		return openapi3.NewSchemaRef("", &openapi3.Schema{
			Type:     openapi3.TypeNumber,
			Nullable: opts.IsPointer,
		}), nil
	case reflect.Bool:
		return openapi3.NewSchemaRef("", &openapi3.Schema{
			Type:     openapi3.TypeBoolean,
			Nullable: opts.IsPointer,
		}), nil
	case reflect.Pointer:
		ref, err := getSchema(schemas, t.Elem(), getSchemaOpts{IsPointer: true})
		if err != nil {
			return nil, fmt.Errorf("error getting schema of pointer to %v: %w", t.Elem(), err)
		}
		return ref, err
	case reflect.Struct:
		schema := openapi3.NewObjectSchema()
		schema.Properties = make(openapi3.Schemas)
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
			if f.Anonymous {
				// Add all the embedded fields to this type.
				// Create an empty
				embedded, err := getSchema(schemas, f.Type, getSchemaOpts{IsEmbedded: true})
				if err != nil {
					return nil, fmt.Errorf("error getting schema of embedded type: %w", err)
				}
				for name, ref := range embedded.Value.Properties {
					schema.Properties[name] = ref
				}
				continue
			}
			schema.Properties[name], err = getSchema(schemas, f.Type, getSchemaOpts{})
		}
		value := openapi3.NewSchemaRef("", schema)
		if opts.IsEmbedded {
			return value, nil
		}
		schemaName := t.Name()
		if schemaName == "" {
			schemaName = fmt.Sprintf("AnonymousType%d", len(schemas))
		}
		schemas[schemaName] = value

		// Return a reference.
		return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", schemaName), nil), nil
	}

	return nil, fmt.Errorf("unsupported type: %v", t.Name())
}
