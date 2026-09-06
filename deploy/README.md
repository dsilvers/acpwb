# ACPWB production deploy (no Docker)

Assumes project checked out to `/home/acpwb/acpwb`, run as system user `acpwb`,
with Postgres 16 + TimescaleDB and Redis installed locally on the box.

## Install units

```bash
sudo cp deploy/acpwb-*.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now acpwb-gunicorn acpwb-ws acpwb-botseed-ws acpwb-botseed-processor acpwb-go
```

## acpwb_go (archive/policy render service)

`acpwb_go/` is a standalone Go service that serves the highest-traffic
honeypot pages (archive/policy day-level content) directly, bypassing
Django/gunicorn for those routes — see
`/Users/dan/.claude/plans/realistically-what-can-we-zippy-wave.md`. It talks
to the same Redis (`REDIS_URL`) as Django, pushing onto the same
`acpwb:crawler_queue`/`acpwb:archive_queue` lists the existing
`drain_crawler_queue`/`drain_archive_queue` crons already consume — no
changes needed on that side.

The box needs a Go toolchain to build it (none of the other services need
one — this is the first Go component):

```bash
# one-time, if Go isn't already installed:
curl -fsSL https://go.dev/dl/go1.24.linux-amd64.tar.gz | sudo tar -C /usr/local -xz
# add /usr/local/go/bin to acpwb's PATH (or build as another user and copy
# the resulting binary — it's a static, CGO_ENABLED=0 binary with no
# runtime deps beyond matching OS/arch)

cd /home/acpwb/acpwb/acpwb_go
go build -o acpwb_go ./cmd/acpwb_go
```

Rebuild (`go build -o acpwb_go ./cmd/acpwb_go`) and `sudo systemctl restart
acpwb-go` on every deploy that touches `acpwb_go/`, same as any other code
change here needing its process restarted.

`deploy/acpwb-go.service` binds it to `127.0.0.1:8091`; nginx routes
`/archive/<year>/<month>/<day>/...`, `/public-policy/...`, and
archive-subdomain day-level content (`archives-YYYY.acpwb.com/<month>/<day>/...`,
via the `$archive_era_subdomain` map in `nginx/conf.d/upstream-acpwb.conf`)
to it, while everything else (including `/archive/` and `/archive/<year>/`
index pages, the archive-subdomain root/month index, and policy-subdomain
rendering) still goes to `django_backend`, unchanged.

**Not yet cut over**: policy subdomain (`policy-<agency>.acpwb.com`)
rendering — that still needs host-based routing wired into `acpwb_go`'s
server before nginx can route that traffic to it. Until then, policy
subdomain traffic keeps going to Django exactly as before this change, so
there's no functional regression from deploying the pieces above on their
own.

### Benchmarks

Two independent, real (not estimated) measurements motivated and validated
this move, plus one gap that's worth being honest about:

**1. Django template rendering vs. hand-written Python string building**
(`acpwb/apps/honeypot/management/commands/bench_template.py`, run with
`python manage.py bench_template --n 200`, median of 200 iterations per
route to filter out GC-pause/cache-miss outliers — this is the same
benchmark whose single-scenario result (3.82ms → 0.96ms, archive default)
justified the earlier "raw-templates" migration in commits `454bd17` →
`bfa0524`). A later, more complete run broken out per template variant:

| Route | Django p50 | Python p50 | Speedup (p50) |
|---|---|---|---|
| archive default (main) | 5.45ms | 1.56ms | 3.49x |
| archive compliance (main) | 5.50ms | 1.55ms | 3.55x |
| archive minutes (main) | 5.57ms | 1.58ms | 3.53x |
| archive default (era/subdomain) | 2.25ms | 1.01ms | 2.23x |
| policy detail | 2.20ms | 0.52ms | 4.23x |
| policy subdomain index | 1.34ms | 0.40ms | 3.35x |
| policy month | 0.66ms | 0.32ms | 2.06x |
| policy subdomain year | 0.71ms | 0.35ms | 2.03x |
| policy year | 1.06ms | 0.73ms | 1.45x |
| policy index | 4.05ms | 2.94ms | 1.38x |

Every converted page is faster on the Python path, 1.4x–4.2x depending on
template weight — heavier pages (archive, policy detail) see the largest
win since Jinja2/Django template overhead scales with template complexity,
while already-thin pages (policy year/index) see a smaller relative gain.
Medians are the reliable number here; raw per-request times are noisier
than this table suggests because a small fraction of requests on both
paths spike into the hundreds of ms from GC pauses or first-touch cache
misses — that's ordinary jitter, not a render-path effect, so don't read
too much into any single sample.

**2. Go render cost, on the actual production hardware** —
`acpwb_go/archive/bench_test.go` (`go test ./archive/ -bench
BenchmarkRenderArchiveDefault -benchmem`), archive default page, run
directly on this box (Intel Xeon E5-2695 v4):

```
BenchmarkRenderArchiveDefault-72    500    1491072 ns/op    796595 B/op    1689 allocs/op
```

~1.49ms/render — in the same range as the already-fast Python string-builder
path above (1.56ms for the same page), not dramatically faster on raw
render cost alone. Go's actual win over the Django/gunicorn/gevent stack
isn't primarily "Go renders faster than Python renders" — it's that
gevent's cooperative scheduling doesn't preempt CPU-bound work, so one
heavy request blocks every other connection on that worker until it
yields (see the plan doc's root-cause analysis), whereas Go's OS-thread-
backed goroutines don't have that failure mode. The render-time numbers
above are a useful sanity check that the port isn't slower, not the
headline result.

**3. Known gap — no persisted end-to-end HTTP throughput comparison.**
Ad-hoc load testing against the running `acpwb_go` service (via `hey` and
a custom Go load-test tool, both cross-compiled for `linux/amd64` to run
inside the local Docker environment) was done earlier in this project to
sanity-check real req/s and confirm `MaxIdleConnsPerHost` wasn't
throttling results, but that tooling and its output were not committed or
saved anywhere durable — there is no reusable script or saved report to
point to here. If a real req/s-under-load number is needed (e.g. to
validate the "10x" target directly rather than by proxy through the
render-time numbers above), it needs to be re-run and this section
updated with the result.

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

### Redis had the same undersized backlog (2026-09-06)

Same bug class as the nginx fix above, on a different socket. After
`acpwb_go` went live (see the acpwb_go section above), `CrawlerVisit`
volume for the `archive`/`policy` trap types it now serves dropped ~62%/
~37% respectively compared to pre-cutover levels — real requests were
still being served correctly (verified via direct `curl` and content
checks), but a large fraction of the corresponding visit-log writes to
Redis were silently failing (`visitqueue`'s fire-and-forget `push()`
swallows errors by design). `redis-cli info stats` showed
`rejected_connections: 93,984` and `connected_clients: 4,620` — Redis's
`tcp-backlog` was still at its default of `511`, identical to nginx's
default before the 2026-09-03 fix, and never updated to match
`net.core.somaxconn` (already 65535 from that fix).

Fix: `/etc/redis/redis.conf` — `tcp-backlog 511` → `tcp-backlog 65535`,
then `sudo systemctl restart redis-server` (a full restart is required —
`tcp-backlog` isn't a `CONFIG SET`-able runtime parameter, unlike most
Redis settings). Confirmed safe to restart: RDB snapshotting is active
with default save points (`3600 1 300 100 60 10000`), and this box's
sustained write volume (tens of thousands of queue pushes/min) easily
clears the "60 seconds, 10000+ changes" threshold, so `lastsave` is
never more than about a minute stale — worst-case data loss from a
restart is roughly that last minute of queued-but-undrained
`crawler_queue`/`archive_queue` items, not the full queue.

Post-fix: `rejected_connections` reset to 0 and stayed there,
`connected_clients` settled at ~214 (down from the pre-fix 4,620) —
strong evidence the inflated client count was itself connection-retry
churn from rejected connections, not genuine concurrent load. All other
Redis consumers on this box (`ws_service`, `botseed_processor`, the
drain crons, `RequestStreamMiddleware`) reconnected cleanly with no
manual intervention needed.

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
