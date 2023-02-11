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

type Router interface {
	Routes() []*Route
}

// Route models a single API route.
type Route struct {
	Method  string
	Pattern string
	Params  RouteParams
	Models  Models
}

type RouteParams struct {
	//TODO: Think about URL, querystring and headers.
}

type APIOpts func(*API)

func WithRouter(r Router) APIOpts {
	return func(api *API) {
		//TODO: Walk the routes and add them to the API.
	}
}

// NewAPI creates a new API from the router.
func NewAPI(name string, opts APIOpts) *API {
	return &API{
		Name:       name,
		KnownTypes: defaultKnownTypes,
		Routes:     make(map[Pattern]MethodToRoute),
		// map of model name to schema.
		models: map[string]*openapi3.Schema{},
	}
}

var defaultKnownTypes = map[reflect.Type]*openapi3.Schema{
	reflect.TypeOf(time.Time{}):  openapi3.NewDateTimeSchema(),
	reflect.TypeOf(&time.Time{}): openapi3.NewDateTimeSchema().WithNullable(),
}

type MethodToRoute map[Method]*Route
type Pattern string
type Method string

// API is a model of a REST API's routes, along with their
// request and response types.
type API struct {
	// Name of the API.
	Name string
	// Routes of the API.
	// From patterns, to methods, to route.
	Routes map[Pattern]MethodToRoute
	// StripPkgPaths to strip from the type names in the OpenAPI output to avoid
	// leaking internal implementation details such as internal repo names.
	//
	// This increases the risk of type clashes in the OpenAPI output, i.e. two types
	// in different namespaces that are set to be stripped, and have the same type Name
	// could clash.
	//
	// Example values could be "github.com/a-h/rest".
	StripPkgPaths []string

	// Models are the models that are in use in the API.
	// It's possible to customise the models prior to generation of the OpenAPI specification
	// by editing this value.
	models map[string]*openapi3.Schema

	// KnownTypes are added to the OpenAPI specification output.
	// The default implementation:
	//   Maps time.Time to a string.
	KnownTypes map[reflect.Type]*openapi3.Schema

	// configureSpec is executed after the spec is auto-generated, and can be used to
	// adjust the OpenAPI specification.
	configureSpec func(spec *openapi3.T)

	// handler is a HTTP handler that serves up the OpenAPI specification and Swagger UI.
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
	m.Handle("/", http.FileServer(http.FS(swaggerUI)))
	m.HandleFunc("/swagger-ui/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write(specBytes)
	})
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

func (api *API) Route(method, pattern string) (r *Route) {
	methodToRoute, ok := api.Routes[Pattern(pattern)]
	if !ok {
		methodToRoute = make(MethodToRoute)
		api.Routes[Pattern(pattern)] = methodToRoute
	}
	route, ok := methodToRoute[Method(method)]
	if !ok {
		route = &Route{
			Method:  method,
			Pattern: pattern,
			Models: Models{
				Responses: make(map[int]Model),
			},
		}
		methodToRoute[Method(method)] = route
	}
	return route
}

func (api *API) Get(pattern string) (r *Route) {
	return api.Route(http.MethodGet, pattern)
}
func (api *API) Post(pattern string) (r *Route) {
	return api.Route(http.MethodPost, pattern)
}

//TODO: Add the other verbs.

func (rm *Route) HasResponseModel(status int, response Model) *Route {
	rm.Models.Responses[status] = response
	return rm
}

func (rm *Route) HasRequestModel(request Model) *Route {
	rm.Models.Request = request
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

func ModelFromType(t reflect.Type) Model {
	return Model{
		Type: t,
	}
}

type Model struct {
	Type reflect.Type
}
