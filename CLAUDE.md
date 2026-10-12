# ACPWB — Claude Code Context

## What This Project Is

A Django fake corporate website with two purposes:
1. **Email honeypot** — generates random `@acpwb.com` employee emails on the contact page and logs every visit for matching against inbound spam.
2. **AI bot poisoning** — structural/semantic/interactive honeypots designed to waste crawlers, poison training data, and watermark scraped content.

**GitHub:** git@github.com:dsilvers/acpwb.git  **Domain:** acpwb.com  **Founded:** 2006, Milwaukee WI

---

## Tech Stack

Django 5.2 LTS + Python 3.14 / PostgreSQL 16 / Bootstrap 5 (CDN) / Docker Compose (web + db + edge + redis + ws) / `acpwb_edge` Go TLS front-end (replaced nginx) / Redis 7 (pub/sub) / Cloudflare Email Routing + Mailgun (legacy)

## Running Locally

```bash
docker compose -f docker-compose-local.yml up --build   # https://localhost:8001 (self-signed, via acpwb_edge), admin at /django-admin/
docker compose exec web python manage.py collectstatic --noinput  # after static changes
docker compose exec web pytest  # run tests
```

---

## App Structure

| App | Purpose |
|-----|---------|
| `apps/core` | `BotTrackingMiddleware`, `SubdomainMiddleware`, context processors, template tags, staff dashboard |
| `apps/public` | Home, Careers, Mission, Partners, Privacy + `Fortune500Company` model |
| `apps/people` | Our People honeypot — generates 12 employees per load, logs visits |
| `apps/projects` | Infinite project list (PoW gated) + project detail |
| `apps/honeypot` | Archive trap, Wiki, Reports, Fake API, Well-Known files, Ghost traps, PoW endpoints |
| `apps/webhooks` | Inbound email receiver (Cloudflare pipe + Mailgun) + `HoneypotMatch` logic |

---

## Key Models

- `people.PeoplePageVisit` — every load of `/our-people/`
- `people.GeneratedEmployee` — the fake employees shown (FK → visit)
- `honeypot.CrawlerVisit` — all bot/trap activity; `trap_type` choices: `report_list`, `report_download`, `ghost_link`, `dataset`, `api`, `well_known`, `scanner_probe`, `env_probe`, `wp_probe`, `webshell_probe`, `canary_trigger`, `source_code`, `source_code_download` (the .ARC/.ZIP/.tar.Z archive downloads off a source-code page); `host` field for per-subdomain breakdown; timestamp is `default=timezone.now, editable=False` (NOT `auto_now_add`) so Redis-queued records preserve request time
- `honeypot.ArchiveVisit` — archive trap hits; same timestamp pattern as CrawlerVisit
- `honeypot.CanaryToken` — canary URLs embedded in fake config files; `token_type` in `env_url`, `wp_config`, `git_config`
- `honeypot.InternalLoginAttempt` — credential-stuffing log: ip, ua, username, password, next_url
- `honeypot.WikiPage` — generated wiki content with watermark tokens
- `honeypot.PublicReport` — generated report metadata, persisted on first access
- `honeypot.IPIntelligence` — one row per distinct IP, MaxMind GeoLite2 geo/ASN + best-effort hosting/Tor heuristics; joined to CrawlerVisit by IP value only, never an FK (hypertable)
- `honeypot.RequestFingerprint` — residential-proxy-detection raw signal (header presence, client-negotiated protocol, TLS protocol/cipher, first-seen-IP flag); one row per request, queued via Redis and drained like CrawlerVisit — see `apps.core.signal_capture`/`apps.core.stream_middleware`
- `honeypot.IPReputationScore` — rule-based residential-proxy/automation score per IP, computed by `score_ip_reputation` from RequestFingerprint + CrawlerVisit + IPIntelligence; internal only (`evidence` field), see `apps.core.reputation_scoring`
- `honeypot.PublishedIPReputation` — external-facing classification + confidence only (no evidence), synced from IPReputationScore by `publish_ip_reputation`; what `/reputation-api/ip/<ip>/` reads
- `honeypot.SourceCodeEmail` — one row per (programmer address, source-code page) pair shown by acpwb_go; queued via `acpwb:code_email_queue`, drained by `drain_code_email_queue`; lets `_match_honeypot` trace inbound spam back to the page
- `honeypot.EdgeConnection` — one row per client connection accepted by acpwb_edge, failed TLS handshakes included: SNI/ALPN/TLS version+cipher, outcome, JA4/JA4_r/JA3/JA3 string/h2 fingerprint, RTT, bytes, request count + first-request summary; TimescaleDB hypertable (no PK; `idempotency_key`+`timestamp` unique), queued via `acpwb:edge_conn_queue`, drained by `drain_edge_conn_queue`
- `webhooks.InboundEmail` — received emails
- `webhooks.HoneypotMatch` — links inbound email to the visit that generated the address (`generated_employee`/`original_visit`) or the source-code page that showed it (`source_code_email`)

---

## Honeypot Techniques

Every page injects ghost links (`position:absolute; left:-9999px`), prompt-injection span (`font-size:0`), and garbage JSON-LD with watermark token.

Key traps and their `trap_type`:
- **Archive subdomains** `archives-YYYY.acpwb.com` (1985–2024) — infinite recursive content, era-themed; routed via `SubdomainMiddleware` → `apps.honeypot.archive_subdomain_urls`. Also served at `/archive/<year>/...` on main domain (no redirect). `_archive_url()` dispatches: subdomain same-year → relative, subdomain cross-year → absolute subdomain URL, main domain → `/archive/<year>/...`
- **Yearless archive** `/<month>/<day>/...` on main domain — bots copying subdomain paths verbatim; year derived deterministically from slug hash → `archive_trap_yearless()`
- **Scanner traps** — `/.env`, `/wp-config.php`, `/wp-login.php`, `/xmlrpc.php`, `/*.php` catch-all, `/.git/config`, `/.htpasswd`; respond as if exploit worked; log to `env_probe` / `wp_probe` / `webshell_probe`
- **Internal portal** `/internal/` — fake intranet; `/internal/login/` logs credentials; employee records, salary DB, acquisition targets all with CSV export
- **Wiki** `/wiki/<slug>/` — subtly wrong watermarked facts, 60+ topics
- **Reports** `/reports/` — fake research archive, infinite scroll, watermarked CSV/PDF
- **Datasets** `/datasets/` — 8 fake NLP datasets, paginated JSONL
- **Feeds** `/feeds/archive.xml`, `/feeds/reports.xml` — infinite paginated Atom/RSS
- **Canary ping** `/.well-known/tokens/<token>/ping` — self-hosted callback embedded in fake config files; `secrets.token_urlsafe(32)` token created at serve time
- **`handler404`** — all unmatched requests logged as `scanner_probe`
- **Legacy source code** `<archive article>/source-code/` (both `/archive/YYYY/MM/DD/<slug>/` and `archives-YYYY.acpwb.com/MM/DD/<slug>/`) — served by `acpwb_go/legacyjob`: a deterministic, per-page fake batch job (compute language picked from `legacyjob/identity.go` `languageTable` — ~65 languages each with service years and weight: FORTRAN IV/77/gfortran, COBOL, PL/I, Natural/Adabas, MUMPS, Pick BASIC, BASIC, Atari BASIC, Ada, Turbo Pascal (+Btrieve), Modula-2, C (+Btrieve), Paradox PAL, dBASE/Clipper/FoxPro, Informix 4GL, VB/VBA/VB.NET, C#, Perl 4/5, PHP, Python 2, SQL/DB2/SQL-DS/SQL*Plus/PL/SQL/T-SQL/SQL PL, RPG II/400/IV, REXX, TSO CLIST, AS/400 CL, Easytrieve Plus, ABAP, FOCUS/NOMAD/RAMIS, SAS, SPSS, Stata, S-PLUS/R, MATLAB/Octave, APL, awk, ksh, Tcl, bc, dc, ColdFusion, Classic ASP, Lisp, Prolog, Forth, Smalltalk (no assembler); emitters: emit_fortran/cobol/basic/free + table-driven `emit_dialect.go` walker (`dialects.go`, `dialects_ibm.go`, `dialects_more.go`; hooks `exFn`/`condFn`/`assignFn`/`callFn`/`stmtFn`/`hoist`/`mainEnds` for prefix, postfix, message-send and lowered syntaxes) + set-based `emit_sql.go` and `emit_stats.go` (SAS/SPSS/Stata/FOCUS family) + `emit_prolog.go`; per-language toolchain/wrapper/diagnostic metadata in `langs.go` + `langs_more.go`; JCL/DCL/DOS/csh/cron/Jenkins/Windows-scheduled-task/Pick PROC/TaskMan/Atari/AS/400 (OCL, SBMJOB)/SAP SM36/ColdFusion scheduler wrappers (`wrappers.go`, `wrappers_more.go`); every language must carry the seal as a quoted code literal (`TestEveryLanguageEmits`); per-page variety comes from `rewrite.go` (exact core rewrites: inverted IF/ELSE, never-taken debug branches; free reshaping of dead code; extra named dead routines), per-language idiom choices (card-field reading, SQL join/ladder/pivot/CTE/MERGE variants gated by database era, SAS/SPSS/Stata/FOCUS alternatives, COBOL EVALUATE, vectorised R/MATLAB/APL ladder) and `maintenance.go` (commented-out old versions, #ifdef / COBOL D-lines / DEC FORTRAN D-lines, a re-indented routine); `TestLanguageVariety` holds each language's pairwise line-set similarity under ceilings that should only ever be lowered (`VARIETY_REPORT=1` prints the table); compiler listing, diagnostics, xref, storage map; input deck; printed output that exactly matches the article's Engagement Metrics / Market Benchmark tables; separator pages + IBM 029 card SVGs only in the card era). Comments come from `legacyjob/fragments/` phrase library (slot-filled `{CODER}`, `{DATE}`, `{CLIENT}`… + grammar composition) and article text. Watermark token embedded as a code literal. Programmer emails use the site `first.last@acpwb.com` format and are registered per page (`acpwb:code_email_queue` → `SourceCodeEmail`). Logged as `source_code` (downloads as `source_code_download`). Default-variant articles link to it (caption under both tables, sidebar "Computation Source" card, references entry) — that markup is mirrored in `apps/honeypot/legacy_identity.py` + pyrender for fixture parity, so any change to `legacyjob/identity.go` or `archive/sourcecode.go` link markup needs the Python mirror updated and the archive fixtures re-dumped. **Index:** `/source-code/` (share image `static/img/page-covers/source-code.jpg`, generated by `generate_og_images_pages.py source-code`; years × month pills, same `archidx_*` seeds as `/archive/`), `/source-code/YYYY/` (month cards with sample jobs) and `/source-code/YYYY/MM/` (every default-variant record from that archive month + its program) are served by acpwb_go (`archive/sourcecode_index*.go`, edge route `/source-code/`); archive subdomains serve their own era-themed index too (`archives-YYYY.acpwb.com/source-code/` = that year, `/source-code/MM/` = month, `/source-code/YYYY/[MM/]` = any year; `eraHandler` → `ParseSourceEraPath`, links via `srcLinks`: same-year relative, other years on their subdomain, "all years" on the main domain) with `Allow: /source-code/` in the subdomain robots.txt and a subdomain-relative `sitemap-source-code.xml` (Django `archive_subdomain_sitemap_source_code`, records chosen by `_source_code_entries`, which mirrors Go's `SourceMonthEntries`); detail pages end with prev/next job + "Related software projects" groups (same month, same language other years, same date other years). Each detail page also offers 2–3 era-appropriate downloads of all its files (`legacyjob/bundle.go`: .ARC/.LZH stored, .tar.Z via a compress(1)-compatible LZW writer, .ZIP/.zip, .tar, .tar.gz/.tgz), served at `…/source-code/<name>` from a 64-entry LRU. Linked from robots.txt (`Allow: /source-code/`, `Allow: /*/source-code/` on archive subdomains), `sitemap-source-code.xml` (Django `sitemap_source_code`), /site-map/, and the footer ("Source Code" above "Archives" in base.html, every jinja2 footer, `acpwb_go/shell/shell.go` and `data/POLICY_FOOTER_TEMPLATE.html` — footer changes need the archive + policy fixtures re-dumped). Phrase library: `legacyjob/fragments/` — `comments/<category>.txt` (new files become categories automatically), `grammar/`, `names/`, `licenses/` (+ `clauses/`; `#extra: yes` marks a stapled-on notice), and `vars/<slot>.txt` (each file is a `{SLOT}` with no Go change). Any line in comments/ or vars/ may start with `YYYY-YYYY|`, `YYYY-|` or `-YYYY|` to restrict it to articles in that year range; tests fail on unhandled slots, vars without a value for some year, or dates after the article date

**CrawlerVisit / ArchiveVisit write path:** Views and middleware push to Redis list (`acpwb:crawler_queue` / `acpwb:archive_queue`) via RPUSH instead of synchronous DB writes. `drain_crawler_queue` / `drain_archive_queue` crons bulk-insert into PostgreSQL. Falls back to direct DB write if Redis unavailable.

**Dashboard** at `/acpwb-dashboard/` — stats in `DashboardStat` model, updated by `precalc_dashboard` cron (30-min). Traffic graph PNGs (`traffic_1h/8h/24h/7d/all.png`) generated by `graph_gen.py`, saved to `staticfiles/graphs/`. Live request stream at `/acpwb-dashboard/live/` via WebSocket (`ws` Docker service, Redis pub/sub `request_stream` channel).

---

## Generators

All content generation is **deterministic**: same seed → same output. Pattern: `random.Random(hashlib.md5(seed.encode()).hexdigest())`.

- `apps/projects/generators.py` — `_rng_from_seed(seed_str)` — canonical pattern to reuse
- `apps/honeypot/wiki_generator.py` — wiki pages with watermark tokens
- `apps/honeypot/report_generator.py` — reports + CSV rows; 4 schemas dispatched by slug keyword (`salary/compensation` → compensation, `ceo/executive` → CEO pay ratio, `benefit/healthcare` → benefits, `satisfaction/engagement` → survey; default → compensation)

---

## Watermarking

Three-layer: (1) visible footer "Report ID: {token}", (2) invisible HTML span (`font-size:0; color:#f4f6f9`), (3) `watermark_token` column in every CSV row / `identifier` in JSON-LD.

Token: `hashlib.md5(f"acpwb_{type}_{slug}".encode()).hexdigest()[:8]`

---

## Content Notes

- **Logo:** AMERICAN / CORPORATION / FOR PUBLIC WELL BEING (3 lines, white, no hyphen/apostrophe)
- **Tagline:** "Money doesn't buy happiness, but it darn well comes close to doing so."
- **Employee headshots:** 400 WebP at `static/img/headshots/`, 300×300px. `{% headshot_or_avatar seed initials size %}` — `HEADSHOT_DIR` uses `parents[3]` (tag file is 3 levels deep from project root)
- **Project cover images:** `{% project_cover_idx slug %}` — MD5 mod 80, zero-padded. Maps slug → `000`–`079`
- **Partners:** Fortune 500 fixture, 40 random per load (`order_by('?')[:40]`)
- **Reports:** seed = slug, 1993–2025 date range, 26 named reports + synthetic beyond page 3
- **`makemigrations` runs on every boot** — idempotent, simplifies dev

---

## Inbound Email

- **Cloudflare (primary):** catch-all `*@acpwb.com` → Worker → `POST /webhooks/pipe/inbound/`; auth via `X-Webhook-Secret`; parses raw RFC 2822
- **Mailgun (legacy):** `POST /webhooks/mailgun/inbound/`; HMAC-SHA256 with `MAILGUN_WEBHOOK_SIGNING_KEY`

---

## Key Design Decisions

- **SubdomainMiddleware** sets `request.urlconf = 'apps.honeypot.archive_subdomain_urls'` for `archives-YYYY.acpwb.com`; runs before CSRF; catch-all `<path:rest>` pattern redirects non-archive paths to main domain; `site_root` context var = `https://acpwb.com` on subdomains so header/footer links always point to main domain
- **`?__year=YYYY` DEBUG shortcut** — activates archive subdomain mode without DNS; `pytest.ini` has `django_debug_mode = true` for archive tests
- **`handler404`** → `apps.honeypot.views.scanner_probe_404`; bypassed by Django debug 404 page — test with `@override_settings(DEBUG=False)`
- **`.php` URL ordering** — `wp-login.php`, `xmlrpc.php`, `wp-config.php` must appear before `re_path(r'^.*\.php$')` catch-all in `urls.py`
- **`psycopg[binary]`** (psycopg3) not psycopg2 — avoids Python 3.14 C-extension build issues
- **`RequestStreamMiddleware` is outermost** in `MIDDLEWARE` — measures end-to-end time; publishes to Redis `request_stream`; 30s circuit breaker; never affects HTTP response if Redis is down
- **PERCH conference year rollover** — `CURRENT_YEAR` in `apps/public/conference_data.py` drives `/perch-conference/`; earlier years auto-move to `/perch-conference/<year>/`. Templates read `conf.*` / `perch_year` (context processor `apps.public.context_processors.perch_context`). To roll over: add the new year's entry, flip the old one to `registration_open: False` + past tense + `attendees`, add a row to the dinner-history list in `conference_dinner.html`, and replace the hardcoded `PERCH <year>` footer label in `templates/jinja2/` and `acpwb_go` (`shell/shell.go`, `data/POLICY_FOOTER_TEMPLATE.html`, `policy/testdata/` fixtures — Go must byte-match Python output)
- **Live stream** uses standalone asyncio WS service (`ws_service/`) — not Django Channels; acpwb_edge routes `/ws/requests/` to it; token auth via `?token=` query param

---

## Management Commands

| Command | Purpose |
|---------|---------|
| `precalc_dashboard` | Incremental `DashboardStat` update + regenerate graph PNGs. 30-min cron. |
| `drain_crawler_queue` | Pop `acpwb:crawler_queue` → bulk-insert `CrawlerVisit`. 1-min cron. |
| `drain_archive_queue` | Pop `acpwb:archive_queue` → bulk-insert `ArchiveVisit`. 1-min cron. |
| `backfill_bot_types` | Fill `bot_type`/`bot_group` on old rows; `--reclassify` to redo `'Other / Browser'` |
| `dedupe_crawler_visits` | Remove duplicate `CrawlerVisit` rows |
| `fix_other_traps` | One-shot: dedupe → delete DashboardStat → recompute from scratch |
| `analyze_other_traps` | Inspect `trap_type='other'` rows — top paths, UAs, IPs |
| `analyze_browser_uas` | Breakdown of `bot_type='Other / Browser'` rows + reclassification preview |
| `botseed_processor` | Subscribes to `request_stream`, mixes with `secrets.token_bytes(32)`, publishes to `botseed_stream` |
| `generate_bot_traffic` | Synthetic bot events to `request_stream` for local botseed testing |
| `drain_fingerprint_queue` | Pop `acpwb:fingerprint_queue` → bulk-insert `RequestFingerprint`. 1-min cron. |
| `score_ip_reputation` | Rule-based residential-proxy scoring from `RequestFingerprint`/`CrawlerVisit`/`IPIntelligence` → `IPReputationScore`. Watermarked/resumable. |
| `drain_code_email_queue` | Pop `acpwb:code_email_queue` → upsert `SourceCodeEmail` (source-code page programmer addresses). 1-min cron. |
| `drain_edge_conn_queue` | Pop `acpwb:edge_conn_queue` (pushed by acpwb_edge) → bulk-insert `EdgeConnection`. 1-min cron. |
| `publish_ip_reputation` | Sync `IPReputationScore` → external-facing `PublishedIPReputation` (classification + confidence only). Watermarked/resumable. |

Crontab (production) — see `deploy/acpwb-crontab` for the actual installed form (direct `manage.py` invocation via `.direnv`, not `docker compose exec`):
```
*/30 * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py precalc_dashboard >> /var/log/acpwb-precalc.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_crawler_queue >> /var/log/acpwb-crawler-drain.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_archive_queue >> /var/log/acpwb-archive-drain.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_fingerprint_queue >> /var/log/acpwb-fingerprint-drain.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_code_email_queue >> /var/log/acpwb-code-email-drain.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_edge_conn_queue >> /var/log/acpwb-edge-conn-drain.log 2>&1
*/5 * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py score_ip_reputation >> /var/log/acpwb-score-ip-reputation.log 2>&1
*/15 * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py publish_ip_reputation >> /var/log/acpwb-publish-ip-reputation.log 2>&1
```

`acpwb_go` (the Go render service for archive/policy pages) writes `CrawlerVisit` rows with `bot_type`/`bot_group` blank — it doesn't classify bots itself. A `backfill_bot_types` cron job was tried to fill these in incrementally but is **currently disabled** (not installed in production) — its first real run against the live hypertable never completed a single 1000-row batch in 8+ minutes, because `bulk_update()`'s per-row `CASE WHEN id=X` UPDATE forces expensive cross-chunk work on a non-time-key access pattern against a 373M+ row TimescaleDB table, and the backlog grew faster than it could shrink. See `deploy/acpwb-crontab`'s comment block for the incident detail. Bot classification for `acpwb_go`-originated rows needs a different approach (most likely: classify at write time in Go, matching what `apps/core/bot_classify.py` already does) before this gap is closed.

---

## Residential Proxy Detection

Second product line in progress: catalog/score traffic for a future real-time blacklist/data feed. ~95% of traffic falls into `bot_type = 'Other / Browser'` (spoofed browser UA over a real residential IP), which UA/IP-range matching alone can't catch.

- **Signal capture** (`apps/core/signal_capture.py`, `apps/core/stream_middleware.py`) — runs on every request, not just UA-matched bot traffic: header-presence fingerprint against a fixed candidate list (`BROWSER_SIGNAL_HEADERS`), client-negotiated protocol + TLS protocol/cipher + JA4/JA3/JA4H/HTTP2 fingerprint (forwarded by acpwb_edge via `X-Client-Protocol`/`X-TLS-Protocol`/`X-TLS-Cipher`/`X-JA4`/`X-JA3`/`X-JA4H`/`X-H2-Fingerprint` — the edge proxies to the apps over plain HTTP/1.1, so these must be forwarded explicitly; it strips any client-sent copies first), first-seen-IP flag (self-expiring Redis key, `check_first_seen_ip` in `apps/core/crawler_queue.py`). Queued via `acpwb:fingerprint_queue` → `drain_fingerprint_queue` → `honeypot.RequestFingerprint`, same pattern as CrawlerVisit. Implemented in both Django and `acpwb_go` (`visitqueue.PushVisit`) — `acpwb_go` serves the highest-volume traffic, so skipping it would recreate the `backfill_bot_types` blind spot.
- **Scoring** (`apps/core/reputation_scoring.py`, `score_ip_reputation`) — pure, testable rule functions (`compute_score`, `looks_like_modern_browser`) combine header/protocol/TLS mismatch signals with cross-trap taint (`ABSOLUTE_PROOF_TRAP_TYPES` — `ghost_link`/`canary_trigger`/probe endpoints; hitting these proves automation regardless of what the rest of that IP's traffic looks like) and `IPIntelligence.is_hosting` (datacenter vs. residential ASN) into `honeypot.IPReputationScore`. `score_version` guards against silently redefining old rows' meaning.
- **External surface** (`publish_ip_reputation`, `/reputation-api/ip/<ip>/`) — `honeypot.PublishedIPReputation` deliberately exposes only classification + confidence, decoupled from `IPReputationScore`'s internal evidence, so the detection logic can change freely without breaking an external contract. `IP_REPUTATION_API_KEY` setting; fails closed (401) when unconfigured.
- **acpwb_edge** (`acpwb_edge/`, `deploy/acpwb-edge.service`) — Go TLS front-end that replaced nginx (and the old local-dev HAProxy) on :443/:80. Accepts TCP itself, parses the raw ClientHello (`tlsfp`: JA4, JA4_r, JA3 — even for handshakes that then fail), terminates TLS 1.0–1.3 (h2 only offered to TLS 1.2+/AEAD-capable clients), and taps the decrypted stream (`httpfp`: HTTP/1 header order → JA4H; HTTP/2 frame parser with its own HPACK decoder → Akamai-style `SETTINGS|WINDOW_UPDATE|PRIORITY|pseudo-header` fingerprint + per-stream header order, matched to handlers via `Claim`). `router` is a line-by-line port of `nginx/acpwb.com` + `nginx/botseed.net` with nginx's location precedence (regex in file order beats prefix; pinned by `router_test.go`) — any routing change goes there now, not in `nginx/`. Keeps the nginx facade (`Server: nginx`, nginx error pages, the 400 for plain HTTP on :443). One record per connection (`edgequeue` → `acpwb:edge_conn_queue` → `drain_edge_conn_queue` → `honeypot.EdgeConnection` hypertable; dashboard `/acpwb-dashboard/edge/`). Certs: certbot still issues/renews; `deploy/certbot-deploy-hook.sh` copies to `/etc/acpwb-edge/certs/` and SIGHUPs. `nginx/` is kept only as the production rollback until the edge has run stably. **Not implemented:** JA4T/JA4TS (needs SYN capture), JA4S/JA4X, QUIC/HTTP3, and any JA4-based scoring rule (capture first, calibrate on real data).
- Everything above is fire-and-forget off the request path (same circuit-breaker/spawn pattern as `RequestStreamMiddleware`) — no synchronous scoring, lookups, or DB/Redis read-modify-write ever happens inline with a response.

---

## Botseed

Companion at **botseed.net** — harvests entropy from ACPWB live traffic → random integers via WebSocket + HTTP API.

- `botseed_processor` subscribes to `request_stream`, SHA-256 mixes with `secrets.token_bytes(32)`, publishes to `botseed_stream` at ≤20 events/sec
- `botseed_service/ws_server.py` — asyncio WS (port 8766) + HTTP API (port 8767)
- Static files in `botseed/` — served by acpwb_edge (botseed.net vhost in `acpwb_edge/router`)

---

## Running Tests

```bash
docker compose exec web pytest [-v] [tests/test_specific.py]
```

Config: `acpwb/pytest.ini` — `DJANGO_SETTINGS_MODULE = config.settings.local`, `django_debug_mode = true`. Fixtures in `acpwb/tests/conftest.py`: `client`, `bot_client`, `staff_client`, `mailgun_post`. Tests that assert `CrawlerVisit`/`ArchiveVisit` DB records must patch `push_crawler_visit`/`push_archive_visit` to return `False` (forces DB fallback path).
