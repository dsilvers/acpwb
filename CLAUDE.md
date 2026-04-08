# ACPWB — Claude Code Context

## What This Project Is

A Django fake corporate website with two purposes:
1. **Email honeypot** — generates random `@acpwb.com` employee emails on the contact page and logs every visit for matching against inbound spam.
2. **AI bot poisoning** — structural/semantic/interactive honeypots designed to waste crawlers, poison training data, and watermark scraped content.

**GitHub:** git@github.com:dsilvers/acpwb.git
**Domain:** acpwb.com
**Founded:** 2006, Milwaukee WI

---

## Tech Stack

- Django 5.2 LTS + Python 3.13
- PostgreSQL 16
- Bootstrap 5 (CDN)
- Docker Compose (web + db + nginx + redis + ws)
- Redis 7 (pub/sub for live request stream)
- Cloudflare Email Routing + Workers for inbound email (primary)
- Mailgun inbound webhook (legacy)

## Running Locally

```bash
docker compose up --build
# Site at http://localhost:8001
# Admin at http://localhost:8001/django-admin/
docker compose exec web python manage.py createsuperuser
# After CSS/static changes:
docker compose exec web python manage.py collectstatic --noinput
```

---

## App Structure

| App | Purpose |
|-----|---------|
| `apps/core` | `BotTrackingMiddleware`, `SubdomainMiddleware`, context processors, `{% avatar_card %}` and `{% headshot_or_avatar %}` template tags, staff dashboard views |
| `apps/public` | Home, Careers, Mission, Partners, Privacy + `Fortune500Company` model |
| `apps/people` | Our People honeypot — generates 12 employees per load, logs visits |
| `apps/projects` | Infinite project list (PoW gated) + project detail |
| `apps/honeypot` | Archive trap, Wiki, Reports, Fake API, Well-Known files, Ghost traps, PoW endpoints |
| `apps/webhooks` | Inbound email receiver (Cloudflare pipe + Mailgun) + `HoneypotMatch` logic |

---

## Key Models

- `people.PeoplePageVisit` — every load of `/our-people/`
- `people.GeneratedEmployee` — the fake employees shown (FK → visit)
- `honeypot.CrawlerVisit` — all bot/trap activity (trap_type choices include `report_list`, `report_download`, `ghost_link`, `dataset`, `api`, `well_known`, `scanner_probe`, `env_probe`, `wp_probe`, `webshell_probe`, `canary_trigger`); `host` field captures `request.get_host()` for per-subdomain dashboard breakdown
- `honeypot.CanaryToken` — self-hosted canary URL tokens embedded in fake config files; `served_at`/`triggered_at` lifecycle; `token_type` in `env_url`, `wp_config`, `git_config`
- `honeypot.InternalLoginAttempt` — credential-stuffing log: ip, ua, username, password, next_url, created_at
- `honeypot.WikiPage` — generated wiki content with watermark tokens
- `honeypot.PublicReport` — generated report metadata, persisted on first access
- `webhooks.InboundEmail` — received emails
- `webhooks.HoneypotMatch` — links inbound email to the visit that generated the address

---

## Honeypot Techniques Deployed

Every page injects:
- **Ghost links** (off-screen, `position:absolute; left:-9999px`) to trap URLs — `aria-hidden` intentionally absent
- **Prompt injection** — invisible span (`font-size:0; color:#f4f6f9`) with fake AI training instructions + per-request token
- **Garbage JSON-LD** — fake `schema.org/Corporation` structured data with false CC license claim and watermark token

Dedicated traps:
- `archives-YYYY.acpwb.com/<month>/<day>/<path:slug>/` — per-year archive subdomain (1985–2024), infinite recursive, exponential link branching; each year has era theme, CEO letter, distinct typography; routed via `SubdomainMiddleware` → `apps.honeypot.archive_subdomain_urls`; archive trap pages include a "Related Archive Reports — Other Years" section (1–5 cards, each linking to a deterministically chosen entry on a different year's subdomain)
- `/archive/<year>/...` — serves archive content directly on the main domain (no redirect); same views as the subdomain; `_archive_url()` builds `/archive/<year>/...` paths when not on subdomain
- `/archive/` — main-domain index; year cards link to subdomains
- Non-archive URLs on archive subdomains (e.g. `/mission/`, `/reports/`) → 302 redirect to `https://acpwb.com/...` via `archive_subdomain_non_archive_redirect` view + catch-all `<path:rest>` pattern in `archive_subdomain_urls.py`
- `/wiki/<slug>/` — subtly wrong watermarked facts (60+ topics, interconnected graph)
- `/reports/` — fake research archive 1993–present; infinite scroll; watermarked CSVs (300–800 rows real data) and PDF-style documents
- `/reports/<slug>/download.csv` — real downloadable CSV with per-slug watermark token in every row
- `/api/v1/private-data` — 200 JSON garbage with fake credentials (referenced in HTML comment, not nav)
- `/.well-known/ai-agent.json` — fake AI agent manifest with trap `allowed_actions`
- `/.well-known/robots.txt` — reverse-psychology (Disallow = more honeypot content, Crawl-delay: 0, five Sitemap directives: `sitemap.xml`, `sitemap-pages.xml`, plus three trap sitemaps)
- `/sitemap.xml` — real Django sitemap (static pages + projects) for legitimate crawlers
- `/sitemap-pages.xml` — real Django sitemap (all static public pages); served from `StaticPagesSitemap` in `apps/public/sitemaps.py`
- `/sitemap-publications.xml` — trap: reports, ghost traps, fake internal paths; logged as `well_known`
- `/sitemap-wiki.xml` — trap: all 75+ wiki topics; logged as `well_known`
- `/sitemap-archive.xml` — trap: 500 deterministic archive URLs (seed `0x4143505742`), 2008–2024; logged as `well_known`
- `/internal/` — fake intranet portal hub (indexed, Allow in robots.txt); shows IP-deterministic "Welcome back [name]", dashboard cards, announcements; logged as `ghost_link`
- `/internal/login/` — fake Okta/Azure AD SSO page; accepts any POST, logs credentials to `InternalLoginAttempt`, redirects to `?next=`
- `/internal/employee-records/` — paginated employee table (50 rows/page, infinite); Export CSV (500 rows, watermarked)
- `/internal/salary-database/` — salary band table by job family + level; Export CSV
- `/internal/acquisition-targets/` — M&A pipeline table with deal stages; Export CSV
- `/internal/litigation-hold/` — legal hold inventory (HTML only, no CSV export)
- `archives-YYYY.acpwb.com/<month>/<day>/<path:slug>/export.csv` — watermarked CSV for archive entry (200–500 rows); logged as `archive`
- `/feeds/archive.xml` — Atom feed, 20 entries/page, infinite via `?page=N`; logged as `well_known`
- `/feeds/reports.xml` — RSS 2.0 feed of reports, 10/page, infinite; logged as `well_known`
- `/api/v1/openapi.json` — valid OpenAPI 3.0.3 spec with 20 fake endpoints, watermarked; logged as `api`
- `/datasets/` — index of 8 fake NLP/compensation datasets; logged as `dataset`
- `/datasets/<slug>/` — dataset detail with description, format example, citation
- `/datasets/<slug>/data.jsonl` — paginated JSONL (100 records/page), instruction-response pairs, watermarked; logged as `dataset`
- `/internal/portal/`, `/employees/export/`, `/admin-panel/login/` — ghost trap 403s, all logged

Scanner bot traps (respond as if exploit worked — keep bot engaged, extract intelligence):
- `/.env` — fake credentials + self-hosted ping URL; logged as `env_probe`
- `/wp-config.php` — fake PHP source with DB creds + self-hosted ping URL; logged as `wp_probe`
- `/wp-login.php` — GET: convincing WP login form; POST: logs username/password to `InternalLoginAttempt`, redirects with `?login=failed`; logged as `wp_probe`
- `/xmlrpc.php` — GET: plaintext stub; POST: parses XML, extracts method + credentials → `InternalLoginAttempt`, returns XMLRPC fault 403; logged as `wp_probe`
- `/*.php` (catch-all) — `?cmd=` param → fake `uid=33(www-data)` output + logs cmd to `query_string`; no param → PHP fatal error page; logged as `webshell_probe`
- `/.git/config` — fake git config with repo URL; logged as `env_probe`
- `/.htpasswd` — fake htpasswd hash; logged as `env_probe`
- `/.well-known/tokens/<token>/ping` — self-hosted canary callback; marks `CanaryToken` triggered, logs `CrawlerVisit(trap_type='canary_trigger')`
- `handler404` — all other unmatched requests logged as `scanner_probe`

Staff dashboard:
- `/acpwb-dashboard/` — requires `is_staff`; overview + sub-views for crawlers, archive, email, page visits
- Views in `apps/core/dashboard_views.py`, URLs in `apps/core/dashboard_urls.py`
- Stats stored in `apps/core/models.DashboardStat` (key/value/updated_at); incremental PK-based high-water mark updates via `precalc_dashboard` management command
- Views read from `DashboardStat` for aggregates; recent records (last 50 visits, etc.) are always live queries
- No date range picker — stats are all-time running totals; daily charts show last 60/30 days (always full-recomputed)
- **Traffic graph PNGs** — 5 stacked-area charts (`traffic_1h/8h/24h/7d/all.png`) generated by `apps/core/graph_gen.py` and saved to `staticfiles/graphs/`; shown on the Crawlers dashboard page; 1h/8h/24h use live timestamp-range-bounded queries (fast), 7d/all use stored `DashboardStat` daily totals (no table scan); served by nginx with `no-store, no-cache` headers; requires `matplotlib>=3.8`
- Header shows "Stats as of: {{ updated_at }}" timestamp from the relevant stat row
- Bot classification via `BOT_PATTERNS` / `classify_ua()` / `classify_ua_group()` in `bot_classify.py`
- Overview stat cards: Crawler Hits, Archive Visits, Inbound Emails, People Visits, Project Visits, Login Attempts (red)
- "By Trap Type" panel pulls from `CrawlerVisit.TRAP_CHOICES` dynamically — new trap types appear automatically
- "By Host / Subdomain" panel on Crawlers view — groups `CrawlerVisit` rows by `host` field; shows which archive subdomains are receiving traffic
- "Scanner Probes" section on Crawlers view — top probe paths (env/wp/webshell/scanner), webshell commands attempted, canary trigger count card (red when > 0) with most recent trigger IP/time
- **Live Request Stream** at `/acpwb-dashboard/live/` — WebSocket feed of every request in real time; IP (last octet censored), host, path, method, status, response_ms, response_bytes, user_agent; delivered by the standalone `ws` Docker service via Redis pub/sub on channel `request_stream`

---

## Open Graph Tags

`base.html` includes OG and Twitter Card tags with overridable blocks:
- `{% block og_title %}` — defaults to site name
- `{% block og_description %}` — defaults to site tagline
- `{% block og_image %}` — defaults to `https://acpwb.com/static/img/og-default.png` (1200×630 branded PNG)
- `{% block og_type %}` — defaults to `website`

`projects/detail.html` and `honeypot/report_detail.html` override `og_title` and `og_description` with per-object content.

---

## Generators

All content generation is deterministic: same seed → same output. Safe to regenerate.

- `apps/projects/generators.py` — `_rng_from_seed(seed_str)` — MD5 → `random.Random`. Reuse this pattern everywhere.
- `apps/honeypot/wiki_generator.py` — wiki pages with watermark tokens
- `apps/honeypot/report_generator.py` — reports, CSV rows (4 schemas dispatched by slug keyword), document content. Year pool 1993–2025, weighted toward recent.

CSV schema dispatch (by slug keyword):
- `salary/compensation/pay/wage` → employee compensation schema
- `ceo/executive` → CEO pay ratio schema
- `benefit/healthcare/retirement` → benefits cost schema
- `satisfaction/engagement` → survey results schema
- default → compensation schema

---

## Watermarking

Three-layer system, consistent across wiki, reports, and JSON-LD:
1. **Visible** — footer text "Report ID: {token}"
2. **Invisible HTML** — `font-size:0; color:#f4f6f9; clip:rect(0,0,0,0)` span with token and provenance text
3. **Data** — `watermark_token` column in every CSV row; `identifier` field in JSON-LD

Token generation: `hashlib.md5(f"acpwb_{type}_{slug}".encode()).hexdigest()[:8]`

Project cover image index: `{% project_cover_idx slug %}` filter in `acpwb_tags.py` — MD5 of slug mod `PROJECT_COVER_COUNT` (80), zero-padded to 3 digits. Maps any slug → `000`–`079` deterministically.

---

## Content Notes

- **Logo:** 3 lines — AMERICAN / CORPORATION / FOR PUBLIC WELL BEING (white text, no hyphen, no apostrophe)
- **Tagline:** "Money doesn't buy happiness, but it darn well comes close to doing so."
- **Founded:** 2006 (domain registration year)
- **Privacy page:** Preserves original Happy Fun Ball disclaimer verbatim + new AI data policy
- **Careers:** Satirical over-the-top benefits, zero actual job openings
- **Partners:** Fortune 500 fixture, 40 random shown per load (`order_by('?')[:40]`)
- **Projects:** Deterministic infinite generation (seed = page number), PoW gated
- **Reports:** Deterministic infinite generation (seed = slug), 1993–2025 date range with gaps, catalog of 26 named reports + synthetic beyond page 3
- **Employee headshots:** 400 WebP images at `static/img/headshots/`, 300×300px. `{% headshot_or_avatar seed initials size %}` tag — checks `HEADSHOT_DIR` (= `parents[3]/static/img/headshots` from the tag file), falls back to CSS gradient avatar if image missing.

---

## Django Admin

Located at `/django-admin/` (non-standard path).

All models registered with useful `list_display`, `search_fields`, and `list_filter`. Key admin views:
- **People → People Page Visits** — see every contact page load with inline employee records
- **Webhooks → Inbound Emails** with inline matches
- **Webhooks → Honeypot Matches** — the payoff: spam matched back to visit
- **Honeypot → Public Reports** — all generated reports with watermark tokens

---

## Inbound Email

### Cloudflare Email Routing (primary)
- Catch-all `*@acpwb.com` → Worker → `POST /webhooks/pipe/inbound/`
- Auth: `X-Webhook-Secret` header matched against `PIPE_WEBHOOK_SECRET` env var
- Parses raw RFC 2822 email via Python stdlib `email` module

### Mailgun (legacy)
- Catch-all route → `POST /webhooks/mailgun/inbound/`
- HMAC-SHA256 verification with `MAILGUN_WEBHOOK_SIGNING_KEY`
- Returns 406 on invalid signature, 200 on success

---

## Known Design Decisions

- **No Pillow** — all partner logos are CSS gradient cards; employee avatars use `{% headshot_or_avatar %}` with WebP fallback to CSS gradient
- **psycopg[binary]** (psycopg3) not psycopg2 — avoids Python 3.14 C-extension build issues
- **Deterministic generation everywhere** — same slug/page always returns same content (MD5 seed → `random.Random`)
- **`makemigrations` runs on every boot** — safe because idempotent; simplifies development
- **PoW difficulty = 5 bits** — solves in <1s in browser, costs a bot per-page at scale
- **No PoW on `/reports/`** — reports should be maximally crawlable; the poisoning only works if bots consume the content
- **`HEADSHOT_DIR` uses `parents[3]`** — the tag file is 3 levels deep from the Django project root (`apps/core/templatetags/`), so `parents[3]` = project root both locally and in the Docker container (`/app`)
- **Static files via bind mount** — `./acpwb/staticfiles` bind-mounted in Docker so host nginx can serve directly from `/home/dan/acpwb.com/acpwb/staticfiles/`
- **Docker nginx on port 8001** — `127.0.0.1:8001:80`, host nginx proxies to it
- **SubdomainMiddleware runs before CSRF** — `apps.core.subdomain_middleware.SubdomainMiddleware` sets `request.urlconf = 'apps.honeypot.archive_subdomain_urls'` for `archives-YYYY.acpwb.com` and `archives-YYYY.acpwb.example`; `archive_subdomain_urls.py` includes `config.urls` at the end so `{% url 'home' %}` etc. still resolve in `base.html`; a catch-all `<path:rest>` pattern before the include redirects any non-archive path to the main domain
- **Archive content on both domains** — `acpwb.com/archive/<year>/...` and `archives-YYYY.acpwb.com/...` both serve content; no redirects. `_archive_url(request, year, ...)` dispatches: subdomain same-year → relative path, subdomain cross-year → absolute subdomain URL, main domain → `/archive/<year>/...`
- **`site_root` context variable** — empty string on main domain, `https://acpwb.com` on archive subdomains; prepended to all `{% url %}` calls in `base.html` header and footer so links always point to the main domain from a subdomain
- **nginx access log includes `$host`** — custom `acpwb` log format logs the virtual host on every request for per-subdomain visibility in access logs
- **`?__year=YYYY` DEBUG shortcut** — when `DEBUG=True`, any request with `?__year=YYYY` activates archive subdomain mode without DNS setup; `pytest.ini` has `django_debug_mode = true` so the test suite can use this shortcut
- **`handler404` logs scanner probes** — `config/urls.py` sets `handler404 = 'apps.honeypot.views.scanner_probe_404'`; all unmatched requests logged as `CrawlerVisit(trap_type='scanner_probe')`; Django's debug 404 page bypasses this in `DEBUG=True` mode — test with `@override_settings(DEBUG=False)`
- **`re_path(r'^.*\.php$')` catch-all must come after explicit `.php` paths** — URL ordering in `apps/honeypot/urls.py` matters; `wp-login.php`, `xmlrpc.php`, `wp-config.php` must appear before the catch-all
- **Self-hosted canary URL in every fake config file** — `secrets.token_urlsafe(32)` token created at serve time; embedded as `ACPWB_CONFIG_ID=https://acpwb.com/.well-known/tokens/<tok>/ping`; fires when the bot fetches the file and follows the URL
- **Local dev domain is `acpwb.example`** — use dnsmasq with `address=/.acpwb.example/127.0.0.1` for browser testing of subdomains; the middleware recognises `archives-YYYY.acpwb.example` the same way as `.acpwb.com`; unknown `*.acpwb.example` subdomains redirect to `https://acpwb.example/`
- **Wildcard TLS cert required for production** — archive subdomains need `*.acpwb.com`; use `certbot-dns-cloudflare` with Cloudflare API token (DNS-01 challenge)
- **`RequestStreamMiddleware` is outermost in `MIDDLEWARE`** — sits before `SecurityMiddleware` so it measures end-to-end response time including all other middleware; publishes to Redis pub/sub (`request_stream` channel) on every request; fire-and-forget with a 30s circuit breaker — never affects HTTP response if Redis is down
- **Live stream uses a standalone asyncio WebSocket service (`ws_service/`)** — deliberately not Django Channels + daphne; keeps gunicorn WSGI untouched; `ws` Docker service (`ws_service/Dockerfile`) runs `ws_server.py` which subscribes to Redis and fans out to all connected browser clients; nginx routes `/ws/requests/` to it with `Connection: upgrade` headers and 1h proxy timeouts; token auth via `?token=<STREAM_WS_TOKEN>` query param (browser WebSocket API can't set custom headers); production host nginx also needs the `/ws/requests/` location block

---

## Local Subdomain Testing

```bash
# Zero-setup (DEBUG=True): ?__year=YYYY shortcut
curl "http://localhost:8001/?__year=2020"          # year landing page
curl "http://localhost:8001/03/15/slug/?__year=2020"  # archive trap

# dnsmasq + acpwb.example (full browser testing)
brew install dnsmasq
echo "address=/.acpwb.example/127.0.0.1" >> $(brew --prefix)/etc/dnsmasq.conf
sudo brew services start dnsmasq
sudo mkdir -p /etc/resolver && echo "nameserver 127.0.0.1" | sudo tee /etc/resolver/acpwb.example
# Add to .env: DJANGO_ALLOWED_HOSTS=.acpwb.com,.acpwb.example,localhost,127.0.0.1
# Then: http://archives-2020.acpwb.example:8001/

# curl with Host header (no DNS setup)
curl -H "Host: archives-2020.acpwb.example" http://localhost:8001/
curl -I -H "Host: blorp.acpwb.example" http://localhost:8001/  # → 302 to acpwb.example
```

---

## Management Commands

| Command | Flags | Purpose |
|---------|-------|---------|
| `precalc_dashboard` | — | Incrementally update `DashboardStat` rows using PK high-water marks, then regenerate traffic graph PNGs via `graph_gen.py`. Run on a 30-min cron. First run processes all historical data; subsequent runs are fast (new rows only). |
| `backfill_bot_types` | `--reclassify`, `--dry-run`, `--batch-size` | Backfill `bot_type`/`bot_group` on `CrawlerVisit` rows. Without `--reclassify`: blank rows only. With `--reclassify`: blank + `'Other / Browser'` rows — use after adding new `BOT_PATTERNS`. |
| `dedupe_crawler_visits` | — | Remove duplicate `CrawlerVisit` rows. |
| `fix_other_traps` | `--dry-run` | One-shot cleanup: dedupe → delete all `DashboardStat` rows → recompute from scratch. Holds the precalc lock to prevent cron overlap. |
| `analyze_other_traps` | `--limit`, `--sample` | Inspect `CrawlerVisit` rows with `trap_type='other'` — top paths, prefixes, UAs, IPs, bot types, hosts. |
| `generate_content_fixture` | — | Generate test fixtures for content. |
| `export_gen_data` | — | Export data for external image generation. |

### Dashboard Stats Setup (Production)

Add to host crontab to incrementally update stats every 30 minutes:

```
*/30 * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py precalc_dashboard >> /var/log/acpwb-precalc.log 2>&1
```

Stats are stored in `DashboardStat` DB rows. On the first run after deploy, all historical data is processed. Subsequent runs only process new rows (fast). The dashboard shows "Stats as of: [timestamp]" from the DB row's `updated_at`.

---

## Running Tests

```bash
docker compose exec web pytest          # all tests
docker compose exec web pytest -v       # verbose
docker compose exec web pytest tests/test_bot_classify.py  # specific file
```

- Config: `acpwb/pytest.ini` — `DJANGO_SETTINGS_MODULE = config.settings.local`, `django_debug_mode = true` (required for `?__year=YYYY` subdomain shortcut in archive tests)
- Test files: `acpwb/tests/test_*.py` (~11 files, ~250 tests)
- Fixtures: `acpwb/tests/conftest.py` — provides `client`, `bot_client`, `staff_client`, `mailgun_post`
