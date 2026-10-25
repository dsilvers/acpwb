# Production crontab and bot-type backfill

Detail moved out of CLAUDE.md.

Crontab (production) — see `deploy/acpwb-crontab` for the actual installed form (direct `manage.py` invocation via `.direnv`, not `docker compose exec`):
```
*/30 * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py precalc_dashboard >> /var/log/acpwb-precalc.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_crawler_queue >> /var/log/acpwb-crawler-drain.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_archive_queue >> /var/log/acpwb-archive-drain.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_fingerprint_queue >> /var/log/acpwb-fingerprint-drain.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_code_email_queue >> /var/log/acpwb-code-email-drain.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_compliance_queue >> /var/log/acpwb-compliance-drain.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_edge_conn_queue >> /var/log/acpwb-edge-conn-drain.log 2>&1
* * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py drain_edge_request_queue >> /var/log/acpwb-edge-request-drain.log 2>&1
*/5 * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py score_ip_reputation >> /var/log/acpwb-score-ip-reputation.log 2>&1
*/15 * * * * docker compose -f /home/dan/acpwb.com/docker-compose.yml exec -T web python manage.py publish_ip_reputation >> /var/log/acpwb-publish-ip-reputation.log 2>&1

`acpwb_go` (the Go render service for archive/policy pages) writes `CrawlerVisit` rows with `bot_type`/`bot_group` blank — it doesn't classify bots itself. A `backfill_bot_types` cron job was tried to fill these in incrementally but is **currently disabled** (not installed in production) — its first real run against the live hypertable never completed a single 1000-row batch in 8+ minutes, because `bulk_update()`'s per-row `CASE WHEN id=X` UPDATE forces expensive cross-chunk work on a non-time-key access pattern against a 373M+ row TimescaleDB table, and the backlog grew faster than it could shrink. See `deploy/acpwb-crontab`'s comment block for the incident detail. Bot classification for `acpwb_go`-originated rows needs a different approach (most likely: classify at write time in Go, matching what `apps/core/bot_classify.py` already does) before this gap is closed.
