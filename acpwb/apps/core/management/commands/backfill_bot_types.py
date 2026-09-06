"""
Backfill bot_type and bot_group on CrawlerVisit rows that are blank or
classified as 'Other / Browser' (pattern didn't match at write time).

Use --reclassify to also update rows currently set to 'Other / Browser'
so that newly added BOT_PATTERNS take effect on historical data.

Usage:
    python manage.py backfill_bot_types               # blank rows only
    python manage.py backfill_bot_types --reclassify  # blank + Other / Browser
    python manage.py backfill_bot_types --reclassify --dry-run
    python manage.py backfill_bot_types --batch-size 2000
"""

import fcntl

from django.core.management.base import BaseCommand
from django.db import transaction

from apps.core.bot_classify import bot_type_to_group, classify_ua_or_ip
from apps.honeypot.models import CrawlerVisit

BATCH_SIZE = 1000


class Command(BaseCommand):
    help = "Backfill bot_type and bot_group on CrawlerVisit rows with blank or unmatched values."

    def add_arguments(self, parser):
        parser.add_argument(
            "--dry-run",
            action="store_true",
            help="Print stats without updating anything",
        )
        parser.add_argument(
            "--batch-size",
            type=int,
            default=BATCH_SIZE,
            help=f"Rows per bulk_update call (default: {BATCH_SIZE})",
        )
        parser.add_argument(
            "--reclassify",
            action="store_true",
            help="Also reclassify rows currently set to 'Other / Browser' (picks up new patterns)",
        )

    def handle(self, *args, **options):
        # precalc_dashboard's crawlers.by_bot_type/by_bot_group (and the
        # daily/recent-bucket charts derived from them) are incremental,
        # driven by a timestamp high-water-mark that only ever advances. If
        # it scans a row in the same instant this command is mid-UPDATE on
        # that row's bot_type, the row gets permanently counted under the
        # empty-string bucket — the HWM moves past it, so no later run
        # re-visits it (see precalc_dashboard.py's --reset-bot-types flag,
        # which exists specifically to correct exactly this class of
        # staleness after a backfill run). Sharing precalc_dashboard's own
        # lock file here means this command simply skips its run (retrying
        # next minute, at negligible cost) whenever precalc_dashboard is
        # already running, rather than racing it.
        with open('/tmp/precalc_dashboard.lock', 'w') as precalc_lockfile:
            try:
                fcntl.flock(precalc_lockfile, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError:
                self.stdout.write('precalc_dashboard is running — skipping this tick.')
                return
            self._run(options)
            fcntl.flock(precalc_lockfile, fcntl.LOCK_UN)

    def _run(self, options):
        dry_run = options["dry_run"]
        batch_size = options["batch_size"]
        reclassify = options["reclassify"]

        if reclassify:
            qs = CrawlerVisit.objects.filter(
                bot_type__in=["", "Other / Browser"]
            ).only("id", "user_agent", "ip_address")
            label = "blank or 'Other / Browser'"
        else:
            qs = CrawlerVisit.objects.filter(bot_type="").only("id", "user_agent", "ip_address")
            label = "blank"

        total = qs.count()
        self.stdout.write(f"CrawlerVisit rows with {label} bot_type: {total:,}")

        if total == 0:
            self.stdout.write(self.style.SUCCESS("Nothing to do."))
            return

        if dry_run:
            self.stdout.write(self.style.WARNING(f"DRY RUN — would update {total:,} rows"))
            return

        updated = 0
        batch = []

        # PgBouncer transaction-pooling mode can hand this iterator()'s
        # server-side cursor to a different backend connection between
        # FETCHes unless the whole streamed loop stays in one transaction.
        with transaction.atomic():
            for visit in qs.iterator(chunk_size=batch_size):
                ua = visit.user_agent or ""
                ip = visit.ip_address or ""
                visit.bot_type = classify_ua_or_ip(ua, ip)
                visit.bot_group = bot_type_to_group(visit.bot_type)
                batch.append(visit)

                if len(batch) >= batch_size:
                    CrawlerVisit.objects.bulk_update(batch, ["bot_type", "bot_group"])
                    updated += len(batch)
                    batch = []
                    self.stdout.write(f"  updated {updated:,} / {total:,} ...")

            if batch:
                CrawlerVisit.objects.bulk_update(batch, ["bot_type", "bot_group"])
                updated += len(batch)

        self.stdout.write(self.style.SUCCESS(f"Done — updated {updated:,} rows."))
