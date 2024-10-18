package chiadapter_test

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"

	"github.com/a-h/rest"
	"github.com/a-h/rest/chiadapter"
)

func TestMerge(t *testing.T) {
	// Arrange.
	pattern := `/organisation/{orgId:\d+}/user/{userId}/{role}`
	router := chi.NewRouter()
	router.Method(http.MethodGet, pattern,
		http.RedirectHandler("/elsewhere", http.StatusMovedPermanently))
	api := rest.NewAPI("test")

	// Act.
	err := chiadapter.Merge(api, router)
	if err != nil {
		t.Fatalf("failed to merge: %v", err)
	}
	api.Get(pattern).HasPathParameter("role", rest.PathParam{
		Description: "Role of the user",
	})

	// Assert.
	expected := rest.Params{
		Path: map[string]rest.PathParam{
			"orgId":  {Regexp: `\d+`},
			"userId": {},
			"role":   {Description: "Role of the user"},
		},
		Query: make(map[string]rest.QueryParam),
	}
	if diff := cmp.Diff(expected, api.Get(pattern).Params); diff != "" {
		t.Error(diff)
	}
}
