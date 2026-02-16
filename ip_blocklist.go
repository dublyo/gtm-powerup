package main

import (
	"net"
	"net/http"
)

// ipBlocklistMiddleware blocks requests from IPs or CIDR ranges
// configured in the blocklist. Returns 403 for blocked IPs.
func ipBlocklistMiddleware(cfg *IPBlocklistConfig, next http.Handler) http.Handler {
	// Pre-parse CIDRs and single IPs at init time
	var nets []*net.IPNet
	var singles []net.IP

	for _, entry := range cfg.IPs {
		_, cidr, err := net.ParseCIDR(entry)
		if err == nil {
			nets = append(nets, cidr)
			continue
		}
		if ip := net.ParseIP(entry); ip != nil {
			singles = append(singles, ip)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipStr := clientIP(r)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Check single IPs
		for _, blocked := range singles {
			if blocked.Equal(ip) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Forbidden"))
				return
			}
		}

		// Check CIDRs
		for _, cidr := range nets {
			if cidr.Contains(ip) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Forbidden"))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
