# REST

Document a REST API with an OpenAPI 3.0 specification.

* Code, not configuration.
* No magic comments, tags, or decorators.
* Use with or without a Go web framework.
* Populates schema automatically using reflection.

## Why would I want to use this?

* Add OpenAPI documentation to an API.
  * Create a `swagger.json` or `swagger.yaml` file.
* Serve the Swagger UI to customers.

## Examples

See the [./examples](./examples) directory for complete examples.

### Create an OpenAPI 3.0 (swagger) file

```go
// Configure the models.
api := rest.NewAPI("messages")
api.StripPkgPaths = []string{"github.com/a-h/rest/example", "github.com/a-h/respond"}

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
  HasResponseModel(http.StatusInternalServerError, rest.ModelOf[respond.Error]())

// Create the specification.
spec, err := api.Spec()
if err != nil {
  log.Fatalf("failed to create spec: %v", err)
}

// Write to stdout.
enc := json.NewEncoder(os.Stdout)
enc.SetIndent("", " ")
enc.Encode(spec)
```

### Serve API documentation alongside your API

```go
// Create routes.
router := http.NewServeMux()
router.Handle("/topics", &get.Handler{})
router.Handle("/topic", &post.Handler{})

api := rest.NewAPI("messages")
api.StripPkgPaths = []string{"github.com/a-h/rest/example", "github.com/a-h/respond"}

// Register the error type with customisations.
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
```

## Tasks

### test

```
go test ./...
```

### run-example

Dir: ./examples/stdlib

```
go run main.go
```
