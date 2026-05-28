package httpx

import (
	"net/http"

	"github.com/ml444/gkit/transport/httpx/pprof"
)

// RegisterHealth mounts a simple health endpoint.
func (s *Server) RegisterHealth(path string) {
	if s == nil || s.router == nil {
		return
	}
	if path == "" {
		path = "/healthz"
	}
	s.router.GET(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// RegisterPprof mounts net/http/pprof handlers under /debug/pprof/.
func (s *Server) RegisterPprof() {
	if s == nil || s.router == nil {
		return
	}
	s.router.HandlePrefix("/debug/pprof/", pprof.NewHandler())
}

