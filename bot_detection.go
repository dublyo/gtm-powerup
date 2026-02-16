package main

import (
	"net/http"
	"regexp"
	"strings"
)

// Known bot User-Agent patterns. Covers major crawlers, monitoring tools,
// SEO bots, social media scrapers, and common headless browsers.
var botPatterns = []string{
	// Search engines
	`googlebot`, `bingbot`, `yandexbot`, `baiduspider`, `duckduckbot`,
	`slurp`, `sogou`, `exabot`, `ia_archiver`, `archive\.org_bot`,
	// Social / messengers
	`facebookexternalhit`, `facebot`, `twitterbot`, `linkedinbot`,
	`pinterest`, `slackbot`, `telegrambot`, `whatsapp`, `discordbot`,
	// SEO / monitoring
	`semrushbot`, `ahrefsbot`, `mj12bot`, `dotbot`, `rogerbot`,
	`screaming frog`, `deepcrawl`, `sitebulb`,
	// Uptime / monitoring
	`uptimerobot`, `pingdom`, `statuscake`, `newrelicpinger`,
	`site24x7`, `datadog`,
	// Headless browsers / tools
	`headlesschrome`, `phantomjs`, `puppeteer`, `playwright`,
	`selenium`, `webdriver`,
	// Generic patterns
	`bot`, `crawler`, `spider`, `scraper`, `curl`, `wget`, `python-requests`,
	`go-http-client`, `java/`, `libwww`, `httpunit`, `nutch`,
	`biglotron`, `teoma`, `convera`, `gigablast`, `ia_archiver`,
}

var botRegex *regexp.Regexp

func init() {
	pattern := strings.Join(botPatterns, "|")
	botRegex = regexp.MustCompile(`(?i)(` + pattern + `)`)
}

// botDetectionMiddleware checks User-Agent against known bot patterns.
// If a bot is detected, it adds an X-Bot header. If blockBots is true,
// it returns 403 Forbidden for bot traffic.
func botDetectionMiddleware(cfg *BotDetectionConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.UserAgent()
		isBot := botRegex.MatchString(ua)

		if isBot {
			r.Header.Set(cfg.HeaderName, "true")

			if cfg.BlockBots {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Forbidden"))
				return
			}
		} else {
			r.Header.Set(cfg.HeaderName, "false")
		}

		next.ServeHTTP(w, r)
	})
}
