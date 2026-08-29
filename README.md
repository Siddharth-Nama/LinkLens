# LinkLens

A small HTTPS API that takes a LinkedIn profile URL and gives you clean JSON back — name, headline, experience, education, skills, and more.

Built for the **Tross engineering challenge**. It talks to LinkedIn the same way the website does (Voyager HTTP APIs). No browser automation, no Playwright, no Puppeteer.

**Live API:** https://link-lens-5ejb.vercel.app

---

## Why this exists

LinkedIn profiles are useful data — for recruiting, research, integrations — but there is no simple public API to pull a full profile from a URL.

Most scrapers spin up a headless browser, log in, and pray. That is slow, brittle, and painful to host.

LinkLens does something simpler:

1. You pass a profile URL.
2. The server calls LinkedIn's internal Voyager API with a real logged-in session.
3. It maps LinkedIn's messy nested JSON into one flat, predictable response.

That matters because:

- **Fast** — one HTTP call per profile, not a full browser session.
- **Deployable** — runs as a normal Go server on Vercel, Render, Railway, or Docker.
- **Honest** — if a section is missing or hidden, the JSON tells you (`partial`, `missingSections`).

> **Heads up:** This uses your personal LinkedIn cookies. That is against LinkedIn's Terms of Service for production use. Fine for this assignment on an account you control. Sessions expire, rate limits happen, and LinkedIn can change endpoints anytime.

---

## Recommended: run locally with your own LinkedIn session

The hosted URL (Vercel) proves the assignment requirement — **public HTTPS API**. But LinkedIn cookies **die often** (hours to days). Every time they expire you must:

1. Log into LinkedIn in your browser again
2. Copy fresh `li_at` + `JSESSIONID`
3. Update env vars on Vercel
4. Redeploy

That is painful for you **and** for anyone reviewing the project.

**So for actually using the API or verifying it works, run it locally with your own session:**

```bash
git clone https://github.com/Siddharth-Nama/LinkLens.git
cd LinkLens
cp .env.example .env
# paste YOUR cookies into .env (see flow below)
make run
```

Then hit `http://localhost:8080` — updating `.env` takes 30 seconds. No redeploy, no Vercel dashboard.

The live deployment may return **502** if cookies on the server have expired. That does **not** mean the code is broken. Clone the repo, add your cookies, run locally — you will get **200**.

---

## How to get a 200 response (strict flow)

Follow these steps in order. Skip none.

### Step 1 — Fresh LinkedIn cookies (your browser)

1. Open Chrome and go to [linkedin.com](https://www.linkedin.com) — **must be logged in**.
2. DevTools (`F12`) → **Application** → **Cookies** → `https://www.linkedin.com`
3. Copy these two values:

| Browser cookie | Put in `.env` as | Example shape |
|----------------|------------------|---------------|
| `li_at` | `LI_AT` | long alphanumeric string |
| `JSESSIONID` | `LI_JSESSIONID` | `ajax:1739...` |

**Rules:**
- Paste **without** surrounding quotes → `ajax:123` not `"ajax:123"`
- Do not paste the whole `JSESSIONID=...` line — value only
- Use cookies from the **same** LinkedIn account you are logged in as

### Step 2 — Configure `.env`

```bash
cp .env.example .env
```

```env
LI_AT=<your li_at value>
LI_JSESSIONID=<your jsessionid value, no quotes>
API_KEY=any-secret-you-want          # optional locally
PORT=8080
```

### Step 3 — Start the server

```bash
make run
```

Wait for log: `linklens ready` / `listening`.

### Step 4 — Health check (must pass first)

```bash
curl http://localhost:8080/health
```

**Expected:**

```json
{"status":"ok","linkedin_configured":true}
```

If `linkedin_configured` is `false` → `LI_AT` is missing or empty in `.env`. Fix and restart.

### Step 5 — Profile request (the actual 200)

```bash
curl "http://localhost:8080/v1/profiles?url=https://www.linkedin.com/in/williamhgates/"
```

If you set `API_KEY` in `.env`, add the header:

```bash
curl -H "X-API-Key: your-secret" \
  "http://localhost:8080/v1/profiles?url=https://www.linkedin.com/in/williamhgates/"
```

**Expected:** HTTP **200** + JSON with `firstName`, `lastName`, `headline`, etc.

Test a second profile to confirm the session works for more than one slug:

```bash
curl -H "X-API-Key: your-secret" \
  "http://localhost:8080/v1/profiles?url=https://www.linkedin.com/in/siddharth-nama/"
```

### Step 6 — If you do not get 200

| HTTP | Error code | Fix |
|------|------------|-----|
| 400 | `INVALID_URL` | Use a full `https://www.linkedin.com/in/{slug}/` URL |
| 401 | `UNAUTHORIZED` | Add correct `X-API-Key` header (matches `API_KEY` in `.env`) |
| 404 | `NOT_FOUND` | Profile hidden or slug wrong |
| 502 | `SESSION_EXPIRED` | Cookies expired → repeat **Step 1** and restart server |
| 502 | `UPSTREAM_ERROR` | Usually expired cookies or bad `JSESSIONID` format → repeat **Step 1** |

**You do not need new cookies for each profile.** One good session fetches many profiles. Refresh only when **all** profiles start failing.

### Hosted API (Vercel) — same flow, extra steps

For `https://link-lens-5ejb.vercel.app` to return 200:

1. Complete Steps 1–5 locally first (proves your cookies work).
2. Copy the **same** `LI_AT`, `LI_JSESSIONID`, and `API_KEY` into [Vercel env vars](https://vercel.com/hi-ea86/link-lens-5ejb/settings/environment-variables).
3. **Redeploy** (env changes are not live until redeploy).
4. Test within an hour — cookies may die again later.

```bash
curl -H "X-API-Key: your-secret" \
  "https://link-lens-5ejb.vercel.app/v1/profiles?url=https://www.linkedin.com/in/williamhgates/"
```

If Vercel returns 502 but local returns 200 → server cookies are stale. Update Vercel env and redeploy.

---

## How it works

```mermaid
flowchart LR
    A[Client\ncurl / Postman] -->|GET /v1/profiles?url=...| B[LinkLens API\ncmd/server]
    B --> C{API key set?}
    C -->|invalid| D[401 Unauthorized]
    C -->|ok| E[profileurl\nvalidate /in/slug]
    E -->|bad URL| F[400 Invalid URL]
    E -->|ok| G{Cache hit?}
    G -->|yes| H[Return cached JSON]
    G -->|no| I[linkedin client\nVoyager API]
    I -->|cookies + CSRF| J[www.linkedin.com\n/voyager/api/identity/dash/profiles]
    J --> K[map.go\nincluded[] → Profile]
    K --> L[Cache + JSON response]
    L --> A
```

**Step by step:**

1. **Request comes in** — `GET /v1/profiles?url=https://www.linkedin.com/in/someone/`
2. **Auth check** — if `API_KEY` is set in env, the `X-API-Key` header must match.
3. **URL parsing** — only `/in/{slug}` profile links are accepted. Company pages, jobs, etc. are rejected.
4. **Cache lookup** — same slug within `CACHE_TTL` (default 15 min) skips LinkedIn entirely.
5. **LinkedIn fetch** — authenticated GET to Voyager with `li_at`, `JSESSIONID`, and `Csrf-Token` headers.
6. **Mapping** — walks LinkedIn's `included[]` array and picks out profile, positions, schools, skills, etc.
7. **Response** — structured JSON with `partial` / `missingSections` when data is incomplete.

You do **not** need fresh cookies for every new profile. One valid session can fetch many profiles. Refresh cookies only when the session dies — see [How to get a 200 response](#how-to-get-a-200-response-strict-flow).

---

## What's included

### API endpoints

| Endpoint | Auth | What it does |
|----------|------|--------------|
| `GET /health` | None | Server up? LinkedIn cookies configured? |
| `GET /v1/profiles?url=...` | `X-API-Key` if `API_KEY` is set | Full profile JSON for a LinkedIn URL |

### Profile fields returned

| Section | Fields |
|---------|--------|
| **Basics** | `firstName`, `lastName`, `fullName`, `headline`, `about`, `industry` |
| **Identity** | `publicIdentifier`, `profileId`, `inputUrl`, `canonicalUrl`, `fetchedAt` |
| **Location** | `country`, `city`, `raw` |
| **Media** | `profilePictures`, `backgroundImage` |
| **Social** | `connectionsCount`, `followersCount` |
| **Career** | `experience` (title, company, dates, description, current flag) |
| **Education** | `education` (school, degree, field, dates) |
| **Skills** | `skills` (name, endorsement count) |
| **More** | `certifications`, `languages`, `volunteer`, `projects`, `honors`, `publications` |
| **Quality** | `partial`, `missingSections` — tells you if something could not be loaded |

### Codebase

```
cmd/server/           HTTP server, routes, API key auth
internal/config/      env vars (cookies, timeouts, cache)
internal/profileurl/  URL validation (/in/slug only)
internal/linkedin/    Voyager client, fetch, JSON mapper
internal/profile/     public response types + error codes
internal/service/     orchestration (fetch → map → cache)
internal/cache/       in-memory TTL cache
```

### Also in the repo

- **Tests** — `httptest` fixtures, no live LinkedIn calls in CI
- **Dockerfile** — container deploy
- **Makefile** — `run`, `test`, `vet`, `fmt`, `docker-build`
- **vercel.json** — Go deploy on Vercel
- **`.env.example`** — template for local setup

### Tech stack

- **Go 1.23**
- LinkedIn **Voyager / Rest.li** API (`application/vnd.linkedin.normalized+json+2.1`)
- In-memory cache (no Redis needed for this scope)
- JSON structured logging (`slog`)

---

## Quick start

### Prerequisites

- Go 1.23+
- LinkedIn account (logged in, to copy cookies from browser)
- Docker (optional)

### Session cookies

1. Log in at [linkedin.com](https://www.linkedin.com).
2. DevTools → **Application** → **Cookies** → `https://www.linkedin.com`
3. Copy:
   - `li_at` → `LI_AT`
   - `JSESSIONID` → `LI_JSESSIONID`

**Important:** paste values **without** surrounding quotes. Use `ajax:1234567890`, not `"ajax:1234567890"`. Quotes break cookies on some hosts (Vercel logs will show `invalid byte '"' in Cookie.Value`).

Cookies last days to weeks. When **all** new profiles start failing, grab fresh ones — not before every request.

### Local setup

```bash
cp .env.example .env
# paste LI_AT, LI_JSESSIONID, and optionally API_KEY

make run
# or: go run ./cmd/server
```

**Health:**

```bash
curl http://localhost:8080/health
```

**Profile:**

```bash
curl "http://localhost:8080/v1/profiles?url=https://www.linkedin.com/in/williamhgates/"
```

**With API key:**

```bash
curl -H "X-API-Key: your-secret" \
  "http://localhost:8080/v1/profiles?url=https://www.linkedin.com/in/williamhgates/"
```

### Docker

```bash
make docker-build
docker run --rm -p 8080:8080 --env-file .env linklens
```

---

## API reference

### `GET /health`

```json
{"status":"ok","linkedin_configured":true}
```

### `GET /v1/profiles?url=<linkedin-profile-url>`

| Query | Required | Description |
|-------|----------|-------------|
| `url` | yes | Full LinkedIn profile URL (`https://www.linkedin.com/in/{slug}/`) |

| Header | When |
|--------|------|
| `X-API-Key` | Required only if `API_KEY` env var is set |

**Success (200)** — truncated example:

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
  "location": {"city": "Mumbai", "country": "India", "raw": "Mumbai, India"},
  "experience": [],
  "education": [],
  "skills": [],
  "partial": false,
  "missingSections": [],
  "fetchedAt": "2026-08-29T12:00:00Z"
}
```

**Errors:**

```json
{"error":{"code":"INVALID_URL","message":"profile url query param is required"}}
```

| HTTP | Code | When |
|------|------|------|
| 400 | `INVALID_URL` | Missing or bad profile URL |
| 401 | `UNAUTHORIZED` | Wrong or missing API key |
| 404 | `NOT_FOUND` | Profile not found or not visible to your session |
| 429 | `RATE_LIMITED` | LinkedIn rate limited your session |
| 502 | `SESSION_EXPIRED` | Cookies missing, expired, or auth redirect |
| 502 | `UPSTREAM_ERROR` | LinkedIn returned an unexpected error |

---

## Deploy (HTTPS)

The assignment needs a **public HTTPS** URL. The repo is deployed at:

**https://link-lens-5ejb.vercel.app**

> **Note for reviewers:** The hosted API uses the submitter's LinkedIn session cookies in server env vars. Those cookies expire regularly. If the live URL returns 502, please **clone and run locally** with your own `LI_AT` / `LI_JSESSIONID` — see [How to get a 200 response](#how-to-get-a-200-response-strict-flow). The code and tests are the real deliverable; the hosted URL shows deployability.

### Vercel (current)

1. Push repo to GitHub and connect on [vercel.com](https://vercel.com).
2. `vercel.json` already sets `"framework": "go"` — Vercel picks up `cmd/server/main.go`.
3. Add env vars in project settings:
   - `LI_AT`
   - `LI_JSESSIONID` (no quotes)
   - `API_KEY` (recommended)
4. Redeploy after changing env vars.

```bash
curl -H "X-API-Key: your-secret" \
  "https://link-lens-5ejb.vercel.app/v1/profiles?url=https://www.linkedin.com/in/williamhgates/"
```

### Railway / Render

Same env vars. Connect the GitHub repo, use the Dockerfile or `go run ./cmd/server`, and set `LI_AT`, `LI_JSESSIONID`, `API_KEY`. Both give you HTTPS automatically.

---

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP port |
| `LI_AT` | — | LinkedIn session cookie (**required**) |
| `LI_JSESSIONID` | — | CSRF cookie (**strongly recommended**) |
| `API_KEY` | — | If set, `X-API-Key` header required on `/v1/profiles` |
| `CACHE_TTL` | `15m` | Profile cache TTL (`0` = off) |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `HTTP_READ_TIMEOUT` | `15s` | Server read timeout |
| `HTTP_WRITE_TIMEOUT` | `15s` | Server write timeout |
| `LINKEDIN_TIMEOUT` | `20s` | LinkedIn request timeout |

Never commit `.env` or real cookie values to git.

---

## LinkedIn integration details

**Endpoint used:**

```
GET /voyager/api/identity/dash/profiles
  ?q=memberIdentity
  &memberIdentity={slug}
  &decorationId=com.linkedin.voyager.dash.deco.identity.profile.FullProfileWithEntities-93
```

**Headers sent:**

- `Accept: application/vnd.linkedin.normalized+json+2.1`
- `X-RestLi-Protocol-Version: 2.0.0`
- `Csrf-Token` (from `JSESSIONID`, quotes stripped)
- `Referer: https://www.linkedin.com/in/{slug}/`
- Cookies: `li_at`, `JSESSIONID`

The mapper reads LinkedIn's `included[]` graph and matches entities by `$type` (Profile, Position, Education, Skill, etc.).

---

## Known limitations

- **Not an official API** — runs on your personal LinkedIn session.
- **Sessions expire** — refresh `LI_AT` / `LI_JSESSIONID` when lookups fail; **local `.env` is much easier than updating Vercel every time**.
- **Hosted demo is fragile** — Vercel cookies go stale; prefer local run for hands-on testing.
- **Rate limits** — LinkedIn may return 429; cache helps but does not remove the limit.
- **Visibility** — you only see what your logged-in account can see.
- **Partial profiles** — private or sparse profiles may return `partial: true`.
- **Decoration IDs** — LinkedIn changes these sometimes; may need a one-line update in code.
- **One profile per request** — no batch endpoint.
- **Ephemeral cache on serverless** — Vercel cache lives in memory per instance; cold starts start fresh.

---

## Development

```bash
make test    # all tests
make vet     # go vet
make fmt     # gofmt
```

Tests use JSON fixtures under `internal/linkedin/testdata/` — CI does not hit LinkedIn.

---

## License

MIT — assignment submission.
