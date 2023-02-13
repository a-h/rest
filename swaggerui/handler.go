package swaggerui

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:embed swagger-ui/*
var swaggerUI embed.FS

func New(spec *openapi3.T) (h http.Handler, err error) {
	specBytes, err := json.MarshalIndent(spec, "", " ")
	if err != nil {
		return h, fmt.Errorf("swaggerui: failed to marshal specification: %w", err)
	}

	m := http.NewServeMux()
	m.Handle("/", http.FileServer(http.FS(swaggerUI)))
	m.HandleFunc("/swagger-ui/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write(specBytes)
	})

	return m, nil
}
