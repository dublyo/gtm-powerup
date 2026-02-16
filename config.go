package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// PowerupsConfig is parsed from the POWERUPS_CONFIG environment variable (JSON).
type PowerupsConfig struct {
	CookieKeeper *CookieKeeperConfig `json:"cookieKeeper,omitempty"`
	UserID       *UserIDConfig        `json:"userId,omitempty"`
	BotDetection *BotDetectionConfig  `json:"botDetection,omitempty"`
	IPBlocklist  *IPBlocklistConfig   `json:"ipBlocklist,omitempty"`
}

type CookieKeeperConfig struct {
	Enabled bool `json:"enabled"`
	MaxAge  int  `json:"maxAge"` // seconds, default 34560000 (400 days)
}

type UserIDConfig struct {
	Enabled bool   `json:"enabled"`
	Salt    string `json:"salt"`             // random salt for hashing
	Header  string `json:"header,omitempty"` // default "X-Stape-User-Id"
}

type BotDetectionConfig struct {
	Enabled     bool `json:"enabled"`
	BlockBots   bool `json:"blockBots"`   // return 403 for bots
	HeaderName  string `json:"headerName,omitempty"` // default "X-Bot"
}

type IPBlocklistConfig struct {
	Enabled bool     `json:"enabled"`
	IPs     []string `json:"ips"` // IPs or CIDRs to block
}

func loadConfig() (*PowerupsConfig, error) {
	raw := os.Getenv("POWERUPS_CONFIG")
	if raw == "" {
		return &PowerupsConfig{}, nil
	}

	var cfg PowerupsConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("invalid POWERUPS_CONFIG: %w", err)
	}

	// Apply defaults
	if cfg.CookieKeeper != nil && cfg.CookieKeeper.MaxAge == 0 {
		cfg.CookieKeeper.MaxAge = 34560000 // 400 days
	}
	if cfg.UserID != nil && cfg.UserID.Header == "" {
		cfg.UserID.Header = "X-Stape-User-Id"
	}
	if cfg.BotDetection != nil && cfg.BotDetection.HeaderName == "" {
		cfg.BotDetection.HeaderName = "X-Bot"
	}

	return &cfg, nil
}
