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

func createOpenAPI(api APIModel) (spec *openapi3.T, err error) {
	spec = &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   api.Name,
			Version: "0.0.0",
		},
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
		Paths: openapi3.Paths{},
	}

	// Add all the routes.
	for _, r := range api.Routes {
		path := &openapi3.PathItem{}
		methodToOperation := make(map[string]*openapi3.Operation)
		for _, method := range allMethods {
			if handler, hasMethod := r.MethodToHandlerMap[method]; hasMethod {
				op := &openapi3.Operation{}

				// Get the models.
				reqModel, resModels := handler.Handler.Models()

				// Handle request types.
				if reqModel.Type != nil {
					ref := upsertSchema(spec.Components.Schemas, reqModel.Type)
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
				for status, model := range resModels {
					ref := upsertSchema(spec.Components.Schemas, model.Type)
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

func upsertSchema(schemas openapi3.Schemas, t reflect.Type) *openapi3.SchemaRef {
	// If it already exists in the collection, return a reference to it.
	if _, hasExisting := schemas[t.Name()]; hasExisting {
		return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", t.Name()), nil)
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
		if isString(f) {
			schema.Properties[name] = openapi3.NewSchemaRef("", openapi3.NewStringSchema())
			continue
		}
		if isInteger(f) {
			schema.Properties[name] = openapi3.NewSchemaRef("", openapi3.NewIntegerSchema())
			continue
		}
		if isArray(f) {
			arraySchema := openapi3.NewArraySchema()
			arraySchema.Items = upsertSchema(schemas, f.Type.Elem())
			schema.Properties[name] = openapi3.NewSchemaRef("", arraySchema)
			continue
		}
		if isBoolean(f) {
			schema.Properties[name] = openapi3.NewSchemaRef("", openapi3.NewBoolSchema())
			continue
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
		schema.Properties[name] = upsertSchema(schemas, f.Type)
	}

	ref := openapi3.NewSchemaRef("", schema)
	schemas[t.Name()] = ref

	// Return a reference.
	return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", t.Name()), nil)
}

func isArray(f reflect.StructField) bool {
	return f.Type.Kind() == reflect.Array || f.Type.Kind() == reflect.Slice
}

func isBoolean(f reflect.StructField) bool {
	return f.Type.Name() == "bool"
}

func isInteger(f reflect.StructField) bool {
	return f.Type.Name() == "int" ||
		f.Type.Name() == "int64" ||
		f.Type.Name() == "int32" ||
		f.Type.Name() == "uint" ||
		f.Type.Name() == "uint32" ||
		f.Type.Name() == "uint64" ||
		f.Type.Name() == "uint8"
}

func isString(f reflect.StructField) bool {
	return f.Type.Name() == "string"
}
