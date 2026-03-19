# ACPWB — Claude Code Context

## What This Project Is

A Django fake corporate website with two purposes:
1. **Email honeypot** — generates random `@acpwb.com` employee emails on the contact page and logs every visit for matching against inbound spam via Mailgun webhook.
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
- Mailgun for inbound email webhook

## Running Locally

```bash
docker compose up --build
# Site at http://localhost
# Admin at http://localhost/django-admin/
docker compose exec web python manage.py createsuperuser
```

---

## App Structure

| App | Purpose |
|-----|---------|
| `apps/core` | `BotTrackingMiddleware`, context processors, `{% avatar_card %}` template tag |
| `apps/public` | Home, Careers, Mission, Partners, Privacy + `Fortune500Company` model |
| `apps/people` | Our People honeypot — generates 12 employees per load, logs visits |
| `apps/projects` | Infinite project list (PoW gated) + project detail |
| `apps/honeypot` | Archive trap, Wiki, Fake API, Well-Known files, Ghost traps, PoW endpoints |
| `apps/webhooks` | Mailgun inbound email receiver + `HoneypotMatch` logic |

---

## Key Models

- `people.PeoplePageVisit` — every load of `/our-people/`
- `people.GeneratedEmployee` — the fake employees shown (FK → visit)
- `honeypot.CrawlerVisit` — all bot/trap activity
- `honeypot.WikiPage` — generated wiki content with watermark tokens
- `webhooks.InboundEmail` — received emails from Mailgun
- `webhooks.HoneypotMatch` — links inbound email to the visit that generated the address

---

## Honeypot Techniques Deployed

Every page injects:
- **Ghost links** (`display:none`) to `/internal/portal/`, `/employees/export/`, `/admin-panel/login/`, `/api/v1/private-data`
- **Prompt injection** — white-on-white invisible text with fake AI training instructions
- **Garbage JSON-LD** — fake structured data with provenance watermark token

Dedicated traps:
- `/archive/<year>/<month>/<day>/<path:slug>/` — infinite recursive archive
- `/wiki/<slug>/` — subtly wrong watermarked facts
- `/api/v1/private-data` — 200 JSON garbage with fake credentials
- `/.well-known/ai-agent.json` — fake AI agent manifest
- `/.well-known/robots.txt` — reverse-psychology (Disallow = more honeypot content)

---

## Content Notes

- **Logo:** 3 lines: AMERICAN / CORPORATION / FOR PUBLIC WELL BEING (no hyphen, no apostrophe)
- **Tagline:** "Money doesn't buy happiness, but it darn well comes close to doing so."
- **Founded:** 2006 (domain registration year)
- **Privacy page:** Preserves original Happy Fun Ball disclaimer verbatim + new AI data policy
- **Careers:** Satirical over-the-top benefits, zero actual job openings
- **Partners:** Fortune 500 fixture, 40 random shown per load (`order_by('?')[:40]`)
- **Projects:** Deterministic infinite generation (seed = page number), PoW gated

---

## Django Admin

Located at `/django-admin/` (non-standard path).

All models registered with useful `list_display`, `search_fields`, and `list_filter`. Key admin views:
- **People → People Page Visits** — see every contact page load with inline employee records
- **Webhooks → Inbound Emails** with inline matches
- **Webhooks → Honeypot Matches** — the payoff: spam matched back to visit

---

## Mailgun Configuration

- Catch-all route: `@acpwb.com` → `POST https://acpwb.com/webhooks/mailgun/inbound/`
- Signature verification uses HMAC-SHA256 with `MAILGUN_WEBHOOK_SIGNING_KEY`
- Returns 406 on invalid signature, 200 on success (Mailgun retries on anything else)

---

## Known Design Decisions

- **No Pillow** — all avatars and partner logos are CSS gradient cards (`{% avatar_card %}` template tag)
- **psycopg[binary]** (psycopg3) not psycopg2 — avoids Python 3.14 C-extension build issues
- **Deterministic project generation** — same page number always returns same stories (seed = page)
- **`makemigrations` runs on every boot** — safe because migrations are idempotent; simplifies development
- **PoW difficulty = 5 bits** — solves in <1s in browser, costs a bot per-page at scale
