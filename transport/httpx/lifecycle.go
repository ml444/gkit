package httpx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"

	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/internal/lifecycle"
)

// Managed takes exclusive ownership of s's lifecycle. Call it after configuring
// the server and before using Start/Stop. Legacy Endpoint-created listeners may
// be adopted. Use the returned adapter for all subsequent lifecycle operations;
// direct Serve/Shutdown/Close calls on the embedded http.Server are unsupported.
func Managed(s *Server, opts transport.LifecycleOptions) (transport.ManagedServer, error) {
	if s == nil || s.Server == nil {
		return nil, errors.New("httpx: Managed requires an initialized server")
	}
	c, err := lifecycle.New(opts, lifecycle.Hooks{
		Bind: func(ctx context.Context) (*url.URL, error) {
			if err := s.bind(ctx, true); err != nil {
				return nil, err
			}
			return s.endpointSnapshot(), nil
		},
		Serve: func(ctx context.Context) error {
			return s.serve(context.WithoutCancel(ctx))
		},
		Shutdown: func(ctx context.Context) error {
			err := s.Shutdown(ctx)
			if err != nil {
				err = errors.Join(err, s.Close())
			}
			return errors.Join(err, lifecycle.CloseListener(s.listenerSnapshot()))
		},
		IsNormalStop: func(err error) bool {
			return errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed)
		},
	})
	if err != nil {
		return nil, err
	}
	s.bindMu.Lock()
	defer s.bindMu.Unlock()
	s.resourceMu.Lock()
	defer s.resourceMu.Unlock()
	if s.managed || s.legacyUsed {
		return nil, transport.ErrLifecycleOwned
	}
	s.managed = true
	// Detach an endpoint supplied by the caller before publishing snapshots.
	s.endpoint = lifecycle.CloneURL(s.endpoint)
	return c, nil
}
