package csrf

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
)

var ErrCSRF = errorx.CreateError(403, 40302, "CSRF: token mismatch")

// Options configures CSRF protection.
type Options struct {
	CookieName string
	HeaderName string
	// SkipSafe skips GET/HEAD/OPTIONS when true (default true).
	SkipSafe bool
	// SkipBearer skips CSRF when Authorization: Bearer is present (default true).
	// Pure Bearer/token APIs do not need double-submit CSRF; enable CSRF only for cookie/session web apps.
	SkipBearer bool
}

// DefaultOptions returns options suited for browser cookie/session apps.
func DefaultOptions() Options {
	return Options{
		SkipSafe:   true,
		SkipBearer: true,
	}
}

// HTTPMiddleware validates double-submit CSRF token for unsafe methods.
func HTTPMiddleware(opt Options) middleware.HttpMiddleware {
	if opt.CookieName == "" {
		opt.CookieName = "gkit_csrf"
	}
	if opt.HeaderName == "" {
		opt.HeaderName = "X-CSRF-Token"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opt.SkipBearer && hasBearerAuth(r) {
				next.ServeHTTP(w, r)
				return
			}
			if opt.SkipSafe && isSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie(opt.CookieName)
			if err != nil || cookie.Value == "" {
				token := newToken()
				http.SetCookie(w, &http.Cookie{Name: opt.CookieName, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
				if isSafeMethod(r.Method) {
					w.Header().Set(opt.HeaderName, token)
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, ErrCSRF.Error(), http.StatusForbidden)
				return
			}
			if r.Header.Get(opt.HeaderName) != cookie.Value {
				http.Error(w, ErrCSRF.Error(), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasBearerAuth(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	return strings.HasPrefix(auth, "Bearer ") && len(strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))) > 0
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func newToken() string {
	var b [32]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
