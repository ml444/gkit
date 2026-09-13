package grpcx

import (
	"context"
	"errors"
	"net"
	"net/url"

	"google.golang.org/grpc"

	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/internal/lifecycle"
)

// Managed takes exclusive ownership of s's lifecycle. Configure options and
// register services first. A listener created by legacy Endpoint may be adopted;
// subsequent Start/Stop/Endpoint and RegisterDiscovery calls must not be mixed
// with this adapter. It does not publish or deregister services automatically.
func Managed(s *Server, opts transport.LifecycleOptions) (transport.ManagedServer, error) {
	if s == nil || s.iServer == nil {
		return nil, errors.New("grpcx: Managed requires an initialized server")
	}
	c, err := lifecycle.New(opts, lifecycle.Hooks{
		Bind: func(ctx context.Context) (*url.URL, error) {
			if err := s.bind(ctx, true); err != nil {
				return nil, err
			}
			scheme := "grpc"
			if s.tlsConf != nil || (s.credentials != nil && s.credentials.Info().SecurityProtocol == "tls") {
				scheme = "grpcs"
			}
			return &url.URL{Scheme: scheme, Host: s.endpointSnapshot()}, nil
		},
		Serve: func(context.Context) error {
			return s.iServer.Serve(s.listenerSnapshot())
		},
		Shutdown: func(ctx context.Context) error {
			if s.health != nil {
				s.health.Shutdown()
			}
			done := make(chan struct{})
			go func() {
				s.GracefulStop()
				close(done)
			}()
			var err error
			select {
			case <-done:
			case <-ctx.Done():
				err = ctx.Err()
				s.iServer.Stop()
				<-done
			}
			return errors.Join(err, lifecycle.CloseListener(s.listenerSnapshot()))
		},
		IsNormalStop: func(err error) bool {
			return errors.Is(err, grpc.ErrServerStopped) || errors.Is(err, net.ErrClosed)
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
	return c, nil
}
