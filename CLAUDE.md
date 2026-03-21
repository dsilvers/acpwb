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
- Docker Compose (web + db + nginx)
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
| `apps/core` | `BotTrackingMiddleware`, context processors, `{% avatar_card %}` and `{% headshot_or_avatar %}` template tags, staff dashboard views |
| `apps/public` | Home, Careers, Mission, Partners, Privacy + `Fortune500Company` model |
| `apps/people` | Our People honeypot — generates 12 employees per load, logs visits |
| `apps/projects` | Infinite project list (PoW gated) + project detail |
| `apps/honeypot` | Archive trap, Wiki, Reports, Fake API, Well-Known files, Ghost traps, PoW endpoints |
| `apps/webhooks` | Inbound email receiver (Cloudflare pipe + Mailgun) + `HoneypotMatch` logic |

---

## Key Models

- `people.PeoplePageVisit` — every load of `/our-people/`
- `people.GeneratedEmployee` — the fake employees shown (FK → visit)
- `honeypot.CrawlerVisit` — all bot/trap activity (trap_type choices include `report_list`, `report_download`, `ghost_link`, `dataset`, `api`, `well_known`)
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
- `/archive/<year>/<month>/<day>/<path:slug>/` — infinite recursive archive, exponential link branching
- `/wiki/<slug>/` — subtly wrong watermarked facts (60+ topics, interconnected graph)
- `/reports/` — fake research archive 1993–present; infinite scroll; watermarked CSVs (300–800 rows real data) and PDF-style documents
- `/reports/<slug>/download.csv` — real downloadable CSV with per-slug watermark token in every row
- `/api/v1/private-data` — 200 JSON garbage with fake credentials (referenced in HTML comment, not nav)
- `/.well-known/ai-agent.json` — fake AI agent manifest with trap `allowed_actions`
- `/.well-known/robots.txt` — reverse-psychology (Disallow = more honeypot content, Crawl-delay: 0, points to trap sitemaps)
- `/sitemap.xml` — real Django sitemap (static pages + projects) for legitimate crawlers
- `/sitemap-publications.xml` — trap: reports, ghost traps, fake internal paths; logged as `well_known`
- `/sitemap-wiki.xml` — trap: all 75+ wiki topics; logged as `well_known`
- `/sitemap-archive.xml` — trap: 500 deterministic archive URLs (seed `0x4143505742`), 2008–2024; logged as `well_known`
- `/internal/` — fake intranet portal hub (indexed, Allow in robots.txt); shows IP-deterministic "Welcome back [name]", dashboard cards, announcements; logged as `ghost_link`
- `/internal/login/` — fake Okta/Azure AD SSO page; accepts any POST, logs credentials to `InternalLoginAttempt`, redirects to `?next=`
- `/internal/employee-records/` — paginated employee table (50 rows/page, infinite); Export CSV (500 rows, watermarked)
- `/internal/salary-database/` — salary band table by job family + level; Export CSV
- `/internal/acquisition-targets/` — M&A pipeline table with deal stages; Export CSV
- `/internal/litigation-hold/` — legal hold inventory (HTML only, no CSV export)
- `/archive/<year>/<month>/<day>/<path:slug>/export.csv` — watermarked CSV for archive entry (200–500 rows); logged as `archive`
- `/feeds/archive.xml` — Atom feed, 20 entries/page, infinite via `?page=N`; logged as `well_known`
- `/feeds/reports.xml` — RSS 2.0 feed of reports, 10/page, infinite; logged as `well_known`
- `/api/v1/openapi.json` — valid OpenAPI 3.0.3 spec with 20 fake endpoints, watermarked; logged as `api`
- `/datasets/` — index of 8 fake NLP/compensation datasets; logged as `dataset`
- `/datasets/<slug>/` — dataset detail with description, format example, citation
- `/datasets/<slug>/data.jsonl` — paginated JSONL (100 records/page), instruction-response pairs, watermarked; logged as `dataset`
- `/internal/portal/`, `/employees/export/`, `/admin-panel/login/` — ghost trap 403s, all logged

Staff dashboard:
- `/acpwb-dashboard/` — requires `is_staff`; overview + sub-views for crawlers, archive, email, page visits
- Views in `apps/core/dashboard_views.py`, URLs in `apps/core/dashboard_urls.py`
- Bot classification via `BOT_PATTERNS` / `classify_ua()` / `classify_ua_group()` in `dashboard_views.py`
- Date range controlled by `?range=` preset or `?from=`/`?to=` custom; `_daily_chart()` returns `{'bars': [...], 'start': '...', 'end': '...'}`
- Overview stat cards: Crawler Hits, Archive Visits, Inbound Emails, People Visits, Project Visits, Login Attempts (red)
- "By Trap Type" panel pulls from `CrawlerVisit.TRAP_CHOICES` dynamically — new trap types appear automatically

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
