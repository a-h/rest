package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/respond"
	"github.com/a-h/rest"
	"github.com/a-h/rest/chiadapter"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
)

type Topic struct {
	Namespace string `json:"namespace"`
	Topic     string `json:"topic"`
	Private   bool   `json:"private"`
	ViewCount int64  `json:"viewCount"`
}

type TopicsPostRequest struct {
	Topic
}

type TopicsPostResponse struct {
	ID string `json:"id"`
}

type TopicsGetResponse struct {
	Topics []TopicRecord `json:"topics"`
}

type TopicRecord struct {
	ID string `json:"id"`
	Topic
}

func main() {
	// Define routes in any router.
	router := chi.NewRouter()

	router.Get("/topic/{id}", func(w http.ResponseWriter, r *http.Request) {
		resp := Topic{
			Namespace: "example",
			Topic:     "topic",
			Private:   false,
			ViewCount: 412,
		}
		respond.WithJSON(w, resp, http.StatusOK)
	})

	router.Get("/topics", func(w http.ResponseWriter, r *http.Request) {
		resp := TopicsGetResponse{
			Topics: []TopicRecord{
				{
					ID: "testId",
					Topic: Topic{
						Namespace: "example",
						Topic:     "topic",
						Private:   false,
						ViewCount: 412,
					},
				},
			},
		}
		respond.WithJSON(w, resp, http.StatusOK)
	})

	router.Post("/topics", func(w http.ResponseWriter, r *http.Request) {
		resp := TopicsPostResponse{ID: "123"}
		respond.WithJSON(w, resp, http.StatusOK)
	})

	// Create the API definition.
	api := rest.NewAPI("Messaging API")

	// Create the routes and parameters of the Router in the REST API definition with an
	// adapter, or do it manually.
	chiadapter.Merge(api, router)

	// Because this example is all in the main package, we can strip the `main_` namespace from
	// the types.
	api.StripPkgPaths = []string{"main", "github.com/a-h"}

	// It's possible to customise the OpenAPI schema for each type.
	api.RegisterModel(rest.ModelOf[respond.Error](), rest.WithDescription("Standard JSON error"), func(s *openapi3.Schema) {
		status := s.Properties["statusCode"]
		status.Value.WithMin(100).WithMax(600)
	})

	// Document the routes.
	api.Get("/topic/{id}").
		HasResponseModel(http.StatusOK, rest.ModelOf[TopicsGetResponse]()).
		HasResponseModel(http.StatusInternalServerError, rest.ModelOf[respond.Error]())

	api.Get("/topics").
		HasResponseModel(http.StatusOK, rest.ModelOf[TopicsGetResponse]()).
		HasResponseModel(http.StatusInternalServerError, rest.ModelOf[respond.Error]())

	api.Post("/topics").
		HasRequestModel(rest.ModelOf[TopicsPostRequest]()).
		HasResponseModel(http.StatusOK, rest.ModelOf[TopicsPostResponse]()).
		HasResponseModel(http.StatusInternalServerError, rest.ModelOf[respond.Error]())

	api.ConfigureSpec(func(spec *openapi3.T) {
		spec.Info.Version = "v1.0.0"
		spec.Info.Description = "Messages API"
	})

	// Attach the swagger UI definition to your router.
	router.Handle("/swagger-ui*", api)
	// And start listening.
	fmt.Println("Listening on :8080...")
	fmt.Println("Visit http://localhost:8080/swagger-ui to see API definitions")
	http.ListenAndServe(":8080", router)
}
