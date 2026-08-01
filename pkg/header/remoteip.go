package header

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/ml444/gkit/transport"
)

// ClientIPOptions configures client IP extraction behind proxies.
// ⚠️ 安全警告：如果你的服务直接暴露在公网（没有反向代理），或者
// 前置反代未配置 X-Forwarded-For 的重写（Strip/Overwrite），
// 请勿开启 TrustForwarded。否则，攻击者可以通过伪造 HTTP 头部绕过 IP 过滤器。
type ClientIPOptions struct {
	// TrustForwarded enables X-Forwarded-For / X-Real-IP / CDN headers.
	TrustForwarded bool
	TrustedProxies []*net.IPNet // 仅当 RemoteAddr 属于这些段时才采信转发头
}

// ClientIPFromRequest returns the client IP from an HTTP request.
func ClientIPFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	return ClientIPFromHeaders(r.Header, r.RemoteAddr, ClientIPOptions{})
}

// ClientIPFromHeaders 提取客户端真实 IP
// extracts client IP from headers and remote address.
func ClientIPFromHeaders(h http.Header, remoteAddr string, opt ClientIPOptions) string {
	remoteIPStr := hostFromAddr(remoteAddr)
	if !opt.TrustForwarded {
		return remoteIPStr
	}

	// 如果配置了 TrustedProxies，必须先校验 RemoteAddr 是否属于可信代理
	if len(opt.TrustedProxies) > 0 {
		remoteIP := net.ParseIP(remoteIPStr)
		if remoteIP == nil || !isTrustedProxy(remoteIP, opt.TrustedProxies) {
			// 发起请求的直接连接 IP 不在可信代理列表中，不能采信任何 Header
			return remoteIPStr
		}
	}

	// 1. 优先尝试高优先级且通常由单一可信网关写入的单值 Header (如 CDN、AppEngine 等)
	// 注意：如果这些 Header 可由客户端伪造，且未在最外层网关剥离，仍存在风险。
	if ip := firstHeader(h, HeaderCFConnectingIP); ip != "" {
		return ip
	}
	if ip := firstHeader(h, HeaderXAppEngineRemoteIP); ip != "" {
		return ip
	}

	// 2. 解析 X-Forwarded-For (从右向左安全剥离)
	if xff := h.Get(HeaderXForwardedFor); xff != "" {
		if ip := getIPFromXFF(xff, opt.TrustedProxies); ip != "" {
			return ip
		}
	}

	if ip := firstHeader(h, HeaderXRealIP); ip != "" {
		return ip
	}
	if ip := firstHeader(h, RemoteIPKey); ip != "" {
		return ip
	}

	return remoteIPStr
}

// getIPFromXFF 从右向左遍历 XFF，找到第一个非可信代理的 IP
func getIPFromXFF(xff string, trustedProxies []*net.IPNet) string {
	parts := strings.Split(xff, ",")
	// 从右往左遍历
	for i := len(parts) - 1; i >= 0; i-- {
		ipStr := strings.TrimSpace(parts[i])
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		// 如果没有配置可信代理，或者当前 IP 不属于可信代理，则当前 IP 就是客户端真实 IP
		if len(trustedProxies) == 0 || !isTrustedProxy(ip, trustedProxies) {
			return ipStr
		}
	}
	// 如果全部都是可信代理（理论上极少发生，说明最左侧也是代理），退而求其次返回最左侧的 IP
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}

// isTrustedProxy 判断 IP 是否在可信网段内
func isTrustedProxy(ip net.IP, trustedProxies []*net.IPNet) bool {
	for _, ipNet := range trustedProxies {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
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
