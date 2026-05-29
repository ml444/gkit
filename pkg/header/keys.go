package header

// Well-known HTTP / metadata keys for cross-service propagation.
// Prefer these constants instead of literal strings in middleware and transport code.
const (
	TraceIDKey    = "X-Trace-Id"
	RequestIDKey  = "X-Request-ID"
	RemoteIPKey   = "X-Remote-Ip"
	HTTPMethodKey = "X-Http-Method"
	HTTPPathKey   = "X-Http-Path"

	ClientIDKey   = "X-Client-Id"
	ClientTypeKey = "X-Client-Type"

	// Standard proxy / CDN client IP headers.
	HeaderCFConnectingIP     = "CF-Connecting-IP"
	HeaderXAppEngineRemoteIP = "X-Appengine-Remote-Addr"
	HeaderXForwardedFor      = "X-Forwarded-For"
	HeaderXRealIP            = "X-Real-IP"
)

// traceKeyVariants lists common spellings for trace ID lookup (HTTP + gRPC metadata).
var traceKeyVariants = []string{TraceIDKey, "x-trace-id", "X-Trace-ID"}

// requestIDKeyVariants lists common spellings for request ID lookup.
var requestIDKeyVariants = []string{RequestIDKey, "x-request-id", "X-Request-Id"}
