package post

import (
	"net/http"

	"github.com/the-pawn-2017/rest/examples/stdlib/models"
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
