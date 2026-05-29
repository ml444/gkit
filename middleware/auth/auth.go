package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
)

var (
	ErrUnauthorized = errorx.CreateError(401, 40101, "AUTH: unauthorized")
	ErrForbidden    = errorx.CreateError(403, 40301, "AUTH: forbidden")
)

type ctxKey struct{}

// Claims holds authenticated principal data.
type Claims map[string]any

// FromContext returns auth claims from context.
func FromContext(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(Claims)
	return c, ok
}

// TokenValidator validates a bearer token or API key.
type TokenValidator interface {
	Validate(ctx context.Context, token string) (Claims, error)
}

// Options configures auth middleware.
type Options struct {
	APIKeyHeader   string
	SkipPaths      map[string]bool
	TokenValidator TokenValidator
}

// Option configures Options.
type Option func(*Options)

func WithAPIKeyHeader(h string) Option {
	return func(o *Options) { o.APIKeyHeader = h }
}

func WithSkipPaths(paths ...string) Option {
	return func(o *Options) {
		if o.SkipPaths == nil {
			o.SkipPaths = make(map[string]bool)
		}
		for _, p := range paths {
			o.SkipPaths[p] = true
		}
	}
}

func WithValidator(v TokenValidator) Option {
	return func(o *Options) { o.TokenValidator = v }
}

func applyOpts(opts []Option) Options {
	o := Options{APIKeyHeader: "X-API-Key"}
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// Server returns service middleware enforcing authentication.
func Server(opts ...Option) middleware.Middleware {
	o := applyOpts(opts)
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if o.TokenValidator == nil {
				return next(ctx, req)
			}
			token := tokenFromContext(ctx, o.APIKeyHeader)
			if token == "" {
				return nil, ErrUnauthorized
			}
			claims, err := o.TokenValidator.Validate(ctx, token)
			if err != nil {
				return nil, ErrUnauthorized
			}
			ctx = context.WithValue(ctx, ctxKey{}, claims)
			return next(ctx, req)
		}
	}
}

// HTTPMiddleware enforces Bearer or API key auth on HTTP requests.
func HTTPMiddleware(opts ...Option) middleware.HttpMiddleware {
	o := applyOpts(opts)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if o.SkipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			if o.TokenValidator == nil {
				next.ServeHTTP(w, r)
				return
			}
			token := extractToken(r, o.APIKeyHeader)
			if token == "" {
				http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}
			claims, err := o.TokenValidator.Validate(r.Context(), token)
			if err != nil {
				http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request, apiKeyHeader string) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if apiKeyHeader != "" {
		if k := r.Header.Get(apiKeyHeader); k != "" {
			return k
		}
	}
	return ""
}

func tokenFromContext(ctx context.Context, apiKeyHeader string) string {
	if tr, ok := transport.FromContext(ctx); ok {
		if md := tr.In(); md != nil {
			if v := md.Get("Authorization"); len(v) > 0 {
				if strings.HasPrefix(v[0], "Bearer ") {
					return strings.TrimPrefix(v[0], "Bearer ")
				}
				return v[0]
			}
			if apiKeyHeader != "" {
				if v := md.Get(apiKeyHeader); len(v) > 0 {
					return v[0]
				}
			}
		}
	}
	return ""
}

// StaticValidator validates against a fixed token map.
type StaticValidator map[string]Claims

func (v StaticValidator) Validate(_ context.Context, token string) (Claims, error) {
	if c, ok := v[token]; ok {
		return c, nil
	}
	return nil, ErrUnauthorized
}
