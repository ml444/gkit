package response

import "net/http"

// HttpRawHeader marks responses that must not be wrapped by WrapHttpResponse.
// Generated protoc-gen-go-http handlers for raw/pluck routes set this header.
const HttpRawHeader = "X-Gkit-Http-Raw"

// MarkHttpRaw marks w as a raw HTTP response (binary or non-JSON body).
func MarkHttpRaw(w http.ResponseWriter) {
	w.Header().Set(HttpRawHeader, "1")
}

// IsHttpRaw reports whether headers indicate a raw HTTP response.
func IsHttpRaw(h http.Header) bool {
	return h.Get(HttpRawHeader) == "1"
}
