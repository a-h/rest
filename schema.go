package rest

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

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
