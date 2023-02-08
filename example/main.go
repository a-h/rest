package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/a-h/rest"
	"github.com/a-h/rest/example/handlers/topic/post"
	"github.com/a-h/rest/example/handlers/topics/get"
)

func main() {
	api := rest.API("messages",
		rest.Route("/topics").Get(&get.Handler{}),
		rest.Route("/topic").Post(&post.Handler{}),
	)

	spec, err := api.Spec()
	if err != nil {
		log.Fatalf("failed to generate spec: %v", err)
	}
	v, _ := json.MarshalIndent(spec, "", " ")

	fmt.Print(string(v))
}
