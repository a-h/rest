package rest

import (
	"net/http"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
)

// API creates a new APIModel.
func API(name string, routes ...*RouteModel) *APIModel {
	return &APIModel{
		Name:   name,
		Routes: routes,
	}
}

// APIModel is a model of a REST API's routes, along with their
// request and response types.
type APIModel struct {
	// Name of the API.
	Name string
	// Routes of the API.
	Routes []*RouteModel
}

// Spec creates an OpenAPI 3.0 specification document for the API.
func (api APIModel) Spec() (spec *openapi3.T, err error) {
	return createOpenAPI(api)
}

// Route creates a new RouteModel.
func Route(path string) *RouteModel {
	return &RouteModel{
		Path:               path,
		MethodToHandlerMap: make(map[string]RouteHandler),
	}
}

// RouteModel models a single API route.
type RouteModel struct {
	Path               string
	MethodToHandlerMap map[string]RouteHandler
}

// Get adds a GET request handler to the route.
func (r *RouteModel) Get(handler Handler) *RouteModel {
	r.MethodToHandlerMap[http.MethodGet] = RouteHandler{
		Handler: handler,
	}
	return r
}

// Post adds a POST request handler to the route.
func (r *RouteModel) Post(handler Handler) *RouteModel {
	r.MethodToHandlerMap[http.MethodPost] = RouteHandler{
		Handler: handler,
	}
	return r
}

// RouteHandler contains information about each route.
type RouteHandler struct {
	Handler Handler
}

type Handler interface {
	http.Handler
	Models() (request Model, responses map[int]Model)
}

func Request[T any]() Model {
	var t T
	return Model{
		Type: reflect.TypeOf(t),
	}
}

type Model struct {
	Type reflect.Type
}

func Responses(models ...ResponseModel) (resp map[int]Model) {
	resp = make(map[int]Model)
	for _, m := range models {
		//TODO: Complain if multiple types are registered.
		resp[m.Status] = m.Model
	}
	return resp
}

func Response[T any](status int) ResponseModel {
	var t T
	return ResponseModel{
		Status: status,
		Model: Model{
			Type: reflect.TypeOf(t),
		},
	}
}

type ResponseModel struct {
	Status int
	Model  Model
}
