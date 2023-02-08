package get

import (
	"net/http"

	"github.com/a-h/rest"
	"github.com/a-h/rest/example/models"
)

type TopicsGetResponse struct {
	Topics []models.Topic `json:"topics"`
}

type Handler struct {
}

func (h *Handler) Models() (request rest.Model, responses map[int]rest.Model) {
	responses = rest.Responses(
		rest.Response[TopicsGetResponse](http.StatusOK),
	)
	return
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
