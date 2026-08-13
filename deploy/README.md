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
- On boxes provisioned with `direnv`/`layout pyenv`, the venv direnv creates at
  `.direnv/python-<version>/` is the one all units should point at — that's
  what these unit files assume. It won't have `websockets` installed by
  default (only needed by `ws_service`/`botseed_service`):
  ```bash
  /home/acpwb/acpwb/.direnv/python-3.14/bin/pip install "websockets>=14.0"
  ```
  If a box instead uses per-service venvs (e.g. `acpwb/venv`,
  `ws_service/venv`, `botseed_service/venv`), update the `ExecStart` paths in
  each unit file to match.
- Any directory in the path up to `staticfiles/` and `templates/` needs
  traverse (`x`) permission for both the `acpwb` owner and `www-data` (nginx).
  A `chmod` that leaves the owner without `+x` on `/home/<user>` itself is a
  silent trap: root-run checks won't notice (root bypasses permission bits),
  but gunicorn (running as `acpwb`) will get `PermissionError` trying to open
  files under it — including its own `templates/500.html` fallback, which
  turns an ordinary missing-template case into a worker crash and shows up as
  intermittent nginx 502s. Sanity check with `namei -l <path-to-a-static-file>`
  as a non-root user, not as root.

## nginx tuning (main `nginx.conf`, not tracked in this repo)

The stock Debian `nginx.conf` defaults (`worker_connections 768`, no
`worker_rlimit_nofile`, one TCP connection to gunicorn per request) fall over
under real traffic. On the production box these were bumped to:

- `worker_rlimit_nofile 100000;` (top-level, matches `systemctl show nginx -p
  LimitNOFILE`, which must be >= this value)
- `events { worker_connections 16384; multi_accept on; }`
- Full `gzip_*` settings (comp level 4, min length 1024, standard text/JS/XML
  types)
- `nginx/conf.d/upstream-acpwb.conf` — a `django_backend` upstream with
  `keepalive 128` — and `nginx/acpwb.com`'s `proxy_pass` directives point at
  `http://django_backend` instead of a raw `127.0.0.1:8000`, so nginx reuses
  persistent connections to gunicorn instead of opening a new one per request.
  Install with:
  ```bash
  sudo cp nginx/conf.d/upstream-acpwb.conf /etc/nginx/conf.d/
  ```
- `/ws/requests/` must point at the standalone `ws_service` (`127.0.0.1:8765`),
  **not** gunicorn — this was wrong in the config inherited from the old
  Docker-based `dan` box and silently broke the live dashboard stream.

`nginx.conf` itself isn't tracked here since it's the stock package file with
inline edits — diff against `/etc/nginx/nginx.conf.bak.*` on the box (created
before each edit) if you need the exact before/after.
