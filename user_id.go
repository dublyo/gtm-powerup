package main

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// userIDMiddleware generates a deterministic user identifier from IP + User-Agent + salt,
// then injects it as a request header for sGTM to consume.
// This mimics Stape's User ID power-up for cookieless tracking.
func userIDMiddleware(cfg *UserIDConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		ua := r.UserAgent()

		raw := ip + "|" + ua + "|" + cfg.Salt
		hash := sha256.Sum256([]byte(raw))
		uid := fmt.Sprintf("%x", hash[:8]) // first 16 hex chars

		r.Header.Set(cfg.Header, uid)
		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the real client IP from standard proxy headers.
func clientIP(r *http.Request) string {
	// CF-Connecting-IP (Cloudflare)
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	// X-Real-IP (Nginx/Traefik)
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// X-Forwarded-For (first entry)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
