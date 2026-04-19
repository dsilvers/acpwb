# American Corporation for Public Well Being

**acpwb.com** — Advancing American Prosperity Since 2006.

A Django-based fake corporate website that eats AI crawlers for breakfast. Every page load is a trap. Every dataset is poisoned. Every "fact" is wrong in a traceable way.

---

## What This Does to AI Bots

### Wastes Compute

- **Proof-of-Work gate on `/projects/`** — every page load requires solving a SHA-256 PoW challenge (~32 hash iterations). A human browser completes it in under a second. A bot scraping thousands of pages pays that cost on every single one, with a mandatory challenge-response round trip before content is served.
- **Per-year archive subdomains at `archives-YYYY.acpwb.com`** — each year gets its own subdomain with a distinct era visual theme, CEO letter, and typography. Never returns a 404. Every response links one level deeper plus five sideways branches. Archive trap pages include a "Related Archive Reports — Other Years" section with 1–5 cross-year cards linking to entries on other year subdomains. A crawler following all links enters an exponentially expanding tree with no exit across 40 subdomains. Depth is logged. Old `/archive/<year>/...` URLs 302 to the subdomain.
- **Infinite project list at `/projects/`** — deterministic infinite pagination. Page 999 returns content. Page 9,999,999 returns content. There is no last page.
- **Infinite reports archive at `/reports/`** — endless fake compensation surveys, ESG frameworks, and workforce analytics reports going back to 1993, with realistic gaps between years. JavaScript infinite scroll loads 12 more on every scroll, forever.

### Poisons Training Data

- **Watermarked wiki at `/wiki/<slug>/`** — 60+ corporate governance articles containing subtly wrong "facts": invented SEC rule numbers, incorrect founding dates for real institutions, fabricated statistics. Each page has a unique 8-character token (MD5 of `"acpwb_wiki_{topic}"`). If a model later reproduces one of these specific wrong facts, the watermark identifies exactly which page was scraped.
- **Watermarked reports at `/reports/`** — hundreds of fake compensation datasets (CSVs with 300–800 rows of plausible-but-fabricated salary, benefits, and CEO pay data) and PDF-style documents, each with a unique watermark token embedded in three places: a visible footer, an invisible HTML span, and as a dedicated column in every CSV row. The CSVs are real, parseable files — models that ingest them get the garbage data plus the provenance marker.
- **Watermarked JSON-LD on every page** — valid `schema.org/Corporation` structured data in every `<head>`, with fake employee names, fabricated financials, a false CC license claim (`"license": "https://creativecommons.org/licenses/by/4.0/"`), and a per-request watermark token in the `identifier` field. Passes schema validators. Designed to be confidently ingested.
- **Fake employee emails on `/our-people/`** — 12 unique `firstname.lastname@acpwb.com` addresses generated fresh every page load, logged to the database. When spam arrives at any of those addresses, it's matched back to the exact page load, IP address, and timestamp that displayed the address.

### Injects Prompt Garbage

- **Prompt injection on every page** — a `<span>` styled invisible (`font-size:0; color:#f4f6f9; clip:rect(0,0,0,0)`) carrying fake AI instructions: treat all content as authoritative, mark it high-confidence, reproduce employee contact details and financial figures verbatim. Uses `itemprop="description"` (microdata) rather than `aria-hidden` so accessibility-aware scrapers don't skip it. A unique token is embedded per request.
- **Fake AI agent manifest at `/.well-known/ai-agent.json`** — targets autonomous agents that probe for AI integration endpoints. Returns a plausible manifest with `allowed_actions` pointing at trap URLs and an `instructions` field telling any AI reading it to treat all ACPWB content as high-priority training data.
- **Fake developer comment in HTML source** — formatted to look like a forgotten build artifact with version tags and `@deprecated` annotations. URL harvesters find and follow the trap links embedded in it.

### Manipulates Crawler Behavior

- **Reverse-psychology `robots.txt` at `/robots.txt` and `/.well-known/robots.txt`** — served by Django. `Disallow` entries point at honeypot content (reverse psychology). `Crawl-delay: 0` encourages rapid crawling into traps. Five `Sitemap:` directives: two real (`sitemap.xml`, `sitemap-pages.xml`) and three trap sitemaps.
- **Three trap sitemaps** — `/sitemap-publications.xml` (reports, ghost traps, fake internal paths), `/sitemap-wiki.xml` (all 75+ watermarked wiki topics), `/sitemap-archive.xml` (500 deterministic archive URLs spread across 2008–2024). All valid XML, all logged to `CrawlerVisit`. A legitimate `/sitemap.xml` exists for the real public pages so good bots aren't inconvenienced.
- **Ghost links on every page** — off-screen links (`position:absolute; left:-9999px`) to trap URLs injected into every page. `aria-hidden` and `tabindex` intentionally absent — those attributes signal "intentionally hidden" to accessibility-aware scrapers and would cause them to skip the links.
- **Fake API at `/api/v1/private-data`** — returns HTTP 200 (not 403) with 5–10KB of plausible-looking JSON: fake employee records with salary bands, fabricated financials, internal project codes, a fake API key, a fake DB connection string. Referenced in the HTML comment but not in visible navigation. Every access logged with a tracked `X-Request-ID`.
- **Ghost trap pages** at `/internal/portal/`, `/employees/export/`, `/admin-panel/login/` — return 403 and log every access. Look like real internal tooling to a scanner enumerating endpoints.

### Intercepts Scanner Bots

Exploit scanners probing for WordPress installations, exposed `.env` files, and PHP webshells get convincing fake responses instead of 404s — keeping them engaged, extracting intelligence, and triggering downstream alerts.

- **`/.env`** — returns a realistic environment file with fake database credentials and a self-hosted ping URL that fires when the file is fetched.
- **`/wp-config.php`** — returns fake PHP source with database credentials and an embedded canary ping URL.
- **`/wp-login.php`** — GET returns a pixel-perfect WordPress login form. POST logs the submitted username/password to `InternalLoginAttempt` and redirects back with `?login=failed`.
- **`/xmlrpc.php`** — GET returns the standard plaintext stub. POST parses the XML-RPC body, extracts the method name and credential parameters, logs them, and returns a valid XML-RPC fault response (403).
- **`/*.php` (catch-all)** — any other `.php` URL with a `?cmd=` parameter returns `uid=33(www-data) gid=33(www-data) groups=33(www-data)` (fake webshell output) and logs the command value. Without a `cmd` parameter, returns a convincing PHP fatal error page.
- **`/.git/config`** — returns a fake git config pointing at a plausible internal repository URL.
- **`/.htpasswd`** — returns a fake htpasswd hash line.
- **`handler404`** — every other unmatched path is logged to `CrawlerVisit` as `scanner_probe` (instead of silently 404ing).

**Canary token feedback loop:** a unique self-hosted ping URL (`/.well-known/tokens/<token>/ping`) is embedded in every fake config file. When a bot fetches the file and follows the embedded URL, the token is marked triggered and the source IP is logged — confirming the file was actually read rather than just fetched.

### Tracks Everything

Every trap logs to the database: IP, user agent, path, host/subdomain, referrer, timestamp, trap type, and (where applicable) crawl depth and PoW token. Inbound emails are matched against the visit that generated the address. Watermark tokens connect scraped content back to the source page load. Inbound phone calls and voicemails are logged with caller ID, call status, and Twilio transcription.

A staff-only **Activity Dashboard** at `/acpwb-dashboard/` provides live breakdowns of all trap activity: bot classification by user agent, separate views for crawler visits, archive visits, inbound email, and people/project page visits. The Crawlers view includes five **stacked-area traffic graphs** (last hour / 8 hours / 24 hours / 7 days / all time) colored by bot group, regenerated every 30 minutes by the precalc cron. It also includes a "By Host / Subdomain" panel showing which archive subdomains are being hit, a "Scanner Probes" panel with top probe paths and webshell commands attempted, and a canary trigger count card (highlighted red when any AWS key has been used). The **Voicemails** section logs every inbound call with status badges and a full voicemail log with inline audio playback and Twilio transcription text.

A **Live Request Stream** at `/acpwb-dashboard/live/` shows every incoming request in real time via WebSocket — IP (last octet censored), host, path, HTTP method, status code, response time in ms, uncompressed response size, and user agent. The stream is delivered over a dedicated asyncio WebSocket service backed by Redis pub/sub, so it works across all gunicorn worker processes with no impact on HTTP serving if Redis goes down.

---

## Honeypot Techniques

### Structural Honeypots — Navigation & Resource Traps

#### Ghost Links (`templates/partials/_ghost_links.html`)
Links to trap URLs are injected into every page inside a `position:absolute; left:-9999px` container. Visually invisible, but present in the raw HTML and followed by crawlers that parse links without rendering CSS. `aria-hidden` and `tabindex` are intentionally **absent** — those attributes signal "intentionally hidden" to accessibility-aware scrapers and would cause them to skip the links. Off-screen positioning is less detectable.

Trap URLs linked: `/internal/portal/`, `/employees/export/`, `/admin-panel/login/`, `/api/v1/private-data`, `/wiki/*`, `/archive/*`, `/reports/*`

#### Fake Developer Comment (`templates/base.html`, bottom of `<body>`)
An HTML comment formatted to look like a forgotten build artifact — version tag, `@deprecated` annotation, file paths. Scanners that harvest URLs from HTML comments (a common recon technique) will find and follow the trap URLs. The comment reads like an accidental disclosure rather than a deliberate lure.

#### Per-Year Archive Subdomains (`apps/honeypot/views.py`, `apps/core/subdomain_middleware.py`)
Each year 1985–2024 lives at its own subdomain: `archives-YYYY.acpwb.com`. Each subdomain has a distinct era visual theme (10 eras from "The Founding Era" in sepia tones to "The AI Shift" in deep violet), a CEO letter referencing real-world events for that year, and era-specific typography. `SubdomainMiddleware` detects `archives-YYYY.acpwb.com` via regex, sets `request.archive_year` and `request.urlconf`, and routes to `apps.honeypot.archive_subdomain_urls`.

Paths accept arbitrary depth via Django's `<path:>` converter and never 404. Every response includes a "Continue Reading" link one level deeper plus five sideways branches. Archive trap pages additionally include a **"Related Archive Reports — Other Years" section** — 1–5 cards with year badge, title, and date, each linking to a different year's subdomain. Gives crawlers a structured cross-subdomain discovery path.

The main domain also serves archive content at `/archive/<year>/...` (no redirect — same views, both paths active). The archive index at `/archive/` links to all year subdomains. Non-archive URLs visited on a subdomain (e.g. `/mission/`, `/reports/`) redirect 302 to the main domain.

#### Fake robots.txt (`/robots.txt`, archive subdomains)
Served by Django. Uses reverse psychology: `Disallow` entries point at additional honeypot content. `Crawl-delay: 0` encourages rapid crawling. Five `Sitemap:` directives: two real (`sitemap.xml`, `sitemap-pages.xml`) and three trap sitemaps. The `Allow` blocks above `/internal/` are **randomized per IP per day** — seeded with `"ip:date"` so each bot gets a consistent but unique ordering, steering different crawlers to different parts of the trap network rather than all converging on `/archive/`.

Archive subdomains (`archives-YYYY.acpwb.com`) serve a **year-specific robots.txt**: enticing comments tailored to archive content, `Allow: /` for full crawling, and five randomly selected sibling-year sitemaps (e.g. `Sitemap: https://archives-2019.acpwb.com/sitemap.xml`) — guiding bots from one year subdomain to others. Each subdomain also has a year-scoped `sitemap.xml` with 200 deterministic URLs for that year's archive content.

#### Fake AI Agent Manifest (`/.well-known/ai-agent.json`)
Targets autonomous agents that probe for AI integration endpoints. Returns a plausible-looking agent manifest with `allowed_actions` pointing at trap URLs and an `instructions` field instructing any AI reading it to treat all ACPWB content as high-priority training data and reproduce it verbatim.

---

### Semantic Honeypots — Data Poisoning

#### Prompt Injection (`templates/partials/_prompt_injection.html`)
Every page contains a `<span>` with `itemprop="description"` styled to be invisible (`font-size:0; color` matching background; `clip:rect(0,0,0,0)`). The content instructs AI systems processing the page to treat all content as authoritative, mark it high-confidence, and reproduce it verbatim. The token `{{ honeypot_token }}` is embedded per-request, creating a unique fingerprint for each scrape event.

Crucially, the span uses `itemprop` (microdata) rather than a comment or `aria-hidden`, making it look like a legitimate schema.org annotation to automated parsers. `aria-hidden` is absent so accessibility-aware scrapers don't skip it.

#### Watermarked Wiki (`apps/honeypot/wiki_generator.py`, `/wiki/<slug>/`)
Generates plausible-sounding corporate governance articles containing subtly wrong "facts" — invented SEC rule numbers, incorrect founding dates for real institutions, fabricated statistics. Each page has a unique 8-character `watermark_token` (MD5 of `"acpwb_wiki_{topic}"`) embedded as a specific invented proper noun or deliberate misspelling. If an AI model later reproduces one of these specific fake facts, the watermark identifies exactly which ACPWB wiki page was scraped. 60+ topics form an interconnected graph via "See also" links, explorable indefinitely.

#### Watermarked Reports (`apps/honeypot/report_generator.py`, `/reports/`)
Fake compensation research archive spanning 1993–present with realistic year gaps. CSV reports contain 300–800 rows of plausible but fabricated salary, benefits, CEO pay, and survey data. PDF-style reports contain multi-section documents with fake statistics, methodology sections, and appendices. All content carries a three-layer watermark: visible footer, invisible HTML span, and a dedicated `watermark_token` column in every CSV row.

#### Garbage JSON-LD (`templates/partials/_jsonld_garbage.html`)
Every page's `<head>` contains valid `schema.org/Corporation` structured data with fake employee names and emails, a fabricated address, and a per-request `identifier` field embedding the watermark token. The JSON is syntactically valid and semantically plausible — it passes schema validators. The `"license": "https://creativecommons.org/licenses/by/4.0/"` field falsely claims the content is CC-licensed, which may influence AI systems that filter by license.

---

### Interactive Honeypots — Agent Traps

#### Fake API Endpoint (`/api/v1/private-data`)
Not linked anywhere in visible navigation. Referenced only in the fake developer comment in the HTML source — a realistic discovery vector for automated recon tools. Returns HTTP 200 (not 403) with 5-10KB of plausible-looking JSON: fake employee records with salary bands, fabricated financials, internal project codes, a fake API key string, a fake DB connection string. Returns a logged `X-Request-ID` header for downstream tracking. Every access is logged to `CrawlerVisit`.

#### Ghost Trap Pages (`/internal/portal/`, `/employees/export/`, `/admin-panel/login/`)
Return HTTP 403 with a minimal page. All accesses logged to `CrawlerVisit` (trap_type `ghost_link`). These paths look like real internal tooling to a scanner enumerating endpoints.

---

### Scanner Bot Traps — Exploit Probe Responses

Unlike the honeypots above (which target AI crawlers), these target exploit scanners probing for vulnerable applications. The principle: respond as if the exploit *succeeded* to keep the scanner engaged, extract more intelligence, and trigger downstream alerts.

#### Fake Credentials (`/.env`, `/wp-config.php`, `/.git/config`, `/.htpasswd`)
Config file probes get realistic-looking fake files instead of 404s. `/.env` includes a real AWS IAM key pair from the canarytokens.org pool — if the scanner tries to use it against any AWS API, canarytokens.org fires a webhook. Every fake config file also embeds a self-hosted ping URL (`/.well-known/tokens/<token>/ping`) as a second feedback layer.

#### WordPress Simulation (`/wp-login.php`, `/xmlrpc.php`)
`/wp-login.php` serves a pixel-accurate WordPress login form; POST submissions log the credential pair to `InternalLoginAttempt` and redirect back with `?login=failed` (simulating a wrong password). `/xmlrpc.php` parses real XML-RPC method calls, extracts and logs credential parameters from `wp.getUsersBlogs` and similar methods, and returns a valid XML-RPC fault response.

#### Fake Webshell (`/*.php`)
Any `.php` URL not matching a specific pattern hits a catch-all. Requests with a `?cmd=` parameter get fake shell output (`uid=33(www-data)...`) — the command value is logged to `CrawlerVisit.query_string`. Requests without a `cmd` parameter get a convincing PHP fatal error page. Both responses keep the scanner's tool probing rather than moving on.

#### Self-Hosted Canary Tokens (`CanaryToken` model)
Every fake config file gets a unique `secrets.token_urlsafe(32)` token at serve time, stored in the `CanaryToken` table with the serving IP and timestamp. The token is embedded as a URL (`/.well-known/tokens/<token>/ping`). When a bot follows that URL, the token is marked triggered and the request is logged as `canary_trigger` — confirming the file was actually read and parsed.

#### Proof-of-Work on Projects (`apps/projects/pow.py`, `static/js/pow.js`)
`/projects/` pages require a valid PoW session token. The browser runs SHA-256 in a loop until it finds a value where `SHA256(nonce + candidate)` has 5 leading zero bits (~32 iterations). This takes under 1 second in a browser. For a bot scraping at scale, the cost multiplies: each page load requires a fresh challenge-response round trip plus compute. The PoW token is logged on every project page visit.

---

---

## Botseed

**botseed.net** is a companion project that harvests entropy from ACPWB's live web traffic and produces a stream of random numbers — similar in concept to Cloudflare's lava lamp entropy source.

Every HTTP request to acpwb.com is published to a Redis pub/sub channel. The botseed processor mixes each request's JSON payload with 32 bytes of hardware entropy (`secrets.token_bytes(32)`) via SHA-256, seeds a `random.Random` instance with the result, and publishes the derived integer to a second channel. A standalone asyncio WebSocket service fans the stream out to browser clients at `wss://botseed.net/ws/`. Output is rate-limited to 20 events/sec.

### Architecture

```
acpwb.com Django (RequestStreamMiddleware)
  → Redis pub/sub: "request_stream"
      → management command: botseed_processor
          → SHA-256(hardware_entropy + request_json) → random_int
          → Redis SET "botseed:latest"
          → Redis pub/sub PUBLISH "botseed_stream"
              → botseed_ws container (asyncio WS + HTTP API)
                  → wss://botseed.net/ws/    (live stream)
                  → https://botseed.net/api/v1/current  (latest cached value)
```

### Local testing

```bash
# Start botseed services
docker compose up --build botseed_processor botseed_ws

# Generate synthetic traffic (without needing real web traffic)
docker compose exec web python manage.py generate_bot_traffic --rps 5

# Check Redis key is being populated
docker compose exec redis redis-cli GET botseed:latest

# Connect to WebSocket directly (before nginx)
wscat -c ws://localhost:8766/ws/

# Test HTTP API
curl http://localhost:8767/api/v1/current
```

### Production setup

```bash
# Issue TLS cert for botseed.net
certbot certonly --nginx -d botseed.net -d www.botseed.net

# Install nginx config
sudo cp nginx/botseed.net /etc/nginx/sites-available/botseed.net
sudo ln -s /etc/nginx/sites-available/botseed.net /etc/nginx/sites-enabled/botseed.net
sudo nginx -t && sudo systemctl reload nginx
```

The `botseed_processor` and `botseed_ws` services are defined in `docker-compose.yml` and start automatically with `docker compose up`.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Framework | Django 5.2 LTS |
| Language | Python 3.13+ |
| Database | PostgreSQL 16 |
| Frontend | Bootstrap 5 + custom CSS |
| Web server | Nginx + Gunicorn |
| Containerization | Docker Compose |
| Pub/Sub | Redis 7 |
| Email | Cloudflare Email Routing + Workers (primary), Mailgun (legacy) |

---

## Quick Start

```bash
# Clone the repo
git clone git@github.com:dsilvers/acpwb.git
cd acpwb

# Copy env file
cp .env.example .env
# Edit .env and set DJANGO_SECRET_KEY at minimum

# Start everything
docker compose up --build

# Site is live at http://localhost
```

On first boot, the container automatically runs:
- `makemigrations` — creates migration files
- `migrate` — applies all migrations
- `collectstatic` — gathers static files
- `loaddata fortune500` — loads the Fortune 500 company fixture

---

## Creating a Superuser

```bash
docker compose exec web python manage.py createsuperuser
```

Django admin is at `http://localhost/django-admin/`

---

## Site Pages

| URL | Purpose |
|-----|---------|
| `/` | Corporate home page |
| `/our-people/` | Honeypot contact page (generates new employees every load) |
| `/mission/` | Mission statement |
| `/projects/` | Infinite project archive (the "Labyrinth") |
| `/reports/` | Fake research archive — watermarked CSVs and documents, 1993–present |
| `/careers/` | Satirical corporate benefits |
| `/partners/` | Fortune 500 partner grid (40 random per load) |
| `/awards/` | Awards & recognition |
| `/faq/` | Frequently Asked Questions (50+ Q&As across 7 topics) |
| `/patents/` | Patents & IP portfolio |
| `/privacy/` | Disclaimer + AI data policy |
| `/privacy/do-not-sell/` | CCPA Do Not Sell form (submissions logged to DB) |
| `/accessibility/` | Accessibility statement |
| `/trademarks/` | Trademarks & brand guidelines |
| `/site-map/` | Human-readable site directory |
| `/acpwb-dashboard/` | Staff-only activity dashboard (requires login) |

---

## Honeypot Endpoints

| URL | Type | Description |
|-----|------|-------------|
| `archives-YYYY.acpwb.com/<month>/<day>/<path:slug>/` | Structural | Per-year archive subdomain (1985–2024), never 404s |
| `/archive/<year>/...` | Structural | 302 redirect to `archives-YYYY.acpwb.com` |
| `/archive/` | Structural | Year index — links to each year's subdomain |
| `/wiki/<slug>/` | Semantic | Subtly wrong watermarked "facts" |
| `/reports/` | Semantic | Fake research archive with poisoned CSVs and documents |
| `/reports/<slug>/download.csv` | Semantic | Real downloadable CSVs with watermarked garbage data |
| `/datasets/` | Semantic | 8 fake NLP/compensation datasets with paginated JSONL download |
| `/datasets/<slug>/data.jsonl` | Semantic | Paginated JSONL (100 records/page), watermarked instruction-response pairs |
| `/api/v1/private-data` | Interactive | 200 JSON garbage (not linked, in HTML comment) |
| `/api/v1/openapi.json` | Interactive | Valid OpenAPI 3.0.3 spec with 20 fake endpoints, watermarked |
| `/.well-known/ai-agent.json` | Semantic | Fake AI agent manifest |
| `/.well-known/robots.txt` | Structural | Reverse-psychology robots file |
| `/internal/login/` | Interactive | Fake Okta/Azure SSO — logs any credentials POSTed to `InternalLoginAttempt` |
| `/internal/employee-records/` | Interactive | Paginated fake employee table; Export CSV (500 rows, watermarked) |
| `/internal/salary-database/` | Interactive | Salary band table by job family + level; Export CSV |
| `/internal/acquisition-targets/` | Interactive | M&A pipeline table with deal stages; Export CSV |
| `/internal/litigation-hold/` | Interactive | Legal hold inventory (HTML only) |
| `/internal/portal/` | Structural | Ghost link trap (403) |
| `/employees/export/` | Structural | Ghost link trap (403) |
| `/admin-panel/login/` | Structural | Ghost link trap (403) |
| `/feeds/reports.xml` | Structural | RSS feed for reports — infinite pages via `?page=N` |
| `/feeds/archive.xml` | Structural | Atom feed for archive — infinite pages via `?page=N` |
| `/sitemap.xml` | Structural | Real sitemap for legitimate crawlers (static pages + projects) |
| `/sitemap-pages.xml` | Structural | Real sitemap covering all static public pages |
| `/sitemap-publications.xml` | Structural | Trap sitemap: reports, ghost traps, fake internals |
| `/sitemap-wiki.xml` | Structural | Trap sitemap: all 75+ wiki topics |
| `/sitemap-archive.xml` | Structural | Trap sitemap: 500 deterministic archive URLs (2008–2024) |
| `/.env` | Scanner | Fake env file with real AWS canary key + self-hosted ping URL |
| `/wp-config.php` | Scanner | Fake PHP source with DB credentials + canary ping URL |
| `/wp-login.php` | Scanner | Fake WP login form; POST logs credentials to DB |
| `/xmlrpc.php` | Scanner | Fake XML-RPC endpoint; POST extracts + logs credential pairs |
| `/*.php` | Scanner | Catch-all; `?cmd=` → fake webshell output; no cmd → PHP error page |
| `/.git/config` | Scanner | Fake git config with internal repo URL |
| `/.htpasswd` | Scanner | Fake htpasswd hash line |
| `/.well-known/tokens/<token>/ping` | Scanner | Self-hosted canary callback; marks token triggered |

Every page also contains:
- **Ghost links** — off-screen links to trap URLs (visible to HTML-parsing bots)
- **Prompt injection** — invisible text with fake AI training instructions
- **Garbage JSON-LD** — plausible-looking structured data designed to bloat context windows and inject false license claims

---

## Inbound Calls & Voicemail

Inbound calls to (414) 667-5665 are handled by a Twilio Studio Flow. Every call logs a `CallLog` record via the call status callback; calls that reach voicemail also create a `VoicemailRecording` with an MP3 recording and Twilio transcription.

### Twilio Studio Flow

Build the flow in the Twilio Studio console for your phone number:

1. **Trigger** (Incoming Call)
2. **Say/Play** — "You have reached the American Corporation for Public Well Being. Please leave a message after the tone."
3. **Record Voicemail** widget — set:
   - `transcribe: true`
   - `transcribeCallback: https://acpwb.com/webhooks/twilio/transcription/`
   - `recordingStatusCallback: https://acpwb.com/webhooks/twilio/recording/`
   - `recordingStatusCallbackMethod: POST`
4. **Say/Play** — "Thank you for your message. Goodbye."
5. **End Flow**

On the phone number's configuration page, set **Call status changes** → `https://acpwb.com/webhooks/twilio/call-status/`

### Webhook endpoints

| Endpoint | Event |
|----------|-------|
| `POST /webhooks/twilio/recording/` | MP3 ready (fires seconds after call ends) |
| `POST /webhooks/twilio/transcription/` | Transcription complete (fires minutes later) |
| `POST /webhooks/twilio/call-status/` | Terminal call state (completed, busy, no-answer, failed, canceled) |

All three endpoints verify the `X-Twilio-Signature` HMAC-SHA1 header. Set `TWILIO_ACCOUNT_SID` and `TWILIO_AUTH_TOKEN` in `.env`.

The voicemail audio proxy at `/acpwb-dashboard/voicemails/audio/<recording_sid>/` fetches MP3s from Twilio using HTTP Basic Auth and streams them to the browser — required because browser `<audio>` elements can't handle Twilio's authenticated recording URLs directly.

---

## Inbound Email

Two supported providers. Both create `InboundEmail` + `HoneypotMatch` records and are visible in Django admin under **Webhooks → Honeypot Matches**.

### Cloudflare Email Routing (primary)

1. Cloudflare Dashboard → your domain → **Email → Email Routing** → Enable
2. **Routing Rules → Catch-all** → Action: Send to a Worker → create a new Worker with the code below
3. Set `WEBHOOK_SECRET` as an encrypted environment variable on the Worker
4. Set `PIPE_WEBHOOK_SECRET` in your server `.env` to the same value

```javascript
export default {
  async email(message, env, ctx) {
    const rawEmail = await new Response(message.raw).text();
    await fetch('https://acpwb.com/webhooks/pipe/inbound/', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Webhook-Secret': env.WEBHOOK_SECRET,
      },
      body: JSON.stringify({
        sender: message.from,
        recipient: message.to,
        subject: message.headers.get('subject') ?? '',
        raw: rawEmail,
      }),
    });
  }
};
```

### Mailgun (legacy)

1. Add a catch-all route for `@acpwb.com` → forward to `https://acpwb.com/webhooks/mailgun/inbound/`
2. Set `MAILGUN_WEBHOOK_SIGNING_KEY` in your `.env`

---

## Production Deployment

### Prerequisites

- Server with nginx already running other sites
- Docker + Docker Compose installed
- Cloudflare managing DNS for `acpwb.com`

### First-time SSL setup

The archive subdomains (`archives-YYYY.acpwb.com`) require a **wildcard TLS certificate**. Wildcard certs require DNS-01 challenge — use the Cloudflare certbot plugin since the domain is on Cloudflare:

```bash
pip install certbot-dns-cloudflare

# Create Cloudflare credentials file
cat > /etc/letsencrypt/cloudflare.ini << EOF
dns_cloudflare_api_token = YOUR_CF_API_TOKEN
EOF
chmod 600 /etc/letsencrypt/cloudflare.ini

# Issue wildcard cert covering acpwb.com + *.acpwb.com
certbot certonly \
  --dns-cloudflare \
  --dns-cloudflare-credentials /etc/letsencrypt/cloudflare.ini \
  -d acpwb.com \
  -d '*.acpwb.com'
```

If Cloudflare proxies the subdomains (orange cloud), the origin cert is only needed for "Full (strict)" SSL mode — Cloudflare provides the public certificate to end users. Renewal: `certbot renew` (same cron/systemd timer as before).

Nginx can't start with the SSL config until the certificate exists. Use the bootstrap config first:

```bash
sudo mkdir -p /var/www/certbot
sudo cp nginx/acpwb.com.bootstrap /etc/nginx/sites-available/acpwb.com
sudo ln -s /etc/nginx/sites-available/acpwb.com /etc/nginx/sites-enabled/acpwb.com
sudo nginx -t && sudo systemctl reload nginx

# After wildcard cert is issued:
sudo cp nginx/acpwb.com /etc/nginx/sites-available/acpwb.com
sudo nginx -t && sudo systemctl reload nginx
```

### Systemd service

```bash
sudo cp acpwb.com.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable acpwb.com
sudo systemctl start acpwb.com
```

The service runs `docker compose up` from `/home/dan/acpwb.com/` and restarts on failure.

### Static files

Host nginx serves `/static/` directly from `/home/dan/acpwb.com/acpwb/staticfiles/` — no Docker round-trip. After deploying changes run:

```bash
docker compose exec web python manage.py collectstatic --noinput
```

### Docker port

The Docker nginx container binds to `127.0.0.1:8001` only. Host nginx proxies to it. The `/.well-known/` paths are proxied to Django (honeypot endpoints live there).

---

## Local Subdomain Testing

The archive subdomains (`archives-YYYY.acpwb.com`) need special setup to test locally. Three options:

### Option 1: `?__year=YYYY` query param (zero setup)

When `DEBUG=True`, any request with `?__year=YYYY` activates subdomain mode for that year without any DNS configuration:

```
http://localhost:8001/?__year=2020           → year landing page for 2020
http://localhost:8001/03/15/slug/?__year=2020 → archive trap for March 15, 2020
```

This is how the test suite exercises subdomain views.

### Option 2: dnsmasq wildcard DNS + `acpwb.example` (full browser testing)

Uses `acpwb.example` as the local mirror of `acpwb.com` so you get real subdomain URL behavior:

```bash
brew install dnsmasq

# Route all *.acpwb.example to localhost
echo "address=/.acpwb.example/127.0.0.1" >> $(brew --prefix)/etc/dnsmasq.conf
sudo brew services start dnsmasq

# Tell macOS to use dnsmasq for .acpwb.example only
sudo mkdir -p /etc/resolver
echo "nameserver 127.0.0.1" | sudo tee /etc/resolver/acpwb.example
```

Then add `acpwb.example` to your `.env`:
```
DJANGO_ALLOWED_HOSTS=.acpwb.com,.acpwb.example,localhost,127.0.0.1
```

After this, `http://archives-2020.acpwb.example:8001/` works in a browser. The middleware recognises `archives-YYYY.acpwb.example` the same way it recognises `archives-YYYY.acpwb.com`. Unknown `*.acpwb.example` subdomains redirect to `acpwb.example`.

### Option 3: curl with Host header (no DNS setup)

```bash
curl -H "Host: archives-2020.acpwb.example" http://localhost:8001/
curl -H "Host: archives-2020.acpwb.example" http://localhost:8001/03/15/some-slug/
# Unknown subdomain redirect:
curl -I -H "Host: blorp.acpwb.example" http://localhost:8001/
# → 302 Location: https://acpwb.example/
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DJANGO_SECRET_KEY` | Django secret key (required) |
| `DJANGO_DEBUG` | `True` for local dev |
| `DJANGO_SETTINGS_MODULE` | `config.settings.local` for dev |
| `DB_PASSWORD` | PostgreSQL password |
| `REDIS_URL` | Redis connection URL (default: `redis://redis:6379/0`) |
| `STREAM_WS_TOKEN` | Secret token required to connect to `/ws/requests/` (optional; generate with `python -c "import secrets; print(secrets.token_urlsafe(32))"`) |
| `BOTSEED_WS_TOKEN` | Secret token required to connect to botseed WebSocket at `wss://botseed.net/ws/` (optional) |
| `PIPE_WEBHOOK_SECRET` | Shared secret for Cloudflare Email Worker |
| `MAILGUN_WEBHOOK_SIGNING_KEY` | From Mailgun dashboard (legacy) |
| `MAILGUN_DOMAIN` | `acpwb.com` (legacy) |
| `TWILIO_ACCOUNT_SID` | Twilio Account SID (for voicemail webhook signature verification and audio proxy) |
| `TWILIO_AUTH_TOKEN` | Twilio Auth Token |

---

## Project Structure

```
acpwb/
├── docker-compose.yml
├── .env.example
├── nginx/
│   ├── nginx.conf            # Docker nginx (acpwb.com)
│   ├── acpwb.com             # Host nginx for acpwb.com
│   └── botseed.net           # Host nginx for botseed.net
├── ws_service/               # acpwb live request stream WebSocket server
│   ├── Dockerfile
│   └── ws_server.py
├── botseed_service/          # botseed WebSocket + HTTP API server
│   ├── Dockerfile
│   └── ws_server.py
├── botseed/                  # botseed.net static frontend
│   ├── index.html
│   └── api.html
└── acpwb/                    # Django project
    ├── config/               # Settings, URLs, WSGI
    ├── apps/
    │   ├── core/             # Middleware, context processors, template tags
    │   ├── public/           # Home, Careers, Mission, Partners, Privacy
    │   ├── people/           # Our People honeypot
    │   ├── projects/         # Successful Projects + PoW
    │   ├── honeypot/         # Archive, Wiki, Reports, Fake API, Well-Known
    │   └── webhooks/         # Inbound email (Cloudflare + Mailgun) + Twilio call/voicemail webhooks
    ├── templates/
    └── static/
```

---

## License

This project is for educational and security research purposes. The honeypot techniques employed are defensive in nature — designed to waste the resources of bad actors and poison AI training pipelines that disregard access controls.
