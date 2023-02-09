package rest

import (
	"embed"
	"io"
	"log"
	"net/http"
	"testing"

	_ "embed"

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
		input *APIModel
	}{
		{
			name:  "test000.yaml",
			input: API("test000"),
		},
		{
			name: "test001.yaml",
			input: API("test001",
				Route("/test").
					Post(fakeHandler{
						requestResponses{
							request:   Request[TestRequestType](),
							responses: Responses(Response[TestResponseType](http.StatusOK)),
						},
					}),
			),
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
		// Create the actual spec.
		actual, err := test.input.Spec()
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

type requestResponses struct {
	request   Model
	responses map[int]Model
}

type fakeHandler struct {
	rr requestResponses
}

func (fh fakeHandler) Models() (request Model, responses map[int]Model) {
	return fh.rr.request, fh.rr.responses
}

func (fh fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}
