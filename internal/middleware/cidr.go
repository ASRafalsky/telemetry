package middleware

import (
	"net"
	"net/http"
)

// CheckCIDR checks src IP address with CIDR.
func CheckCIDR(h http.Handler, header bool, cidr *net.IPNet) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cidr == nil {
			h.ServeHTTP(w, r)
			return
		}
		var srcIP string
		if header {
			srcIP = r.Header.Get("X-Real-IP")
		} else {
			srcIP = r.RemoteAddr
		}

		if srcIP == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if cidr.Contains(net.ParseIP(srcIP)) {
			h.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusForbidden)
	}
}
