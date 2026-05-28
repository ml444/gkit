package httpx

import "strings"

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
