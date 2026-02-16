package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	upstream := os.Getenv("UPSTREAM_URL")
	if upstream == "" {
		upstream = "http://sgtm:8080"
	}
	target, err := url.Parse(upstream)
	if err != nil {
		log.Fatalf("Invalid UPSTREAM_URL: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %s %s: %v", r.Method, r.URL.Path, err)
		w.WriteHeader(http.StatusBadGateway)
	}

	// Build response modifier for Cookie Keeper
	if cfg.CookieKeeper != nil && cfg.CookieKeeper.Enabled {
		proxy.ModifyResponse = cookieKeeperModifier(cfg.CookieKeeper)
		log.Printf("Cookie Keeper enabled (maxAge=%ds)", cfg.CookieKeeper.MaxAge)
	}

	// Build middleware chain (innermost runs first)
	var handler http.Handler = proxy

	if cfg.UserID != nil && cfg.UserID.Enabled {
		handler = userIDMiddleware(cfg.UserID, handler)
		log.Printf("User ID enabled (header=%s)", cfg.UserID.Header)
	}

	if cfg.BotDetection != nil && cfg.BotDetection.Enabled {
		handler = botDetectionMiddleware(cfg.BotDetection, handler)
		log.Printf("Bot Detection enabled (block=%v, header=%s)", cfg.BotDetection.BlockBots, cfg.BotDetection.HeaderName)
	}

	if cfg.IPBlocklist != nil && cfg.IPBlocklist.Enabled && len(cfg.IPBlocklist.IPs) > 0 {
		handler = ipBlocklistMiddleware(cfg.IPBlocklist, handler)
		log.Printf("IP Blocklist enabled (%d entries)", len(cfg.IPBlocklist.IPs))
	}

	// Health check endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.Handle("/", handler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("GTM Power-ups proxy starting on :%s -> %s", port, upstream)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
