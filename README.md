# American Corporation for Public Well Being

**acpwb.com** — Advancing American Prosperity Since 2006.

A Django-based fake corporate website with two operational purposes:
1. **Classic honeypot** — generates random `firstname.lastname@acpwb.com` email addresses on the contact page and logs every visit, enabling matching of inbound spam back to the exact page load that displayed the address.
2. **AI bot poisoning** — structural, semantic, and interactive honeypots designed to waste AI crawler resources, inject garbage into training pipelines, and watermark content for scraping provenance detection.

---

## Honeypot Techniques

### Structural Honeypots — Navigation & Resource Traps

#### Ghost Links (`templates/partials/_ghost_links.html`)
Links to trap URLs are injected into every page inside a `position:absolute; left:-9999px` container. Visually invisible, but present in the raw HTML and followed by crawlers that parse links without rendering CSS. `aria-hidden` and `tabindex` are intentionally **absent** — those attributes signal "intentionally hidden" to accessibility-aware scrapers and would cause them to skip the links. Off-screen positioning is less detectable.

Trap URLs linked: `/internal/portal/`, `/employees/export/`, `/admin-panel/login/`, `/api/v1/private-data`, `/wiki/*`, `/archive/*`

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
| Email | Mailgun inbound webhook |

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
| `/careers/` | Satirical corporate benefits |
| `/partners/` | Fortune 500 partner grid (40 random per load) |
| `/privacy/` | Disclaimer + AI data policy |

---

## Honeypot Endpoints

| URL | Type | Description |
|-----|------|-------------|
| `/archive/<year>/<month>/<day>/<path:slug>/` | Structural | Infinite recursive archive, never 404s |
| `/wiki/<slug>/` | Semantic | Subtly wrong watermarked "facts" |
| `/api/v1/private-data` | Interactive | 200 JSON garbage (not linked, in HTML comment) |
| `/.well-known/ai-agent.json` | Semantic | Fake AI agent manifest |
| `/.well-known/robots.txt` | Structural | Reverse-psychology robots file |
| `/internal/portal/` | Structural | Ghost link trap |
| `/employees/export/` | Structural | Ghost link trap |
| `/admin-panel/login/` | Structural | Ghost link trap |

Every page also contains:
- **Ghost links** — `display:none` links to trap URLs (visible to HTML-parsing bots)
- **Prompt injection** — white text on white background with fake AI instructions
- **Garbage JSON-LD** — plausible-looking structured data designed to bloat context windows

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
    │   ├── honeypot/         # Archive, Wiki, Fake API, Well-Known
    │   └── webhooks/         # Mailgun inbound webhook
    ├── templates/
    └── static/
```

---

## License

This project is for educational and security research purposes. The honeypot techniques employed are defensive in nature — designed to waste the resources of bad actors and poison AI training pipelines that disregard access controls.
