package main

import (
	"fmt"
	"net/http"
	"strings"
)

// cookieKeeperModifier rewrites Set-Cookie headers from the upstream sGTM response,
// extending Max-Age to defeat Safari ITP and short cookie lifetimes.
// This mimics Stape's Cookie Keeper power-up.
func cookieKeeperModifier(cfg *CookieKeeperConfig) func(*http.Response) error {
	return func(resp *http.Response) error {
		cookies := resp.Header.Values("Set-Cookie")
		if len(cookies) == 0 {
			return nil
		}

		resp.Header.Del("Set-Cookie")
		for _, raw := range cookies {
			resp.Header.Add("Set-Cookie", rewriteCookie(raw, cfg.MaxAge))
		}
		return nil
	}
}

// rewriteCookie parses a raw Set-Cookie header string and sets/replaces Max-Age.
// It preserves all other attributes (Path, Domain, SameSite, Secure, HttpOnly).
func rewriteCookie(raw string, maxAge int) string {
	parts := strings.Split(raw, ";")
	var result []string
	hasMaxAge := false

	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "max-age=") {
			result = append(result, fmt.Sprintf(" Max-Age=%d", maxAge))
			hasMaxAge = true
			continue
		}

		// Remove Expires if we're setting Max-Age (Max-Age takes precedence)
		if strings.HasPrefix(lower, "expires=") && hasMaxAge {
			continue
		}

		if i == 0 {
			result = append(result, trimmed)
		} else {
			result = append(result, " "+trimmed)
		}
	}

	if !hasMaxAge {
		result = append(result, fmt.Sprintf(" Max-Age=%d", maxAge))
	}

	return strings.Join(result, ";")
}
