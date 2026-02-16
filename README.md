# GTM Power-ups

Open-source server-side power-ups for Google Tag Manager by [Dublyo](https://dublyo.com).

A lightweight Go reverse proxy that sits in front of your sGTM (server-side Google Tag Manager) container and adds powerful enhancements — Cookie Keeper, User ID generation, Bot Detection, and IP Blocklist.

## Power-ups

### Cookie Keeper

Extends `Set-Cookie` response headers from sGTM by rewriting `Max-Age`. Defeats Safari ITP and short browser cookie lifetimes that kill attribution accuracy.

- Default Max-Age: 400 days (34,560,000 seconds)
- Configurable per deployment
- Preserves all original cookie attributes (Path, Domain, SameSite, Secure, HttpOnly)

### User ID

Generates a deterministic anonymous user identifier from `SHA256(IP + User-Agent + salt)` and injects it as a request header to sGTM. Useful for cookieless tracking and user deduplication.

- Header: `X-Stape-User-Id` (configurable)
- 16 hex characters (first 8 bytes of SHA-256)
- Supports Cloudflare (`CF-Connecting-IP`), Nginx/Traefik (`X-Real-IP`), and `X-Forwarded-For`

### Bot Detection

Matches User-Agent against 50+ known bot patterns including search engine crawlers, social media scrapers, SEO tools, monitoring services, and headless browsers. Adds an `X-Bot` header for sGTM to filter on.

- Header: `X-Bot: true` or `X-Bot: false`
- Optional: block bots entirely with a 403 response
- Patterns cover: Googlebot, Bingbot, Facebookbot, Semrush, Ahrefs, Puppeteer, Selenium, curl, wget, and many more

### IP Blocklist

Blocks requests from specific IP addresses or CIDR ranges. Blocked IPs receive a `403 Forbidden` response.

- Supports individual IPs (`10.0.0.1`) and CIDR ranges (`192.168.0.0/24`)
- IPs/CIDRs are parsed once at startup for fast runtime matching
- Uses `CF-Connecting-IP`, `X-Real-IP`, or `X-Forwarded-For` to detect real client IP

## Quick Start

### Docker

```bash
docker run -d \
  -e UPSTREAM_URL=http://sgtm:8080 \
  -e PORT=8081 \
  -e POWERUPS_CONFIG='{"cookieKeeper":{"enabled":true,"maxAge":34560000},"userId":{"enabled":true,"salt":"your-random-salt"},"botDetection":{"enabled":true,"blockBots":false},"ipBlocklist":{"enabled":false}}' \
  -p 8081:8081 \
  ghcr.io/dublyo/gtm-powerup:latest
```

### Docker Compose

```yaml
services:
  proxy:
    image: ghcr.io/dublyo/gtm-powerup:latest
    environment:
      - UPSTREAM_URL=http://sgtm:8080
      - PORT=8081
      - POWERUPS_CONFIG={"cookieKeeper":{"enabled":true,"maxAge":34560000},"userId":{"enabled":true,"salt":"change-me"},"botDetection":{"enabled":true,"blockBots":false},"ipBlocklist":{"enabled":false}}
    ports:
      - "8081:8081"
    depends_on:
      - sgtm

  sgtm:
    image: gcr.io/cloud-tagging-10302018/gtm-cloud-image:stable
    environment:
      - CONTAINER_CONFIG=your-container-config
      - PORT=8080
    expose:
      - "8080"
```

## Configuration

All configuration is done via the `POWERUPS_CONFIG` environment variable (JSON string).

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `UPSTREAM_URL` | No | `http://sgtm:8080` | URL of the upstream sGTM server |
| `PORT` | No | `8081` | Port the proxy listens on |
| `POWERUPS_CONFIG` | No | `{}` | JSON configuration for power-ups |

### POWERUPS_CONFIG Schema

```json
{
  "cookieKeeper": {
    "enabled": true,
    "maxAge": 34560000
  },
  "userId": {
    "enabled": true,
    "salt": "your-random-salt",
    "header": "X-Stape-User-Id"
  },
  "botDetection": {
    "enabled": true,
    "blockBots": false,
    "headerName": "X-Bot"
  },
  "ipBlocklist": {
    "enabled": true,
    "ips": ["10.0.0.0/8", "192.168.1.100"]
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `cookieKeeper.enabled` | bool | `false` | Enable Cookie Keeper |
| `cookieKeeper.maxAge` | int | `34560000` | Max-Age in seconds (400 days) |
| `userId.enabled` | bool | `false` | Enable User ID generation |
| `userId.salt` | string | `""` | Salt for SHA-256 hash (required when enabled) |
| `userId.header` | string | `X-Stape-User-Id` | Request header name |
| `botDetection.enabled` | bool | `false` | Enable bot detection |
| `botDetection.blockBots` | bool | `false` | Return 403 for detected bots |
| `botDetection.headerName` | string | `X-Bot` | Request header name |
| `ipBlocklist.enabled` | bool | `false` | Enable IP blocklist |
| `ipBlocklist.ips` | string[] | `[]` | IPs or CIDRs to block |

## Health Check

The proxy exposes a `/healthz` endpoint that returns `200 OK` with body `ok`. This is handled locally by the proxy without forwarding to the upstream.

## How It Works

The proxy builds a middleware chain at startup based on the enabled power-ups:

```
Incoming Request
    │
    ▼
IP Blocklist (if enabled) ── 403 if blocked
    │
    ▼
Bot Detection (if enabled) ── adds X-Bot header, optionally 403
    │
    ▼
User ID (if enabled) ── adds X-Stape-User-Id header
    │
    ▼
Reverse Proxy → sGTM (upstream)
    │
    ▼ (response)
Cookie Keeper (if enabled) ── rewrites Set-Cookie Max-Age
    │
    ▼
Response to client
```

## Building from Source

```bash
# Build binary
go build -o gtm-powerup .

# Build Docker image
docker build -t gtm-powerup .

# Run locally
UPSTREAM_URL=http://localhost:8080 PORT=8081 POWERUPS_CONFIG='{"cookieKeeper":{"enabled":true}}' ./gtm-powerup
```

## Image

Published to GitHub Container Registry on every push to `main`:

```
ghcr.io/dublyo/gtm-powerup:latest
```

Multi-platform: `linux/amd64` and `linux/arm64`.

Image size: ~15MB (multi-stage Alpine build, statically compiled Go binary).

## License

MIT
