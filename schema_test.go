package rest_test

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	_ "embed"

	"github.com/a-h/rest"
	"github.com/a-h/rest/chiadapter"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
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

type KnownTypes struct {
	Time    time.Time  `json:"time"`
	TimePtr *time.Time `json:"timePtr"`
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type OK struct {
	OK bool `json:"ok"`
}

func TestSchema(t *testing.T) {
	tests := []struct {
		name  string
		setup func(api *rest.API) error
	}{
		{
			name:  "test000.yaml",
			setup: func(api *rest.API) error { return nil },
		},
		{
			name: "test001.yaml",
			setup: func(api *rest.API) error {
				api.Post("/test").
					HasRequestModel(rest.ModelOf[TestRequestType]()).
					HasResponseModel(http.StatusOK, rest.ModelOf[TestResponseType]())
				return nil
			},
		},
		{
			name: "basic-data-types.yaml",
			setup: func(api *rest.API) error {
				api.Post("/test").
					HasRequestModel(rest.ModelOf[AllBasicDataTypes]()).
					HasResponseModel(http.StatusOK, rest.ModelOf[AllBasicDataTypes]())
				return nil
			},
		},
		{
			name: "basic-data-types-pointers.yaml",
			setup: func(api *rest.API) error {
				api.Post("/test").
					HasRequestModel(rest.ModelOf[AllBasicDataTypesPointers]()).
					HasResponseModel(http.StatusOK, rest.ModelOf[AllBasicDataTypesPointers]())
				return nil
			},
		},
		{
			name: "anonymous-type.yaml",
			setup: func(api *rest.API) error {
				api.Post("/test").
					HasRequestModel(rest.ModelOf[struct{ A string }]()).
					HasResponseModel(http.StatusOK, rest.ModelOf[struct{ B string }]())
				return nil
			},
		},
		{
			name: "embedded-structs.yaml",
			setup: func(api *rest.API) error {
				api.Get("/embedded").
					HasResponseModel(http.StatusOK, rest.ModelOf[EmbeddedStructA]())
				api.Post("/test").
					HasRequestModel(rest.ModelOf[WithEmbeddedStructs]()).
					HasResponseModel(http.StatusOK, rest.ModelOf[WithEmbeddedStructs]())
				return nil
			},
		},
		{
			name: "with-name-struct-tags.yaml",
			setup: func(api *rest.API) error {
				api.Post("/test").
					HasRequestModel(rest.ModelOf[WithNameStructTags]()).
					HasResponseModel(http.StatusOK, rest.ModelOf[WithNameStructTags]())
				return nil
			},
		},
		{
			name: "known-types.yaml",
			setup: func(api *rest.API) error {
				api.Route(http.MethodGet, "/test").
					HasResponseModel(http.StatusOK, rest.ModelOf[KnownTypes]())
				return nil
			},
		},
		{
			name: "chi-route-params.yaml",
			setup: func(api *rest.API) (err error) {
				router := chi.NewRouter()
				router.Method(http.MethodGet, `/organisation/{orgId:\d+}/user/{userId}`, testHandler)

				// Automatically get the URL params.
				err = chiadapter.Merge(api, router)
				if err != nil {
					return fmt.Errorf("failed to merge: %w", err)
				}

				// Manually configure the responses.
				api.Route(http.MethodGet, `/organisation/{orgId:\d+}/user/{userId}`).
					HasResponseModel(http.StatusOK, rest.ModelOf[User]())

				return
			},
		},
		{
			name: "all-methods.yaml",
			setup: func(api *rest.API) (err error) {
				api.Get("/get").HasResponseModel(http.StatusOK, rest.ModelOf[OK]())
				api.Head("/head").HasResponseModel(http.StatusOK, rest.ModelOf[OK]())
				api.Post("/post").HasResponseModel(http.StatusOK, rest.ModelOf[OK]())
				api.Put("/put").HasResponseModel(http.StatusOK, rest.ModelOf[OK]())
				api.Patch("/patch").HasResponseModel(http.StatusOK, rest.ModelOf[OK]())
				api.Delete("/delete").HasResponseModel(http.StatusOK, rest.ModelOf[OK]())
				api.Connect("/connect").HasResponseModel(http.StatusOK, rest.ModelOf[OK]())
				api.Options("/options").HasResponseModel(http.StatusOK, rest.ModelOf[OK]())
				api.Trace("/trace").HasResponseModel(http.StatusOK, rest.ModelOf[OK]())
				return
			},
		},
	}

	ignoreUnexportedFieldsIn := []any{
		openapi3.T{},
		openapi3.Schema{},
		openapi3.SchemaRef{},
		openapi3.RequestBodyRef{},
		openapi3.ResponseRef{},
		openapi3.ParameterRef{},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
			api.StripPkgPaths = []string{"github.com/a-h/rest"}
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
		})
	}
}

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, World")
})
