package post

import (
	"net/http"

	"github.com/a-h/rest/example/models"
)

type TopicPostRequest struct {
	models.Topic
}

type TopicPostResponse struct {
	OK bool `json:"ok"`
}

type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
