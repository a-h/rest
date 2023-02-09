package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/rest"
	"github.com/a-h/rest/example/handlers/topic/post"
	"github.com/a-h/rest/example/handlers/topics/get"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	api := rest.NewAPI("messages")
	api.StripPkgPaths = []string{"github.com/a-h/rest/example"}

	api.Handle("/topics", &get.Handler{}).
		WithResponseModel(http.MethodGet, http.StatusOK, rest.ModelOf[get.TopicsGetResponse]())

	api.Handle("/topic", &post.Handler{}).
		WithRequestModel(http.MethodPost, rest.ModelOf[post.TopicPostRequest]()).
		WithResponseModel(http.MethodPost, http.StatusOK, rest.ModelOf[post.TopicPostResponse]())

	api.ConfigureSpec(func(spec *openapi3.T) {
		spec.Info.Version = "v1.0.0"
		spec.Info.Description = "Messages API"
	})

	fmt.Println("Listening on :8080...")
	http.ListenAndServe(":8080", api)
}
