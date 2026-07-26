package ipfilter

import (
	"net"
	"net/http"
	"strings"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
)

var ErrDenied = errorx.CreateError(403, 40303, "IPFILTER: access denied")

// Options configures IP filtering.
type Options struct {
	AllowList      []*net.IPNet
	DenyList       []*net.IPNet
	TrustXFF       bool
	TrustedProxies []string
}

// HTTPMiddleware blocks requests from denied IP ranges.
// ⚠️ 安全警告：如果你的服务直接暴露在公网（没有反向代理），
// 或者前置反代未配置 X-Forwarded-For 的重写（Strip/Overwrite），
// 请勿开启 TrustXFF。否则，攻击者可以通过伪造 HTTP 头部绕过 IP 过滤器。
func HTTPMiddleware(opt Options) middleware.HttpMiddleware {
	// 在 ipfilter 中间件初始化时预解析 CIDR 列表
	var trustedNets []*net.IPNet
	for _, cidr := range opt.TrustedProxies {
		if _, ipNet, err := net.ParseCIDR(cidr); err == nil {
			trustedNets = append(trustedNets, ipNet)
		} else if ip := net.ParseIP(cidr); ip != nil {
			// 支持单 IP 配置，将其转化为 /32 或 /128
			mask := net.CIDRMask(32, 32)
			if ip.To4() == nil {
				mask = net.CIDRMask(128, 128)
			}
			trustedNets = append(trustedNets, &net.IPNet{IP: ip, Mask: mask})
		}
	}
	cfg := header.ClientIPOptions{
		TrustForwarded: opt.TrustXFF,
		TrustedProxies: trustedNets,
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := header.ClientIPFromHeaders(r.Header, r.RemoteAddr, cfg)
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

// ParseCIDRs parses CIDR strings and bare IPs into IP nets.
func ParseCIDRs(cidrs ...string) ([]*net.IPNet, error) {
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		// 如果不包含 '/'，说明可能是 bare IP
		if !strings.Contains(c, "/") {
			if ip := net.ParseIP(c); ip != nil {
				// 判断是 IPv4 还是 IPv6
				if ip.To4() != nil {
					c += "/32"
				} else {
					c += "/128"
				}
			}
		}
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}
