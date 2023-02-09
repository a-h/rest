package rest

import (
	"net/http"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
)

// NewAPI creates a new APIModel.
func NewAPI(name string) *API {
	return &API{
		Name: name,
	}
}

// API is a model of a REST API's routes, along with their
// request and response types.
type API struct {
	// Name of the API.
	Name string
	// Routes of the API.
	Routes []*Route
}

// Spec creates an OpenAPI 3.0 specification document for the API.
func (api API) Spec() (spec *openapi3.T, err error) {
	return createOpenAPI(api)
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
