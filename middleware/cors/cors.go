package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ml444/gkit/middleware"
)

// Options configures CORS behavior.
type Options struct {
	AllowOrigins     []string
	AllowMethods     string
	AllowHeaders     string
	ExposeHeaders    string
	AllowCredentials bool
	MaxAge           int
}

// Default returns permissive CORS for development.
func Default() middleware.HttpMiddleware {
	return New(Options{
		AllowOrigins: []string{"*"},
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS",
		AllowHeaders: "Content-Type,Authorization,Accept,Accept-Language,X-Request-ID",
	})
}

// New returns CORS middleware with the given options.
func New(opt Options) middleware.HttpMiddleware {
	methods := opt.AllowMethods
	if methods == "" {
		methods = "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS"
	}
	headers := opt.AllowHeaders
	if headers == "" {
		headers = "Content-Type,Authorization,Accept,Accept-Language,X-Request-ID"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowOrigin := pickOrigin(origin, opt.AllowOrigins)
			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				w.Header().Add("Vary", "Origin")
			}
			if opt.AllowCredentials && allowOrigin != "*" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if methods != "" {
				w.Header().Set("Access-Control-Allow-Methods", methods)
			}
			if headers != "" {
				w.Header().Set("Access-Control-Allow-Headers", headers)
			}
			if opt.ExposeHeaders != "" {
				w.Header().Set("Access-Control-Expose-Headers", opt.ExposeHeaders)
			}
			if opt.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(opt.MaxAge))
			}
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func pickOrigin(requestOrigin string, allowed []string) string {
	if len(allowed) == 0 {
		return ""
	}
	for _, o := range allowed {
		if o == "*" {
			if requestOrigin != "" {
				return requestOrigin
			}
			return "*"
		}
		if strings.EqualFold(o, requestOrigin) {
			return requestOrigin
		}
	}
	return ""
}
