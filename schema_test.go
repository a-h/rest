package rest_test

import (
	"embed"
	"io"
	"log"
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
			t.Fatalf("failed to generate spec: %v", err)
		}

		// Compare.
		if diff := cmp.Diff(expected, actual, cmpopts.IgnoreUnexported(ignoreUnexportedFieldsIn...)); diff != "" {
			t.Error(diff)
			d, err := yaml.Marshal(actual)
			if err != nil {
				log.Fatalf("failed to marshal actual value: %v", err)
			}
			t.Error(string(d))
		}
	}
}

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, World")
})
