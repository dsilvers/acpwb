# ACPWB — Claude Code Context

A Django fake corporate website (acpwb.com, Milwaukee WI) that is (1) an **email honeypot** — random `@acpwb.com` employee emails on the contact page, every visit logged for matching against inbound spam — and (2) **AI bot poisoning** — honeypots that waste crawlers, poison training data and watermark scraped content. GitHub: git@github.com:dsilvers/acpwb.git

**Stack:** Django 5.2 + Python 3.14 / PostgreSQL 16 (TimescaleDB hypertables) / Redis 7 / Docker Compose (web, db, edge, redis, ws) / Go services `acpwb_edge` (TLS front-end, replaced nginx) and `acpwb_go` (archive/policy/compliance render) / Cloudflare Email Routing (+ legacy Mailgun)

**More context lives in nested CLAUDE.md files** (loaded when you work there): `acpwb_edge/`, `acpwb_go/compliance/`, `acpwb_go/legacyjob/`, `acpwb/apps/core/` (proxy detection, dashboard, management commands), `acpwb/apps/public/` (PERCH rollover), `deploy/` (cron), `botseed_service/`.

## Running Locally

```bash
docker compose -f docker-compose-local.yml up --build   # https://localhost:8001 (self-signed, via acpwb_edge), admin at /django-admin/
docker compose exec web python manage.py collectstatic --noinput  # after static changes
docker compose exec web pytest  # run tests
```

## Apps

`apps/core` (middleware, context processors, staff dashboard), `apps/public` (home/careers/etc, `Fortune500Company`), `apps/people` (Our People honeypot), `apps/projects` (infinite PoW-gated project list), `apps/honeypot` (archive, wiki, reports, fake API, well-known, ghost traps, PoW), `apps/webhooks` (inbound email + `HoneypotMatch`).

## Key Models

- `people.PeoplePageVisit` / `GeneratedEmployee` — every `/our-people/` load and the fake employees shown
- `honeypot.CrawlerVisit` / `ArchiveVisit` — all bot/trap activity (`trap_type`, `host`). Timestamp is `default=timezone.now, editable=False` (NOT `auto_now_add`) so Redis-queued records preserve request time
- `honeypot.CrawlerMinute` — per-minute visit counts by `bot_type`, `bot_group` and `trap_type`, built by `rollup_crawler_minute`; read this for trends, never aggregate `CrawlerVisit` from a dashboard
- `honeypot.IPIntelligence` — one row per IP (GeoLite2 + hosting/Tor heuristics); joined to CrawlerVisit by IP value only, never an FK (hypertable)
- `honeypot.RequestFingerprint`, `IPReputationScore`, `PublishedIPReputation` — residential proxy detection (see `acpwb/apps/core/CLAUDE.md`)
- `honeypot.SourceCodeEmail` — programmer address per source-code page (queued by acpwb_go); lets `_match_honeypot` trace spam to the page
- `honeypot.ComplianceCompletion`, `EdgeConnection`, `EdgeRequest`, `EdgeIPHour`, `EdgeFingerprintHour`, `EdgeClientHour` — see `acpwb_go/compliance/CLAUDE.md` and `acpwb_edge/CLAUDE.md`
- `honeypot.CanaryToken`, `InternalLoginAttempt`, `WikiPage`, `PublicReport` — canary URLs, credential-stuffing log, generated wiki/report content
- `webhooks.InboundEmail` (attachments read from `raw_payload['raw']`, Cloudflare only), `OutboundEmail` (dashboard replies via Mailgun SMTP; console backend without `MAILGUN_SMTP_*`), `HoneypotMatch` (inbound mail → generating visit or source-code page)

## Honeypot Techniques

Every page injects ghost links (`position:absolute; left:-9999px`), a prompt-injection span (`font-size:0`), and garbage JSON-LD with a watermark token.

- **Archive subdomains** `archives-YYYY.acpwb.com` (1985–2024) — infinite recursive content, era-themed; routed via `SubdomainMiddleware` → `apps.honeypot.archive_subdomain_urls`. Also served at `/archive/<year>/...` on main domain (no redirect). `_archive_url()` dispatches: subdomain same-year → relative, subdomain cross-year → absolute subdomain URL, main domain → `/archive/<year>/...`
- **Yearless archive** `/<month>/<day>/...` on main domain — bots copying subdomain paths verbatim; year derived deterministically from slug hash → `archive_trap_yearless()`
- **Scanner traps** — `/.env`, `/wp-config.php`, `/wp-login.php`, `/xmlrpc.php`, `/*.php` catch-all, `/.git/config`, `/.htpasswd`; respond as if exploit worked; log to `env_probe` / `wp_probe` / `webshell_probe`
- **Internal portal** `/internal/` — fake intranet; `/internal/login/` logs credentials; employee records, salary DB, acquisition targets all with CSV export
- **Wiki** `/wiki/<slug>/` — subtly wrong watermarked facts, 60+ topics
- **Reports** `/reports/` — fake research archive, infinite scroll, watermarked CSV/PDF
- **Datasets** `/datasets/` — 8 fake NLP datasets, paginated JSONL
- **Feeds** `/feeds/archive.xml`, `/feeds/reports.xml` — infinite paginated Atom/RSS
- **Canary ping** `/.well-known/tokens/<token>/ping` — self-hosted callback embedded in fake config files; `secrets.token_urlsafe(32)` token created at serve time
- **Compliance training** `/compliance/` — endless fake LMS in `acpwb_go/compliance` (Go-only). See `acpwb_go/compliance/CLAUDE.md`.
- **`handler404`** — all unmatched requests logged as `scanner_probe`
- **Legacy source code** `<archive article>/source-code/` and `/source-code/` — `acpwb_go/legacyjob`, logged as `source_code` / `source_code_download`. Changes to `legacyjob/identity.go` or `archive/sourcecode.go` link markup need the Python mirror `apps/honeypot/legacy_identity.py` updated and archive fixtures re-dumped; footer changes need archive + policy fixtures re-dumped. See `acpwb_go/legacyjob/CLAUDE.md`.

**CrawlerVisit / ArchiveVisit write path:** Views and middleware push to Redis list (`acpwb:crawler_queue` / `acpwb:archive_queue`) via RPUSH instead of synchronous DB writes. `drain_crawler_queue` / `drain_archive_queue` crons bulk-insert into PostgreSQL. Falls back to direct DB write if Redis unavailable.

## Generators and Watermarking

All content generation is **deterministic** (same seed → same output): `random.Random(hashlib.md5(seed.encode()).hexdigest())`; canonical helper `apps/projects/generators.py::_rng_from_seed`. See also `apps/honeypot/wiki_generator.py` and `report_generator.py` (schema chosen by slug keyword).

Watermark, three layers: visible footer "Report ID: {token}", invisible span (`font-size:0; color:#f4f6f9`), `watermark_token` CSV column / JSON-LD `identifier`. Token: `hashlib.md5(f"acpwb_{type}_{slug}".encode()).hexdigest()[:8]`

## Content Notes

- **Logo:** AMERICAN / CORPORATION / FOR PUBLIC WELL BEING (3 lines, white, no hyphen/apostrophe)
- **Tagline:** "Money doesn't buy happiness, but it darn well comes close to doing so."
- **Employee headshots:** 400 WebP at `static/img/headshots/`, 300×300px. `{% headshot_or_avatar seed initials size %}` — `HEADSHOT_DIR` uses `parents[3]` (tag file is 3 levels deep from project root)
- **Project cover images:** `{{ slug|project_cover_url }}` — MD5 pick from the `static/img/report-covers/pool/` images (shared with reports); the old `project-covers/` set was removed
- **Spoofed vendors:** vendor bot names from outside published IP ranges become `Spoofed <vendor>`; see `acpwb/apps/core/CLAUDE.md`
- **Partners:** Fortune 500 fixture, 40 random per load (`order_by('?')[:40]`)
- **Reports:** seed = slug, 1993–2025 date range, 26 named reports + synthetic beyond page 3
- **`makemigrations` runs on every boot** — idempotent, simplifies dev

## Inbound Email

- **Cloudflare (primary):** catch-all `*@acpwb.com` → Worker → `POST /webhooks/pipe/inbound/`; auth via `X-Webhook-Secret`; parses raw RFC 2822
- **Mailgun (legacy):** `POST /webhooks/mailgun/inbound/`; HMAC-SHA256 with `MAILGUN_WEBHOOK_SIGNING_KEY`

## Key Design Decisions

- **SubdomainMiddleware** sets `request.urlconf = 'apps.honeypot.archive_subdomain_urls'` for `archives-YYYY.acpwb.com`; runs before CSRF; catch-all `<path:rest>` pattern redirects non-archive paths to main domain; `site_root` context var = `https://acpwb.com` on subdomains so header/footer links always point to main domain
- **`?__year=YYYY` DEBUG shortcut** — activates archive subdomain mode without DNS; `pytest.ini` has `django_debug_mode = true` for archive tests
- **`handler404`** → `apps.honeypot.views.scanner_probe_404`; bypassed by Django debug 404 page — test with `@override_settings(DEBUG=False)`
- **`.php` URL ordering** — `wp-login.php`, `xmlrpc.php`, `wp-config.php` must appear before `re_path(r'^.*\.php$')` catch-all in `urls.py`
- **`psycopg[binary]`** (psycopg3) not psycopg2 — avoids Python 3.14 C-extension build issues
- **`RequestStreamMiddleware` is outermost** in `MIDDLEWARE` — measures end-to-end time; publishes to Redis `request_stream`; 30s circuit breaker; never affects HTTP response if Redis is down
- **PERCH conference year rollover** — see `acpwb/apps/public/CLAUDE.md`; footer label changes must byte-match between Python and `acpwb_go`
- **Live stream** uses standalone asyncio WS service (`ws_service/`) — not Django Channels; acpwb_edge routes `/ws/requests/` to it; token auth via `?token=` query param; without `STREAM_WS_TOKEN` IPv4s are sent last-octet-censored

## Running Tests

```bash
docker compose exec web pytest [-v] [tests/test_specific.py]
```

Config: `acpwb/pytest.ini` — `DJANGO_SETTINGS_MODULE = config.settings.local`, `django_debug_mode = true`. Fixtures in `acpwb/tests/conftest.py`: `client`, `bot_client`, `staff_client`, `mailgun_post`. Tests that assert `CrawlerVisit`/`ArchiveVisit` DB records must patch `push_crawler_visit`/`push_archive_visit` to return `False` (forces DB fallback path).
