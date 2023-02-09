package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/rest"
	"github.com/a-h/rest/example/handlers/topic/post"
	"github.com/a-h/rest/example/handlers/topics/get"
)

func main() {
	api := rest.NewAPI("messages")

	api.Handle("/topics", &get.Handler{}).
		WithResponseModel(http.MethodGet, http.StatusOK, rest.ModelOf[get.TopicsGetResponse]())

	api.Handle("/topic", &post.Handler{}).
		WithRequestModel(http.MethodPost, rest.ModelOf[post.TopicPostRequest]()).
		WithResponseModel(http.MethodPost, http.StatusOK, rest.ModelOf[post.TopicPostResponse]())

	spec, err := api.Spec()
	if err != nil {
		log.Fatalf("failed to generate spec: %v", err)
	}
	v, _ := json.MarshalIndent(spec, "", " ")

	fmt.Print(string(v))
}
