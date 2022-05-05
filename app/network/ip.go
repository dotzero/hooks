package network

import (
	"net"
	"net/http"
	"strings"
)

const (
	// canonical format via http.CanonicalHeaderKey
	xForwardedForHeader = "X-Forwarded-For"
	xRealIPHeader       = "X-Real-Ip"
)

// ClientIP returns an clients RemoteAddr of parsing either
// the X-Forwarded-For header or the X-Real-IP header
func ClientIP(r *http.Request) net.IP {
	var ip string

	if v := r.Header.Get(xForwardedForHeader); v != "" {
		i := strings.Index(v, ",")
		if i == -1 {
			i = len(v)
		}

		ip = v[:i]
	} else if v := r.Header.Get(xRealIPHeader); v != "" {
		ip = v
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return net.ParseIP(ip)
}
