package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/respond"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/a-h/rest"
	"github.com/a-h/rest/examples/stdlib/handlers/topic/post"
	"github.com/a-h/rest/examples/stdlib/handlers/topics/get"
	"github.com/a-h/rest/swaggerui"
)

func main() {
	// Create standard routes.
	router := http.NewServeMux()
	router.Handle("/topics", &get.Handler{})
	router.Handle("/topic", &post.Handler{})

	api := rest.NewAPI("messages")
	api.StripPkgPaths = []string{"github.com/a-h/rest/example", "github.com/a-h/respond"}

	// It's possible to customise the OpenAPI schema for each type.
	// You can use helper functions, or write your own function that works
	// directly on the openapi3.Schema type.
	api.RegisterModel(rest.ModelOf[respond.Error](), rest.WithDescription("Standard JSON error"), func(s *openapi3.Schema) {
		status := s.Properties["statusCode"]
		status.Value.WithMin(100).WithMax(600)
	})

	api.Get("/topics").
		HasResponseModel(http.StatusOK, rest.ModelOf[get.TopicsGetResponse]()).
		HasResponseModel(http.StatusInternalServerError, rest.ModelOf[respond.Error]())

	api.Post("/topic").
		HasRequestModel(rest.ModelOf[post.TopicPostRequest]()).
		HasResponseModel(http.StatusOK, rest.ModelOf[post.TopicPostResponse]()).
		HasResponseModel(http.StatusInternalServerError, rest.ModelOf[respond.Error]())

	// Create the spec.
	spec, err := api.Spec()
	if err != nil {
		log.Fatalf("failed to create spec: %v", err)
	}

	// Apply any global customisation.
	spec.Info.Version = "v1.0.0."
	spec.Info.Description = "Messages API"

	// Attach the Swagger UI handler to your router.
	ui, err := swaggerui.New(spec)
	if err != nil {
		log.Fatalf("failed to create swagger UI handler: %v", err)
	}
	router.Handle("/swagger-ui", ui)
	router.Handle("/swagger-ui/", ui)

	// And start listening.
	fmt.Println("Listening on :8080...")
	fmt.Println("Visit http://localhost:8080/swagger-ui to see API definitions")
	fmt.Println("Listening on :8080...")
	http.ListenAndServe(":8080", router)
}
