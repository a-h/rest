package rest_test

import (
	"embed"
	"io"
	"net/http"
	"testing"

	_ "embed"

	"github.com/a-h/rest"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gopkg.in/yaml.v2"
)

//go:embed tests/*
var testFiles embed.FS

type TestRequestType struct {
	IntField int
}

type TestResponseType struct {
	IntField int
}

type AllBasicDataTypes struct {
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint    uint
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Uintptr uintptr
	Float32 float32
	Float64 float64
	// Complex types are not supported by the Go JSON serializer.
	//Complex64  complex64
	//Complex128 complex128
	Byte   byte
	Rune   rune
	String string
	Bool   bool
}

type AllBasicDataTypesPointers struct {
	Int     *int
	Int8    *int8
	Int16   *int16
	Int32   *int32
	Int64   *int64
	Uint    *uint
	Uint8   *uint8
	Uint16  *uint16
	Uint32  *uint32
	Uint64  *uint64
	Uintptr *uintptr
	Float32 *float32
	Float64 *float64
	// Complex types are not supported by the Go JSON serializer.
	//Complex64  *complex64
	//Complex128 *complex128
	Byte   *byte
	Rune   *rune
	String *string
	Bool   *bool
}

type EmbeddedStructA struct {
	A string
}
type EmbeddedStructB struct {
	B string
}

type WithEmbeddedStructs struct {
	EmbeddedStructA
	EmbeddedStructB
	C string
}

type WithNameStructTags struct {
	Name string `json:"name"`
}

func TestSchema(t *testing.T) {
	tests := []struct {
		name  string
		setup func(api *rest.API)
	}{
		{
			name:  "test000.yaml",
			setup: func(api *rest.API) {},
		},
		{
			name: "test001.yaml",
			setup: func(api *rest.API) {
				api.Handle("/test", testHandler).
					WithRequestModel(http.MethodPost, rest.ModelOf[TestRequestType]()).
					WithResponseModel(http.MethodPost, http.StatusOK, rest.ModelOf[TestResponseType]())
			},
		},
		{
			name: "basic-data-types.yaml",
			setup: func(api *rest.API) {
				api.Handle("/test", testHandler).
					WithRequestModel(http.MethodPost, rest.ModelOf[AllBasicDataTypes]()).
					WithResponseModel(http.MethodPost, http.StatusOK, rest.ModelOf[AllBasicDataTypes]())
			},
		},
		{
			name: "basic-data-types-pointers.yaml",
			setup: func(api *rest.API) {
				api.Handle("/test", testHandler).
					WithRequestModel(http.MethodPost, rest.ModelOf[AllBasicDataTypesPointers]()).
					WithResponseModel(http.MethodPost, http.StatusOK, rest.ModelOf[AllBasicDataTypesPointers]())
			},
		},
		{
			name: "anonymous-type.yaml",
			setup: func(api *rest.API) {
				api.Handle("/test", testHandler).
					WithRequestModel(http.MethodPost, rest.ModelOf[struct{ A string }]()).
					WithResponseModel(http.MethodPost, http.StatusOK, rest.ModelOf[struct{ B string }]())
			},
		},
		{
			name: "embedded-structs.yaml",
			setup: func(api *rest.API) {
				api.Handle("/embedded", testHandler).
					WithResponseModel(http.MethodGet, http.StatusOK, rest.ModelOf[EmbeddedStructA]())
				api.Handle("/test", testHandler).
					WithRequestModel(http.MethodPost, rest.ModelOf[WithEmbeddedStructs]()).
					WithResponseModel(http.MethodPost, http.StatusOK, rest.ModelOf[WithEmbeddedStructs]())
			},
		},
		{
			name: "with-name-struct-tags.yaml",
			setup: func(api *rest.API) {
				api.Handle("/test", testHandler).
					WithRequestModel(http.MethodPost, rest.ModelOf[WithNameStructTags]()).
					WithResponseModel(http.MethodPost, http.StatusOK, rest.ModelOf[WithNameStructTags]())
			},
		},
	}

	ignoreUnexportedFieldsIn := []any{
		openapi3.T{},
		openapi3.Schema{},
		openapi3.SchemaRef{},
		openapi3.RequestBodyRef{},
		openapi3.ResponseRef{},
	}

	for _, test := range tests {
		// Load test file.
		expectedYAML, err := testFiles.ReadFile("tests/" + test.name)
		if err != nil {
			t.Fatalf("could not read file %q: %v", test.name, err)
		}
		expected, err := openapi3.NewLoader().LoadFromData(expectedYAML)
		if err != nil {
			t.Errorf("could not load expected YAML: %v", err)
		}

		// Create the API.
		api := rest.NewAPI(test.name)
		// Configure it.
		test.setup(api)
		// Create the actual spec.
		actual, err := api.Spec()
		if err != nil {
			t.Errorf("failed to generate spec: %v", err)
		}

		// Compare.
		if diff := cmp.Diff(expected, actual, cmpopts.IgnoreUnexported(ignoreUnexportedFieldsIn...)); diff != "" {
			t.Error(diff)
			d, err := yaml.Marshal(actual)
			if err != nil {
				t.Fatalf("failed to marshal actual value: %v", err)
			}
			t.Error("\n\n" + string(d))
		}
	}
}

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, World")
})
