package ipfilter

import (
	"net"
	"net/http"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
)

var ErrDenied = errorx.CreateError(403, 40303, "IPFILTER: access denied")

// Options configures IP filtering.
type Options struct {
	AllowList []*net.IPNet
	DenyList  []*net.IPNet
	TrustXFF  bool
}

// HTTPMiddleware blocks requests from denied IP ranges.
func HTTPMiddleware(opt Options) middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := header.ClientIPFromHeaders(r.Header, r.RemoteAddr, header.ClientIPOptions{TrustForwarded: opt.TrustXFF})
			parsed := net.ParseIP(ip)
			if parsed == nil {
				http.Error(w, ErrDenied.Error(), http.StatusForbidden)
				return
			}
			if matchNets(parsed, opt.DenyList) {
				http.Error(w, ErrDenied.Error(), http.StatusForbidden)
				return
			}
			if len(opt.AllowList) > 0 && !matchNets(parsed, opt.AllowList) {
				http.Error(w, ErrDenied.Error(), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func matchNets(ip net.IP, nets []*net.IPNet) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ParseCIDRs parses CIDR strings into IP nets.
func ParseCIDRs(cidrs ...string) ([]*net.IPNet, error) {
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}
