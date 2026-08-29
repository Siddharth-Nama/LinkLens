# LinkLens

Hosted HTTPS API that accepts a LinkedIn profile URL and returns structured JSON for the data visible on that profile.

Built for the Tross engineering challenge. Uses LinkedIn's authenticated **Voyager HTTP APIs** directly — no browser automation (Playwright/Puppeteer).

> **Disclaimer:** Calling LinkedIn's internal APIs with your personal session violates LinkedIn's Terms of Service. Use only with an account you control, for this assignment. Sessions expire, rate limits apply, and endpoints can change without notice.

## Quick start

### Prerequisites

- Go 1.23+
- A LinkedIn account (logged in via browser to copy cookies)
- Docker (optional, for container deploy)

### 1. Get LinkedIn session cookies

1. Log in to [linkedin.com](https://www.linkedin.com) in Chrome/Firefox.
2. Open DevTools → **Application** → **Cookies** → `https://www.linkedin.com`
3. Copy:
   - `li_at` → `LI_AT`
   - `JSESSIONID` → `LI_JSESSIONID` (quotes are fine; server strips them)

Cookies expire after some days/weeks. Refresh them when you get `SESSION_EXPIRED` errors.

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env and paste your cookies
```

Never commit `.env` or real cookies to git.

### 3. Run locally

```bash
make run
# or
go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8080/health
```

Profile lookup:

```bash
curl "http://localhost:8080/v1/profiles?url=https://www.linkedin.com/in/williamhgates/"
```

With API key (if `API_KEY` is set):

```bash
curl -H "X-API-Key: your-secret" \
  "http://localhost:8080/v1/profiles?url=https://www.linkedin.com/in/williamhgates/"
```

### 4. Docker

```bash
make docker-build
docker run --rm -p 8080:8080 --env-file .env linklens
```

## API

### `GET /health`

No auth required.

```json
{"status":"ok","linkedin_configured":true}
```

### `GET /v1/profiles?url=<linkedin-profile-url>`

Returns structured profile JSON.

| Query param | Required | Description |
|-------------|----------|-------------|
| `url` | yes | Full LinkedIn profile URL (`https://www.linkedin.com/in/{slug}/`) |

| Header | Required | Description |
|--------|----------|-------------|
| `X-API-Key` | only if `API_KEY` env is set | Protects the public endpoint |

**Success (200)** — example (truncated):

```json
{
  "inputUrl": "https://www.linkedin.com/in/jane-doe/",
  "canonicalUrl": "https://www.linkedin.com/in/jane-doe/",
  "publicIdentifier": "jane-doe",
  "firstName": "Jane",
  "lastName": "Doe",
  "fullName": "Jane Doe",
  "headline": "Software Engineer",
  "about": "...",
  "experience": [],
  "education": [],
  "skills": [],
  "partial": false,
  "missingSections": []
}
```

**Errors** — consistent shape:

```json
{"error":{"code":"INVALID_URL","message":"profile url query param is required"}}
```

| HTTP | Code | When |
|------|------|------|
| 400 | `INVALID_URL` | Missing or invalid profile URL |
| 401 | `UNAUTHORIZED` | Wrong/missing API key |
| 404 | `NOT_FOUND` | Profile not found or not visible to your session |
| 429 | `RATE_LIMITED` | LinkedIn rate limited the session |
| 502 | `SESSION_EXPIRED` | Missing/expired `LI_AT` or LinkedIn auth failure |
| 502 | `UPSTREAM_ERROR` | LinkedIn returned an unexpected error |

## Approach

### Why no browser?

Tross clarified the assignment expects **reverse-engineered HTTP calls** to LinkedIn endpoints, not headless Chrome. Browsers are slower, harder to deploy, and easier for LinkedIn to detect.

### How it works

```
Client → LinkLens API → LinkedIn Voyager API → JSON mapper → Response
```

1. **URL parsing** (`internal/profileurl`) — validates `/in/{slug}` URLs, rejects `/company/` etc.
2. **LinkedIn client** (`internal/linkedin`) — sends authenticated GET to:
   ```
   /voyager/api/identity/dash/profiles
     ?q=memberIdentity
     &memberIdentity={slug}
     &decorationId=com.linkedin.voyager.dash.deco.identity.profile.FullProfileWithEntities-93
   ```
   With headers: `li_at` + `JSESSIONID` cookies, `Csrf-Token`, `X-RestLi-Protocol-Version: 2.0.0`, and LinkedIn's normalized JSON accept type.
3. **Mapper** (`internal/linkedin/map.go`) — walks the `included[]` graph and maps `$type` entities (Position, Education, Skill, etc.) into a flat JSON schema.
4. **Cache** — in-memory TTL cache (default 15m) keyed by profile slug to reduce repeat LinkedIn calls.

### Project layout

```
cmd/server/          HTTP server, handlers, auth
internal/config/     env-based configuration
internal/profileurl/ URL validation
internal/linkedin/   Voyager client + mapper
internal/profile/    public JSON types
internal/service/    business logic
internal/cache/      TTL cache
```

## Deploy (HTTPS)

The assignment requires a **public HTTPS** endpoint. Easiest options:

### Railway

1. Push repo to GitHub.
2. [railway.app](https://railway.app) → New Project → Deploy from GitHub.
3. Set env vars: `LI_AT`, `LI_JSESSIONID`, `API_KEY` (recommended).
4. Railway assigns `https://your-app.up.railway.app`.

### Render

1. [render.com](https://render.com) → New Web Service → connect repo.
2. Build: `docker build -t linklens .` or use Dockerfile auto-detect.
3. Set env vars in dashboard.
4. Render provides HTTPS URL automatically.

After deploy, test:

```bash
curl "https://YOUR_HOST/v1/profiles?url=https://www.linkedin.com/in/williamhgates/"
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `LI_AT` | — | LinkedIn session cookie (required) |
| `LI_JSESSIONID` | — | CSRF/session cookie (recommended) |
| `API_KEY` | — | If set, requires `X-API-Key` header |
| `CACHE_TTL` | `15m` | Profile cache TTL (`0` disables) |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `HTTP_READ_TIMEOUT` | `15s` | Server read timeout |
| `HTTP_WRITE_TIMEOUT` | `15s` | Server write timeout |
| `LINKEDIN_TIMEOUT` | `20s` | Upstream LinkedIn timeout |

## Known limitations

- **Session-based auth** — uses your personal LinkedIn cookies; not an official API.
- **Session expiry** — `li_at` dies; you must refresh cookies manually.
- **Rate limits** — LinkedIn may return 429; cache helps but doesn't eliminate this.
- **Decoration drift** — `decorationId` values change when LinkedIn deploys; may need updating.
- **Visibility** — you only see what your logged-in account can see (1st/2nd/3rd degree, etc.).
- **Partial data** — some profiles hide sections; response includes `partial` and `missingSections`.
- **No batch API** — one profile per request.
- **ToS** — not suitable for production scraping.

## Development

```bash
make test    # run all tests
make vet     # go vet
make fmt     # gofmt
```

Tests use `httptest` fixtures — no live LinkedIn calls in CI.

## License

MIT (or your choice) — assignment submission.
