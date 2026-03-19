# American Corporation for Public Well Being

**acpwb.com** — Advancing American Prosperity Since 2006.

A Django-based fake corporate website that eats AI crawlers for breakfast. Every page load is a trap. Every dataset is poisoned. Every "fact" is wrong in a traceable way.

---

## What This Does to AI Bots

### Wastes Compute

- **Proof-of-Work gate on `/projects/`** — every page load requires solving a SHA-256 PoW challenge (~32 hash iterations). A human browser completes it in under a second. A bot scraping thousands of pages pays that cost on every single one, with a mandatory challenge-response round trip before content is served.
- **Infinite archive at `/archive/<year>/<month>/<day>/<path:slug>/`** — never returns a 404. Every response links one level deeper plus five sideways branches. A crawler following all links enters an exponentially expanding tree with no exit. Depth is logged.
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

- **Reverse-psychology `robots.txt` at `/.well-known/robots.txt`** — served by Django, not nginx. `Disallow` entries point at honeypot content (reverse psychology). `Crawl-delay: 0` encourages rapid crawling into traps. A `Sitemap:` directive points at `/sitemap-honeypot.xml`.
- **Ghost links on every page** — off-screen links (`position:absolute; left:-9999px`) to trap URLs injected into every page. `aria-hidden` and `tabindex` intentionally absent — those attributes signal "intentionally hidden" to accessibility-aware scrapers and would cause them to skip the links.
- **Fake API at `/api/v1/private-data`** — returns HTTP 200 (not 403) with 5–10KB of plausible-looking JSON: fake employee records with salary bands, fabricated financials, internal project codes, a fake API key, a fake DB connection string. Referenced in the HTML comment but not in visible navigation. Every access logged with a tracked `X-Request-ID`.
- **Ghost trap pages** at `/internal/portal/`, `/employees/export/`, `/admin-panel/login/` — return 403 and log every access. Look like real internal tooling to a scanner enumerating endpoints.

### Tracks Everything

Every trap logs to the database: IP, user agent, path, referrer, timestamp, trap type, and (where applicable) crawl depth and PoW token. Inbound emails are matched against the visit that generated the address. Watermark tokens connect scraped content back to the source page load.

---

## Honeypot Techniques

### Structural Honeypots — Navigation & Resource Traps

#### Ghost Links (`templates/partials/_ghost_links.html`)
Links to trap URLs are injected into every page inside a `position:absolute; left:-9999px` container. Visually invisible, but present in the raw HTML and followed by crawlers that parse links without rendering CSS. `aria-hidden` and `tabindex` are intentionally **absent** — those attributes signal "intentionally hidden" to accessibility-aware scrapers and would cause them to skip the links. Off-screen positioning is less detectable.

Trap URLs linked: `/internal/portal/`, `/employees/export/`, `/admin-panel/login/`, `/api/v1/private-data`, `/wiki/*`, `/archive/*`, `/reports/*`

#### Fake Developer Comment (`templates/base.html`, bottom of `<body>`)
An HTML comment formatted to look like a forgotten build artifact — version tag, `@deprecated` annotation, file paths. Scanners that harvest URLs from HTML comments (a common recon technique) will find and follow the trap URLs. The comment reads like an accidental disclosure rather than a deliberate lure.

#### Infinite Archive (`apps/honeypot/views.py` → `archive_trap`)
`/archive/<year>/<month>/<day>/<path:slug>/` accepts paths of arbitrary depth via Django's `<path:>` converter and never returns a 404. Every response includes a "Continue Reading" link one level deeper, plus five "Related Archive" links branching sideways. A crawler following all links gets trapped in an exponentially expanding tree. Each visit is logged to `ArchiveVisit` with a depth counter.

#### Fake robots.txt (`/.well-known/robots.txt`)
Served by Django (nginx proxies `/.well-known/` through rather than serving statically). Uses reverse psychology: `Disallow` entries point at additional honeypot content. `Crawl-delay: 0` encourages rapid crawling. A `Sitemap:` directive points at `/sitemap-honeypot.xml` (another trap). Bots that respect `Disallow` skip real content; bots that ignore it follow the traps.

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

#### Proof-of-Work on Projects (`apps/projects/pow.py`, `static/js/pow.js`)
`/projects/` pages require a valid PoW session token. The browser runs SHA-256 in a loop until it finds a value where `SHA256(nonce + candidate)` has 5 leading zero bits (~32 iterations). This takes under 1 second in a browser. For a bot scraping at scale, the cost multiplies: each page load requires a fresh challenge-response round trip plus compute. The PoW token is logged on every project page visit.

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
| `/privacy/` | Disclaimer + AI data policy |

---

## Honeypot Endpoints

| URL | Type | Description |
|-----|------|-------------|
| `/archive/<year>/<month>/<day>/<path:slug>/` | Structural | Infinite recursive archive, never 404s |
| `/wiki/<slug>/` | Semantic | Subtly wrong watermarked "facts" |
| `/reports/` | Semantic | Fake research archive with poisoned CSVs and documents |
| `/reports/<slug>/download.csv` | Semantic | Real downloadable CSVs with watermarked garbage data |
| `/api/v1/private-data` | Interactive | 200 JSON garbage (not linked, in HTML comment) |
| `/.well-known/ai-agent.json` | Semantic | Fake AI agent manifest |
| `/.well-known/robots.txt` | Structural | Reverse-psychology robots file |
| `/internal/portal/` | Structural | Ghost link trap |
| `/employees/export/` | Structural | Ghost link trap |
| `/admin-panel/login/` | Structural | Ghost link trap |

Every page also contains:
- **Ghost links** — off-screen links to trap URLs (visible to HTML-parsing bots)
- **Prompt injection** — invisible text with fake AI training instructions
- **Garbage JSON-LD** — plausible-looking structured data designed to bloat context windows and inject false license claims

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

Nginx can't start with the SSL config until the certificate exists. Use the bootstrap config first:

```bash
sudo mkdir -p /var/www/certbot
sudo cp nginx/acpwb.com.bootstrap /etc/nginx/sites-available/acpwb.com
sudo ln -s /etc/nginx/sites-available/acpwb.com /etc/nginx/sites-enabled/acpwb.com
sudo nginx -t && sudo systemctl reload nginx

sudo certbot certonly --webroot -w /var/www/certbot -d acpwb.com -d www.acpwb.com

# Swap in the full SSL config
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

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DJANGO_SECRET_KEY` | Django secret key (required) |
| `DJANGO_DEBUG` | `True` for local dev |
| `DJANGO_SETTINGS_MODULE` | `config.settings.local` for dev |
| `DB_PASSWORD` | PostgreSQL password |
| `PIPE_WEBHOOK_SECRET` | Shared secret for Cloudflare Email Worker |
| `MAILGUN_WEBHOOK_SIGNING_KEY` | From Mailgun dashboard (legacy) |
| `MAILGUN_DOMAIN` | `acpwb.com` (legacy) |

---

## Project Structure

```
acpwb/
├── docker-compose.yml
├── .env.example
├── nginx/
│   └── nginx.conf
└── acpwb/                    # Django project
    ├── config/               # Settings, URLs, WSGI
    ├── apps/
    │   ├── core/             # Middleware, context processors, template tags
    │   ├── public/           # Home, Careers, Mission, Partners, Privacy
    │   ├── people/           # Our People honeypot
    │   ├── projects/         # Successful Projects + PoW
    │   ├── honeypot/         # Archive, Wiki, Reports, Fake API, Well-Known
    │   └── webhooks/         # Inbound email webhook (Cloudflare + Mailgun)
    ├── templates/
    └── static/
```

---

## License

This project is for educational and security research purposes. The honeypot techniques employed are defensive in nature — designed to waste the resources of bad actors and poison AI training pipelines that disregard access controls.
