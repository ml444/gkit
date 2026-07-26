package httpx

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// JoinPath joins a route prefix with a path template.
// prefix="/api/v1" + path="/users/{id}" => "/api/v1/users/{id}"
func JoinPath(prefix, path string) string {
	prefix = strings.TrimSuffix(prefix, "/")
	if prefix == "" {
		return path
	}
	if path == "" {
		return prefix
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return prefix + path
}

func RoutePattern(req *http.Request) string {
	if route := mux.CurrentRoute(req); route != nil {
		// /path/123 -> /path/{id}
		pathTemplate, _ := route.GetPathTemplate()
		return pathTemplate
	}
	return ""
}
