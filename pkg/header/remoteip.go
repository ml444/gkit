package header

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/ml444/gkit/transport"
)

// ClientIPOptions configures client IP extraction behind proxies.
type ClientIPOptions struct {
	// TrustForwarded enables X-Forwarded-For / X-Real-IP / CDN headers.
	TrustForwarded bool
}

// ClientIPFromRequest returns the client IP from an HTTP request.
func ClientIPFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	return ClientIPFromHeaders(r.Header, r.RemoteAddr, ClientIPOptions{TrustForwarded: true})
}

// ClientIPFromHeaders extracts client IP from headers and remote address.
func ClientIPFromHeaders(h http.Header, remoteAddr string, opt ClientIPOptions) string {
	if opt.TrustForwarded {
		if ip := firstHeader(h, HeaderCFConnectingIP); ip != "" {
			return ip
		}
		if ip := firstHeader(h, HeaderXAppEngineRemoteIP); ip != "" {
			return ip
		}
		if ip := forwardedForIP(h.Get(HeaderXForwardedFor)); ip != "" {
			return ip
		}
		if ip := firstHeader(h, HeaderXRealIP); ip != "" {
			return ip
		}
		if ip := firstHeader(h, RemoteIPKey); ip != "" {
			return ip
		}
	}
	return hostFromAddr(remoteAddr)
}

// ClientIPFromContext resolves client IP from transport metadata or the underlying HTTP request.
func ClientIPFromContext(ctx context.Context) string {
	if ip := FirstIncoming(ctx, RemoteIPKey, HeaderXRealIP, HeaderCFConnectingIP); ip != "" {
		return ip
	}
	if xff := FirstIncoming(ctx, HeaderXForwardedFor); xff != "" {
		if ip := forwardedForIP(xff); ip != "" {
			return ip
		}
	}
	tr, ok := transport.FromContext(ctx)
	if !ok {
		return ""
	}
	if c, ok := tr.(interface{ Request() *http.Request }); ok {
		if req := c.Request(); req != nil {
			return ClientIPFromRequest(req)
		}
	}
	return ""
}

func forwardedForIP(xff string) string {
	if xff == "" {
		return ""
	}
	parts := strings.Split(xff, ",")
	// Prefer the left-most valid IP (original client when proxies append).
	for _, part := range parts {
		ipStr := strings.TrimSpace(part)
		if ip := net.ParseIP(ipStr); ip != nil {
			return ipStr
		}
	}
	return ""
}

func hostFromAddr(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
