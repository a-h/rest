package rest

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func API(name string, routes ...*RouteDef) *APIDef {
	return &APIDef{
		Name:   name,
		Routes: routes,
	}
}

type APIDef struct {
	Name   string
	Routes []*RouteDef
}

var allMethods = []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

func (api APIDef) Spec() (spec *openapi3.T, err error) {
	spec = &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   api.Name,
			Version: "0.0.0",
		},
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
		Paths: openapi3.Paths{},
	}

	// Add all the routes.
	for _, r := range api.Routes {
		path := &openapi3.PathItem{}
		methodToOperation := make(map[string]*openapi3.Operation)
		for _, method := range allMethods {
			if handler, hasMethod := r.MethodToHandlerMap[method]; hasMethod {
				op := &openapi3.Operation{}

				// Get the models.
				reqModel, resModels := handler.Handler.Models()

				// Handle request types.
				if reqModel.Type != nil {
					ref := upsertSchema(spec.Components.Schemas, reqModel.Type)
					op.RequestBody = &openapi3.RequestBodyRef{
						Value: &openapi3.RequestBody{
							Description: "",
							Content: map[string]*openapi3.MediaType{
								"application/json": {
									Schema: ref,
								},
							},
						},
					}
				}

				// Handle response types.
				for status, model := range resModels {
					ref := upsertSchema(spec.Components.Schemas, model.Type)
					op.AddResponse(status, &openapi3.Response{
						Description: pointerTo(""),
						Content: map[string]*openapi3.MediaType{
							"application/json": {
								Schema: ref,
							},
						},
					})
				}

				// Register the method.
				methodToOperation[method] = op
			}
		}

		// Register the routes.
		for method, operation := range methodToOperation {
			switch method {
			case http.MethodGet:
				path.Get = operation
			case http.MethodHead:
				path.Head = operation
			case http.MethodPost:
				path.Post = operation
			case http.MethodPut:
				path.Put = operation
			case http.MethodPatch:
				path.Patch = operation
			case http.MethodDelete:
				path.Delete = operation
			case http.MethodConnect:
				path.Connect = operation
			case http.MethodOptions:
				path.Options = operation
			case http.MethodTrace:
				path.Trace = operation
			default:
				//TODO: Consider error instead?
				panic("uknown verb")
			}
		}
		spec.Paths[r.Path] = path
	}

	data, err := spec.MarshalJSON()
	if err != nil {
		return spec, fmt.Errorf("failed to marshal spec to/from JSON: %w", err)
	}
	spec, err = openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		return spec, fmt.Errorf("failed to load spec to/from JSON: %w", err)
	}
	if err = spec.Validate(context.Background()); err != nil {
		return spec, fmt.Errorf("failed validation: %w", err)
	}
	return spec, err
}

func pointerTo[T any](v T) *T {
	return &v
}

func Route(path string) *RouteDef {
	return &RouteDef{
		Path:               path,
		MethodToHandlerMap: make(map[string]RouteHandler),
	}
}

type RouteDef struct {
	Path               string
	MethodToHandlerMap map[string]RouteHandler
}

func (r RouteDef) String() string {
	var sb strings.Builder
	methods := getSortedKeys(r.MethodToHandlerMap)
	for _, method := range methods {
		sb.WriteString(fmt.Sprintf("%s %s\n", method, r.Path))
	}
	return sb.String()
}

func getSortedKeys[T any](m map[string]T) (keys []string) {
	keys = make([]string, len(m))
	var i int
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func (r *RouteDef) Get(handler Handler) *RouteDef {
	r.MethodToHandlerMap[http.MethodGet] = RouteHandler{
		name:    fmt.Sprintf("GET %v", r.Path),
		Handler: handler,
	}
	return r
}

func (r *RouteDef) Post(handler Handler) *RouteDef {
	r.MethodToHandlerMap[http.MethodPost] = RouteHandler{
		name:    fmt.Sprintf("POST %v", r.Path),
		Handler: handler,
	}
	return r
}

type RouteHandler struct {
	name    string
	Handler Handler
}

func (rh RouteHandler) Name() string {
	return rh.name
}

type Handler interface {
	http.Handler
	Models() (request Model, responses map[int]Model)
}

type Named interface {
	Name() string
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
