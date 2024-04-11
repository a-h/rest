package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/a-h/respond"
	"github.com/aviva-verde/rest"
	"github.com/aviva-verde/rest/examples/offline/models"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	// Configure the models.
	api := rest.NewAPI("messages")
	api.StripPkgPaths = []string{"github.com/aviva-verde/rest/example", "github.com/a-h/respond"}

	api.RegisterModel(rest.ModelOf[respond.Error](), rest.WithDescription("Standard JSON error"), func(s *openapi3.Schema) {
		status := s.Properties["statusCode"]
		status.Value.WithMin(100).WithMax(600)
	})

	api.Get("/topic/{id}").
		HasPathParameter("id", rest.PathParam{
			Description: "id of the topic",
			Regexp:      `\d+`,
		}).
		HasResponseModel(http.StatusOK, rest.ModelOf[models.Topic]()).
		HasResponseModel(http.StatusInternalServerError, rest.ModelOf[respond.Error]()).
		HasTags([]string{"Topic"}).
		HasDescription("Get one topic by id").
		HasOperationID("getOneTopic")

	// Create the specification.
	spec, err := api.Spec()
	if err != nil {
		log.Fatalf("failed to create spec: %v", err)
	}

	// Write to stdout.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(spec)
}
