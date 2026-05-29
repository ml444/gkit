package security

import (
	"net/http"

	"github.com/ml444/gkit/middleware"
)

// Options configures security response headers.
type Options struct {
	HSTS                 string
	ContentTypeNoSniff   bool
	FrameOptions         string
	ReferrerPolicy       string
	PermissionsPolicy    string
}

// DefaultOptions returns recommended baseline security headers.
func DefaultOptions() Options {
	return Options{
		ContentTypeNoSniff: true,
		FrameOptions:       "DENY",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}
}

// HTTPMiddleware sets security headers on every response.
func HTTPMiddleware(opt Options) middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opt.HSTS != "" {
				w.Header().Set("Strict-Transport-Security", opt.HSTS)
			}
			if opt.ContentTypeNoSniff {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}
			if opt.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", opt.FrameOptions)
			}
			if opt.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", opt.ReferrerPolicy)
			}
			if opt.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", opt.PermissionsPolicy)
			}
			next.ServeHTTP(w, r)
		})
	}
}
