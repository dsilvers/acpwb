# American Corporation for Public Well Being

**acpwb.com** — Advancing American Prosperity Since 2006.

A Django-based fake corporate website with two operational purposes:
1. **Classic honeypot** — generates random `firstname.lastname@acpwb.com` email addresses on the contact page and logs every visit, enabling matching of inbound spam back to the exact page load that displayed the address.
2. **AI bot poisoning** — structural, semantic, and interactive honeypots designed to waste AI crawler resources, inject garbage into training pipelines, and watermark content for scraping provenance detection.

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

## Mailgun Setup

1. Add a catch-all route in Mailgun for `@acpwb.com` → forward to `https://acpwb.com/webhooks/mailgun/inbound/`
2. Set `MAILGUN_WEBHOOK_SIGNING_KEY` in your `.env` (found in Mailgun dashboard → Webhooks)

The webhook:
- Verifies HMAC-SHA256 signature
- Creates an `InboundEmail` record
- Queries `GeneratedEmployee` to find matches on the recipient address
- Creates `HoneypotMatch` records linking the spam to the exact page visit that displayed the address

View matches in Django admin under **Webhooks → Honeypot Matches**.

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DJANGO_SECRET_KEY` | Django secret key (required) |
| `DJANGO_DEBUG` | `True` for local dev |
| `DJANGO_SETTINGS_MODULE` | `config.settings.local` for dev |
| `DB_PASSWORD` | PostgreSQL password |
| `MAILGUN_WEBHOOK_SIGNING_KEY` | From Mailgun dashboard |
| `MAILGUN_DOMAIN` | `acpwb.com` |

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
