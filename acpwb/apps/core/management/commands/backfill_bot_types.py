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

from django.core.management.base import BaseCommand

from apps.core.bot_classify import classify_ua, classify_ua_group
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
        dry_run = options["dry_run"]
        batch_size = options["batch_size"]
        reclassify = options["reclassify"]

        if reclassify:
            qs = CrawlerVisit.objects.filter(
                bot_type__in=["", "Other / Browser"]
            ).only("id", "user_agent")
            label = "blank or 'Other / Browser'"
        else:
            qs = CrawlerVisit.objects.filter(bot_type="").only("id", "user_agent")
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

        for visit in qs.iterator(chunk_size=batch_size):
            ua = visit.user_agent or ""
            visit.bot_type = classify_ua(ua)
            visit.bot_group = classify_ua_group(ua)
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
