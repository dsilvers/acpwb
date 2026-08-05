# ACPWB production deploy (no Docker)

Assumes project checked out to `/home/acpwb/acpwb`, run as system user `acpwb`,
with Postgres 16 + TimescaleDB and Redis installed locally on the box.

## Install units

```bash
sudo cp deploy/acpwb-*.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now acpwb-gunicorn acpwb-ws acpwb-botseed-ws acpwb-botseed-processor
```

## Install cron jobs

```bash
sudo crontab -u acpwb deploy/acpwb-crontab
```

Make sure `/var/log/acpwb-*.log` are writable by `acpwb` before the first run.

## Before starting

- `.env` at `/home/acpwb/acpwb/.env` must exist, with `DB_HOST=127.0.0.1` (not
  `db` — that was the Docker Compose service name) and `DJANGO_SETTINGS_MODULE`
  pointed at the production settings module.
- `acpwb/venv`, `ws_service/venv`, `botseed_service/venv` must exist with their
  respective dependencies installed (see each service's `Dockerfile` for the
  package list, or point everything at one shared venv with all deps merged in).
