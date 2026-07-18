"""
Fast SQL-based reclassification of CrawlerVisit bot_type and bot_group.

Instead of iterating rows in Python, generates a CASE WHEN SQL UPDATE from
BOT_PATTERNS and _IP_BOT_RANGE_DEFS, then applies it one TimescaleDB chunk
at a time — no Python row overhead.

Usage:
    python manage.py backfill_bot_types_sql
    python manage.py backfill_bot_types_sql --dry-run
    python manage.py backfill_bot_types_sql --print-sql
"""
from django.core.management.base import BaseCommand
from django.db import connection

from apps.core.bot_classify import BOT_PATTERNS, _IP_BOT_RANGE_DEFS, bot_type_to_group


class Command(BaseCommand):
    help = "Fast CASE WHEN SQL reclassification of CrawlerVisit bot_type/bot_group, chunk by chunk."

    def add_arguments(self, parser):
        parser.add_argument('--dry-run',   action='store_true', help='Show chunk list and row counts without updating')
        parser.add_argument('--print-sql', action='store_true', help='Print the generated UPDATE SQL and exit')

    def handle(self, *args, **options):
        bot_type_sql  = self._build_case('bot_type')
        bot_group_sql = self._build_case('bot_group')

        update_sql = f"""
            UPDATE honeypot_crawlervisit
            SET
                bot_type  = {bot_type_sql},
                bot_group = {bot_group_sql}
            WHERE
                timestamp >= %s AND timestamp < %s
                AND bot_type IN ('', 'Other / Browser')
        """

        if options['print_sql']:
            sample = update_sql.replace('%s', "'<start>'", 1).replace('%s', "'<end>'", 1)
            self.stdout.write(sample)
            return

        chunks = self._get_chunks()
        if not chunks:
            self.stdout.write("No chunks found — nothing to do.")
            return

        self.stdout.write(f"Found {len(chunks)} TimescaleDB chunks to process.\n")

        total_updated = 0
        for i, (start, end) in enumerate(chunks, 1):
            label = f"{start} → {end}"

            if options['dry_run']:
                with connection.cursor() as cursor:
                    cursor.execute(
                        "SELECT COUNT(*) FROM honeypot_crawlervisit "
                        "WHERE timestamp >= %s AND timestamp < %s AND bot_type IN ('', 'Other / Browser')",
                        [start, end],
                    )
                    count = cursor.fetchone()[0]
                self.stdout.write(f"  [{i:>3}/{len(chunks)}] {label}  — {count:,} rows would update")
                continue

            with connection.cursor() as cursor:
                cursor.execute(update_sql, [start, end])
                updated = cursor.rowcount

            total_updated += updated
            self.stdout.write(f"  [{i:>3}/{len(chunks)}] {label}  — {updated:,} rows updated")

        if not options['dry_run']:
            self.stdout.write(self.style.SUCCESS(f"\nDone — {total_updated:,} rows updated total."))

    def _build_case(self, target):
        """Build a SQL CASE WHEN expression for bot_type or bot_group."""
        lines = ["CASE"]

        # Empty / blank UA → special label (matches Python classify_ua() behaviour)
        lines.append("    WHEN user_agent IS NULL OR trim(user_agent) = '' THEN '(empty user agent)'")

        # UA substring patterns — order matches BOT_PATTERNS; first hit wins
        for pattern, name in BOT_PATTERNS:
            value   = name if target == 'bot_type' else bot_type_to_group(name)
            escaped = pattern.replace("'", "''")
            lines.append(f"    WHEN user_agent ILIKE '%{escaped}%' THEN '{value}'")

        # IP range fallback — only reached when no UA pattern matched
        for cidr, name in _IP_BOT_RANGE_DEFS:
            value = name if target == 'bot_type' else bot_type_to_group(name)
            # Guard against non-IPv4 values before casting
            lines.append(
                f"    WHEN ip_address ~ '^[0-9]{{1,3}}(\\.[0-9]{{1,3}}){{3}}$'"
                f" AND ip_address::inet << '{cidr}'::inet THEN '{value}'"
            )

        lines.append("    ELSE 'Other / Browser'")
        lines.append("END")
        return "\n".join(lines)

    def _get_chunks(self):
        """Return [(range_start, range_end), ...] from TimescaleDB chunk metadata."""
        try:
            with connection.cursor() as cursor:
                cursor.execute("""
                    SELECT range_start, range_end
                    FROM timescaledb_information.chunks
                    WHERE hypertable_name = 'honeypot_crawlervisit'
                      AND range_start IS NOT NULL
                    ORDER BY range_start
                """)
                rows = cursor.fetchall()
            if rows:
                return rows
        except Exception as exc:
            self.stdout.write(self.style.WARNING(f"TimescaleDB chunk query failed ({exc}); falling back to monthly intervals."))

        # Fallback: monthly intervals over the affected rows
        from django.utils import timezone

        with connection.cursor() as cursor:
            cursor.execute(
                "SELECT MIN(timestamp), MAX(timestamp) FROM honeypot_crawlervisit "
                "WHERE bot_type IN ('', 'Other / Browser')"
            )
            min_ts, max_ts = cursor.fetchone()

        if not min_ts:
            return []

        # Align to month boundaries
        from datetime import timedelta
        import calendar
        chunks = []
        current = min_ts.replace(day=1, hour=0, minute=0, second=0, microsecond=0)
        while current <= max_ts:
            _, days_in_month = calendar.monthrange(current.year, current.month)
            next_month = current + timedelta(days=days_in_month)
            next_month = next_month.replace(day=1, hour=0, minute=0, second=0, microsecond=0)
            chunks.append((current, next_month))
            current = next_month
        return chunks
