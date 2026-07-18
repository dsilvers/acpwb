"""
Categorize 'Other / Browser' CrawlerVisit rows by IP address.

Groups by IP, sorted by request count, with the distinct user-agent strings
seen from each IP. Use this to identify bots masking behind fake browser UAs
so new entries can be added to _IP_BOT_RANGE_DEFS in bot_classify.py.

Usage:
    python manage.py categorize_other_bots
    python manage.py categorize_other_bots --date 2026-07-15
    python manage.py categorize_other_bots --limit 100 --min-hits 1000
    python manage.py categorize_other_bots --all-time
    python manage.py categorize_other_bots --subnet 24
"""
from datetime import datetime, timedelta

from django.core.management.base import BaseCommand, CommandError
from django.db import connection
from django.utils import timezone


class Command(BaseCommand):
    help = "Group 'Other / Browser' visits by IP to identify uncategorized bots."

    def add_arguments(self, parser):
        group = parser.add_mutually_exclusive_group()
        group.add_argument('--since',    type=int,            help='Lookback window in days (default: 30)')
        group.add_argument('--date',     type=str,            help='Single day to analyse, YYYY-MM-DD')
        group.add_argument('--all-time', action='store_true', help='Scan all rows (no time filter — slow on large tables)')
        parser.add_argument('--limit',    type=int, default=50,  help='Top N IPs to show (default: 50)')
        parser.add_argument('--min-hits', type=int, default=100, help='Skip IPs with fewer hits than this (default: 100)')
        parser.add_argument('--ua-limit', type=int, default=5,   help='Max UAs to show per IP (default: 5)')
        parser.add_argument('--subnet',   type=int, choices=[8, 16, 24], help='Group by subnet prefix length instead of individual IP')

    def handle(self, *args, **options):
        all_time = options['all_time']
        limit    = options['limit']
        min_hits = options['min_hits']
        ua_limit = options['ua_limit']
        subnet   = options['subnet']

        if all_time:
            self.stdout.write(self.style.WARNING(
                "WARNING: --all-time scans all rows with no time filter. "
                "This may be very slow on large TimescaleDB tables.\n"
            ))
            time_clause = ""
            time_params = []
            window_label = "all time"
        elif options['date']:
            try:
                day_start = datetime.strptime(options['date'], '%Y-%m-%d')
            except ValueError:
                raise CommandError("--date must be in YYYY-MM-DD format")
            day_start = timezone.make_aware(day_start)
            day_end   = day_start + timedelta(days=1)
            time_clause = "AND timestamp >= %s AND timestamp < %s"
            time_params = [day_start, day_end]
            window_label = options['date']
        else:
            days = options['since'] or 30
            cutoff = timezone.now() - timedelta(days=days)
            time_clause = "AND timestamp >= %s"
            time_params = [cutoff]
            window_label = f"last {days} days"

        if subnet:
            self._run_subnet(time_clause, time_params, window_label, limit, min_hits, ua_limit, subnet)
        else:
            self._run_by_ip(time_clause, time_params, window_label, limit, min_hits, ua_limit)

    def _run_by_ip(self, time_clause, time_params, window_label, limit, min_hits, ua_limit):
        # Step 1: top IPs by hit count
        with connection.cursor() as cursor:
            cursor.execute(f"""
                SELECT ip_address, COUNT(*) AS hits, COUNT(DISTINCT user_agent) AS unique_uas
                FROM honeypot_crawlervisit
                WHERE bot_type = 'Other / Browser'
                {time_clause}
                GROUP BY ip_address
                HAVING COUNT(*) >= %s
                ORDER BY hits DESC
                LIMIT %s
            """, time_params + [min_hits, limit])
            rows = cursor.fetchall()

        if not rows:
            self.stdout.write("No results.")
            return

        total_in_window = sum(r[1] for r in rows)
        self.stdout.write(self.style.SUCCESS(
            f"\n'Other / Browser' visits — {window_label}\n"
            f"Showing top {len(rows)} IPs (min {min_hits:,} hits each); "
            f"{total_in_window:,} hits combined\n"
        ))

        # Step 2: fetch top UAs for each IP in one query
        ip_list = [r[0] for r in rows]
        ua_map = self._fetch_top_uas(ip_list, ua_limit, time_clause, time_params, group_col='ip_address')

        self.stdout.write("=== Top IPs by Request Count ===\n")
        for ip, hits, unique_uas in rows:
            self.stdout.write(f"  {hits:>10,} hits  {unique_uas:>4} UAs   {ip}")
            for ua in ua_map.get(ip, []):
                self.stdout.write(f"{'':>35}{ua[:120]!r}")
            self.stdout.write("")

    def _run_subnet(self, time_clause, time_params, window_label, limit, min_hits, ua_limit, prefix_len):
        # Build SQL expression for subnet grouping
        if prefix_len == 24:
            subnet_expr = (
                "split_part(ip_address, '.', 1) || '.' || "
                "split_part(ip_address, '.', 2) || '.' || "
                "split_part(ip_address, '.', 3) || '.0/24'"
            )
        elif prefix_len == 16:
            subnet_expr = (
                "split_part(ip_address, '.', 1) || '.' || "
                "split_part(ip_address, '.', 2) || '.0.0/16'"
            )
        else:  # /8
            subnet_expr = "split_part(ip_address, '.', 1) || '.0.0.0/8'"

        with connection.cursor() as cursor:
            cursor.execute(f"""
                SELECT
                    {subnet_expr} AS subnet,
                    COUNT(*)                        AS hits,
                    COUNT(DISTINCT ip_address)      AS unique_ips,
                    COUNT(DISTINCT user_agent)      AS unique_uas,
                    MAX(ip_address)                 AS sample_ip
                FROM honeypot_crawlervisit
                WHERE bot_type = 'Other / Browser'
                {time_clause}
                GROUP BY subnet
                HAVING COUNT(*) >= %s
                ORDER BY hits DESC
                LIMIT %s
            """, time_params + [min_hits, limit])
            rows = cursor.fetchall()

        if not rows:
            self.stdout.write("No results.")
            return

        total_in_window = sum(r[1] for r in rows)
        self.stdout.write(self.style.SUCCESS(
            f"\n'Other / Browser' visits — {window_label}\n"
            f"Showing top {len(rows)} /{prefix_len} subnets (min {min_hits:,} hits each); "
            f"{total_in_window:,} hits combined\n"
        ))

        ip_list = [r[4] for r in rows]
        ua_map = self._fetch_top_uas(ip_list, ua_limit, time_clause, time_params, group_col=None)

        self.stdout.write(f"=== Top /{prefix_len} Subnets by Request Count ===\n")
        for subnet, hits, unique_ips, unique_uas, sample_ip in rows:
            self.stdout.write(
                f"  {hits:>10,} hits  {unique_ips:>5} IPs  {unique_uas:>4} UAs   {subnet}"
            )
            for ua in ua_map.get(sample_ip, []):
                self.stdout.write(f"{'':>43}{ua[:110]!r}")
            self.stdout.write(f"{'':>43}(sample IP: {sample_ip})")
            self.stdout.write("")

        self.stdout.write(
            "\nTo add a range to bot_classify.py, append to _IP_BOT_RANGE_DEFS:\n"
            "    ('x.x.x.0/24', 'Org Name'),  # N hits\n"
            "Then run: python manage.py backfill_bot_types --reclassify\n"
        )

    def _fetch_top_uas(self, ip_list, ua_limit, time_clause, time_params, group_col):
        """
        Fetch the top UA strings for each IP in ip_list.
        Returns a dict: ip -> [ua_string, ...]
        group_col='ip_address' for per-IP mode; None for subnet mode (keyed by sample_ip).
        """
        if not ip_list:
            return {}

        placeholders = ','.join(['%s'] * len(ip_list))
        with connection.cursor() as cursor:
            cursor.execute(f"""
                SELECT ip_address, user_agent, COUNT(*) AS c
                FROM honeypot_crawlervisit
                WHERE bot_type = 'Other / Browser'
                  AND ip_address IN ({placeholders})
                  {time_clause}
                GROUP BY ip_address, user_agent
                ORDER BY ip_address, c DESC
            """, ip_list + time_params)
            raw = cursor.fetchall()

        ua_map = {}
        for ip, ua, _count in raw:
            if ip not in ua_map:
                ua_map[ip] = []
            if len(ua_map[ip]) < ua_limit:
                ua_map[ip].append(ua or '(empty)')
        return ua_map
