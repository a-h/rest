package test

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	_ "embed"

	"github.com/a-h/rest"
	"github.com/a-h/rest/chiadapter"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

//go:embed *
var testFiles embed.FS

type TestRequestType struct {
	IntField int
}

// TestResponseType description.
type TestResponseType struct {
	// IntField description.
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
	Byte    byte
	Rune    rune
	String  string
	Bool    bool
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
	Byte    *byte
	Rune    *rune
	String  *string
	Bool    *bool
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

type StringEnum string

const (
	StringEnumA StringEnum = "A"
	StringEnumB StringEnum = "B"
	StringEnumC StringEnum = "B"
)

type IntEnum int64

const (
	IntEnum1 IntEnum = 1
	IntEnum2 IntEnum = 2
	IntEnum3 IntEnum = 3
)

type WithEnums struct {
	S  StringEnum   `json:"s"`
	SS []StringEnum `json:"ss"`
	I  IntEnum      `json:"i"`
	V  string       `json:"v"`
}

type Pence int64

type WithMaps struct {
	Amounts map[string]Pence `json:"amounts"`
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
		{
			name: "enums.yaml",
			setup: func(api *rest.API) (err error) {
				// Register the enums and values.
				api.RegisterModel(rest.ModelOf[StringEnum](), rest.WithEnumValues(StringEnumA, StringEnumB, StringEnumC))
				api.RegisterModel(rest.ModelOf[IntEnum](), rest.WithEnumValues(IntEnum1, IntEnum2, IntEnum3))

				api.Get("/get").HasResponseModel(http.StatusOK, rest.ModelOf[WithEnums]())
				return
			},
		},
		{
			name: "with-maps.yaml",
			setup: func(api *rest.API) (err error) {
				api.Get("/get").HasResponseModel(http.StatusOK, rest.ModelOf[WithMaps]())
				return
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var expected, actual []byte

			var wg sync.WaitGroup
			wg.Add(2)
			errs := make([]error, 2)

			// Validate the test file.
			go func() {
				defer wg.Done()
				// Load test file.
				expectedYAML, err := testFiles.ReadFile(test.name)
				if err != nil {
					errs[0] = fmt.Errorf("could not read file %q: %v", test.name, err)
					return
				}
				expectedSpec, err := openapi3.NewLoader().LoadFromData(expectedYAML)
				if err != nil {
					errs[0] = fmt.Errorf("error in expected YAML: %w", err)
					return
				}
				expected, errs[0] = specToYAML(expectedSpec)
			}()

			go func() {
				defer wg.Done()
				// Create the API.
				api := rest.NewAPI(test.name)
				api.StripPkgPaths = []string{"github.com/a-h/rest"}
				// Configure it.
				test.setup(api)
				// Create the actual spec.
				spec, err := api.Spec()
				if err != nil {
					t.Errorf("failed to generate spec: %v", err)
				}
				actual, errs[1] = specToYAML(spec)
			}()

			wg.Wait()
			var setupFailed bool
			for _, err := range errs {
				if err != nil {
					setupFailed = true
					t.Error(err)
				}
			}
			if setupFailed {
				t.Fatal("test setup failed")
			}

			// Compare the JSON marshalled output to ignore unexported fields and internal state.
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Error(diff)
				t.Error("\n\n" + string(actual))
			}
		})
	}
}

func specToYAML(spec *openapi3.T) (out []byte, err error) {
	// Use JSON, because kin-openapi doesn't customise the YAML output.
	// For example, AdditionalProperties only has a MarshalJSON capability.
	out, err = json.Marshal(spec)
	if err != nil {
		err = fmt.Errorf("could not marshal spec to JSON: %w", err)
		return
	}
	var m map[string]interface{}
	err = json.Unmarshal(out, &m)
	if err != nil {
		return
	}
	return yaml.Marshal(m)
}

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, World")
})
