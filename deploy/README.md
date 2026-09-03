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

### Listen backlog / SYN queue (2026-09-03)

`net.core.somaxconn` alone does nothing for nginx unless the `listen`
directive's `backlog=` is set to match — nginx defaults to 511 regardless of
the sysctl value. This box was hitting that default under bot/scanner traffic:
`netstat -s` showed ~9.3M listen-queue overflows and ~9.3M dropped SYNs over
~1.75 days uptime (~60/sec sustained), which also produced sustained swap
pressure. FD limits and `worker_connections`/`worker_rlimit_nofile` were
already generous and were not the bottleneck.

Fix, applied 2026-09-03:
- `nginx/acpwb.com` — `backlog=65535` added to the host nginx `listen 80` /
  `listen 443 ssl` directives (and their `[::]` counterparts). Set **once**
  per unique `address:port` — nginx errors on "duplicate listen options" if
  `backlog=` is repeated on a shared socket, even with an identical value, so
  it's only present on the first `listen` block per socket in this file (see
  the comment above the second `:443` block).
- `deploy/99-acpwb-netstack.conf` — `net.core.somaxconn` and
  `net.ipv4.tcp_max_syn_backlog` raised from their prior values (8192 / 4096)
  to 65535 to match. Install with:
  ```bash
  sudo cp deploy/99-acpwb-netstack.conf /etc/sysctl.d/
  sudo sysctl -p /etc/sysctl.d/99-acpwb-netstack.conf
  ```
- Apply the nginx side with `sudo nginx -t && sudo systemctl reload nginx`
  (reload, not restart — zero downtime).

`botseed.net` and the Debian `default` site config were left untouched: they
share the same `0.0.0.0:80`/`:443` sockets as `acpwb.com` (which loads first
alphabetically from `sites-enabled/`), so the single `backlog=65535` on
`acpwb.com` already governs those sockets.

### Other limits checked post-fix (2026-09-03) — headroom is fine, one watch item

After the backlog fix, swap dropped from 8GB/8GB full to ~230MB and the
overflow/drop counters above stopped climbing. Also swept: conntrack (23%
of `nf_conntrack_max`), Redis clients (26% of `maxclients 10000`), TIME_WAIT
sockets (24% of `tcp_max_tw_buckets`, `tcp_tw_reuse` already on), `fs.file-max`,
and disk (56% used, 307G free). All comfortable.

**Watch item: Postgres `max_connections` (100) has no pooling in front of
it.** `DATABASES` in `acpwb/config/settings/base.py` doesn't set
`CONN_MAX_AGE` (defaults to 0 — a fresh connection per request, no
PgBouncer). Only ~10 connections are active under normal load today, because
most honeypot traffic writes go through the Redis queue (`push_crawler_visit`
/ `push_archive_visit`) rather than hitting Postgres directly. But gunicorn
runs 33 gevent workers × `--worker-connections 1000`, so a burst of
concurrent DB-touching requests (dashboard views, report/CSV generation,
admin) could in theory push past 100 concurrent connections before nginx or
the kernel show any strain. Not worth fixing preemptively — but if
`FATAL: sorry, too many clients already` shows up in the Django/Postgres
logs correlating with a traffic spike, the fix is either raising
`max_connections` (memory tradeoff against TimescaleDB's `shared_buffers`)
or adding PgBouncer in front.

### Incident: Postgres connection exhaustion → PgBouncer (2026-09-03)

The watch item above stopped being theoretical a few hours later. Postgres
hit `max_connections` (100) and started rejecting everything, including
superuser connections — `sudo -u postgres psql` itself intermittently failed
with `FATAL: sorry, too many clients already` (connections churned fast
enough that retrying a few times usually got in). Gunicorn's unix socket
backed up in sympathy (`connect() ... failed (11: Resource temporarily
unavailable)`), and the site started 502/500ing.

**Contributing factors:**
- `manage.py hourly_traffic_report` (`apps/core/management/commands/hourly_traffic_report.py`)
  was run manually/ad hoc — it's on no cron or systemd timer, repo or live.
  Its 24h `TruncHour` + `Count(distinct=True)` group-by over `CrawlerVisit`
  (a 373M-row TimescaleDB hypertable at the time) is exactly the shape that
  makes Postgres reach for parallel workers: 8-10 backend PIDs held
  connections active for 24+ minutes each on a single invocation.
- Underlying cause was still the one flagged above: no connection pooling,
  so ordinary request concurrency (33 gevent workers ×
  `--worker-connections 1000`) was enough on its own to keep the connection
  count pinned near 100 once a traffic burst hit — the `hourly_traffic_report`
  run made it worse, but cancelling those backends alone (`pg_cancel_backend`)
  didn't recover the site because gunicorn's own retry storm refilled every
  freed slot within the same second.

**Fix — PgBouncer in transaction-pooling mode**, installed same day:
- `deploy/pgbouncer.ini` → `/etc/pgbouncer/pgbouncer.ini`. `pool_mode =
  transaction`, `default_pool_size = 30` + `reserve_pool_size = 10` (caps
  real backend connections to Postgres at 40, well under `max_connections
  100`, leaving headroom for `psql`/cron/superuser), `server_reset_query =
  DISCARD ALL` (resets session state — e.g. Django's per-connection `SET
  TIME ZONE` — when a server connection returns to the pool between
  transactions; required for transaction-pooling correctness).
  ```bash
  sudo apt-get install pgbouncer
  sudo cp deploy/pgbouncer.ini /etc/pgbouncer/pgbouncer.ini
  sudo chown postgres:postgres /etc/pgbouncer/pgbouncer.ini
  ```
- **Gotcha:** the client-connection-limit parameter is `max_client_conn`
  (no trailing `s`) — `max_client_conns` fails config load with `unknown
  parameter` and pgbouncer won't start at all. Not obvious from the error
  message, which prints the section-qualified name (`pgbouncer/max_client_conns`).
- `/etc/pgbouncer/userlist.txt` — **not tracked in this repo** (see the
  `cloudflare.ini` precedent above for wildcard TLS). Populate it with the
  `acpwb` role's existing SCRAM hash straight from Postgres — no need to
  handle the plaintext password at all:
  ```bash
  sudo -u postgres psql -tAc "SELECT usename, passwd FROM pg_shadow WHERE usename='acpwb';"
  # then, in /etc/pgbouncer/userlist.txt:
  # "acpwb" "SCRAM-SHA-256$....."
  sudo chown postgres:postgres /etc/pgbouncer/userlist.txt
  sudo chmod 640 /etc/pgbouncer/userlist.txt
  ```
- `.env` — point Django (and every cron job, since they all `source .env`)
  at the pooler instead of Postgres directly:
  ```
  DB_HOST=127.0.0.1
  DB_PORT=6432
  ```
- Cutover sequence used: `systemctl stop acpwb-gunicorn` (this alone dropped
  the connection count from 100/pinned to 8, confirming the app itself was
  the load, not something external) → update `.env` → `systemctl start
  pgbouncer` → `systemctl start acpwb-gunicorn`. A `postgresql` restart
  was considered (to force-clear every stuck backend at once) but turned out
  unnecessary once gunicorn was stopped first.

Post-fix, Postgres's direct connection count settled at 67 and held steady
(down from pinned at 100), all `pg_hba.conf` auth for the pooler reuses the
existing `host ... scram-sha-256` entries since PgBouncer connects to
Postgres as a normal TCP client on `127.0.0.1:5432` — no `pg_hba.conf`
changes were needed.
