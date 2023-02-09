package rest

import (
	"embed"
	"encoding/json"
	"net/http"
	"reflect"
	"sync"
	"time"

	_ "embed"

	"github.com/getkin/kin-openapi/openapi3"
)

// NewAPI creates a new APIModel.
func NewAPI(name string) *API {
	return &API{
		Name:       name,
		KnownTypes: defaultKnownTypes,
	}
}

var defaultKnownTypes = map[reflect.Type]*openapi3.Schema{
	reflect.TypeOf(time.Time{}):  openapi3.NewDateTimeSchema(),
	reflect.TypeOf(&time.Time{}): openapi3.NewDateTimeSchema().WithNullable(),
}

// API is a model of a REST API's routes, along with their
// request and response types.
type API struct {
	// Name of the API.
	Name string
	// Routes of the API.
	Routes []*Route
	// StripPkgPaths to strip from the type names in the OpenAPI output to avoid
	// leaking internal implementation details such as internal repo names.
	//
	// This increases the risk of type clashes in the OpenAPI output, i.e. two types
	// in different namespaces that are set to be stripped, and have the same type Name
	// could clash.
	//
	// Example values could be "github.com/a-h/rest".
	StripPkgPaths []string

	// KnownTypes are added to the OpenAPI specification output.
	// The default implementation:
	//   Maps time.Time to a string.
	KnownTypes map[reflect.Type]*openapi3.Schema

	// configureSpec is executed after the spec is auto-generated, and can be used to
	// adjust the OpenAPI specification.
	configureSpec func(spec *openapi3.T)

	// handler is a HTTP handler that serves up the routes.
	handler    http.Handler
	configured bool
	m          sync.Mutex
}

func (api *API) ConfigureSpec(f func(spec *openapi3.T)) {
	api.configureSpec = f
}

//go:embed swagger-ui/*
var swaggerUI embed.FS

func (api *API) configureHandler() {
	defer api.m.Unlock()
	api.m.Lock()

	// Create JSON specification to serve.
	spec, err := api.Spec()
	if err != nil {
		panic("failed to create specification: " + err.Error())
	}
	specBytes, err := json.MarshalIndent(spec, "", " ")
	if err != nil {
		panic("failed to marshal specification: " + err.Error())
	}

	m := http.NewServeMux()
	for _, r := range api.Routes {
		m.Handle(r.Path, r.Handler)
	}
	m.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write(specBytes)
	})
	m.Handle("/swagger-ui/", http.FileServer(http.FS(swaggerUI)))
	api.handler = m

	api.configured = true
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !api.configured {
		api.configureHandler()
	}
	api.handler.ServeHTTP(w, r)
}

// Spec creates an OpenAPI 3.0 specification document for the API.
func (api *API) Spec() (spec *openapi3.T, err error) {
	spec, err = api.createOpenAPI()
	if err != nil {
		return
	}
	if api.configureSpec != nil {
		api.configureSpec(spec)
	}
	return
}

func (api *API) Handle(path string, handler http.Handler) *Route {
	route := &Route{
		Path:           path,
		Handler:        handler,
		MethodToModels: map[string]Models{},
	}
	api.Routes = append(api.Routes, route)
	return route
}

// Route models a single API route.
type Route struct {
	Path           string
	Handler        http.Handler
	MethodToModels map[string]Models
}

func (rm *Route) WithResponseModel(method string, status int, response Model) *Route {
	models := rm.MethodToModels[method]
	if models.Responses == nil {
		models.Responses = make(map[int]Model)
	}
	models.Responses[status] = response
	rm.MethodToModels[method] = models
	return rm
}

func (rm *Route) WithRequestModel(method string, request Model) *Route {
	models := rm.MethodToModels[method]
	models.Request = request
	rm.MethodToModels[method] = models
	return rm
}

type Models struct {
	Request   Model
	Responses map[int]Model
}

func ModelOf[T any]() Model {
	var t T
	return Model{
		Type: reflect.TypeOf(t),
	}
}

type Model struct {
	Type reflect.Type
}
