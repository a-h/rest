package chiadapter

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/a-h/rest"
	"github.com/go-chi/chi/v5"
)

func Merge(target *rest.API, src chi.Router) error {
	walker := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		params, err := getParams(route)
		if err != nil {
			return err
		}
		r := rest.Route{
			Method:  rest.Method(method),
			Pattern: rest.Pattern(route),
			Params:  params,
		}
		target.Merge(r)
		return nil
	}

	return chi.Walk(src, walker)
}

func getParams(s string) (p rest.Params, err error) {
	p.Path = make(map[string]rest.PathParam)
	p.Query = make(map[string]rest.QueryParam)

	u, err := url.Parse(s)
	if err != nil {
		return
	}

	// Path.
	s = u.Path
	s = strings.TrimSuffix(s, "/")
	s = strings.TrimPrefix(s, "/")
	segments := strings.Split(s, "/")
	for _, segment := range segments {
		name, pattern, ok := getPlaceholder(segment)
		if !ok {
			continue
		}
		p.Path[name] = rest.PathParam{
			Regexp: pattern,
		}
	}

	// Query.
	q := u.Query()
	for k := range q {
		name, _, ok := getPlaceholder(q.Get(k))
		if !ok {
			continue
		}
		p.Query[name] = rest.QueryParam{
			Description: "",
			Required:    false,
			AllowEmpty:  false,
		}
	}

	return
}

func getPlaceholder(s string) (name string, pattern string, ok bool) {
	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return
	}
	parts := strings.SplitN(s[1:len(s)-1], ":", 2)
	name = parts[0]
	if len(parts) > 1 {
		pattern = parts[1]
	}
	return name, pattern, true
}
