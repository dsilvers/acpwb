"""
Remove duplicate CrawlerVisit records created by the middleware+view double-logging bug.

Two records are considered duplicates when they share the same ip_address, path,
and user_agent and were created within 1 second of each other (same Django request).
The more specific record is kept (non-'other' trap_type preferred; lower id as
tiebreaker).

Usage:
    python manage.py dedupe_crawler_visits
    python manage.py dedupe_crawler_visits --dry-run
    python manage.py dedupe_crawler_visits --window 2   # seconds, default 1
"""

from datetime import timedelta

from django.core.management.base import BaseCommand
from django.db import transaction

from apps.honeypot.models import CrawlerVisit

BATCH_SIZE = 500


class Command(BaseCommand):
    help = "Remove duplicate CrawlerVisit records from the middleware+view double-logging bug."

    def add_arguments(self, parser):
        parser.add_argument(
            "--dry-run",
            action="store_true",
            help="Print what would be deleted without deleting anything",
        )
        parser.add_argument(
            "--window",
            type=float,
            default=1.0,
            help="Max seconds between two records to consider them duplicates (default: 1)",
        )

    def handle(self, *args, **options):
        dry_run = options["dry_run"]
        window = timedelta(seconds=options["window"])

        self.stdout.write(
            f"Scanning CrawlerVisit records for duplicates "
            f"(window={options['window']}s, dry_run={dry_run})..."
        )

        total = CrawlerVisit.objects.count()
        self.stdout.write(f"  Total records: {total:,}")

        # Stream all records ordered so duplicates are adjacent
        qs = CrawlerVisit.objects.order_by(
            "ip_address", "path", "user_agent", "timestamp", "id"
        ).values("id", "ip_address", "path", "user_agent", "timestamp", "trap_type")

        to_delete = []
        prev = None
        examined = 0
        duplicates_found = 0

        for visit in qs.iterator(chunk_size=2000):
            examined += 1
            if examined % 50000 == 0:
                self.stdout.write(f"  examined {examined:,} / {total:,} ...")

            if (
                prev is not None
                and prev["ip_address"] == visit["ip_address"]
                and prev["path"] == visit["path"]
                and prev["user_agent"] == visit["user_agent"]
                and (visit["timestamp"] - prev["timestamp"]) <= window
            ):
                # Duplicate pair — keep the more specific record
                if prev["trap_type"] == "other" and visit["trap_type"] != "other":
                    # prev is the generic middleware record — drop it
                    to_delete.append(prev["id"])
                    prev = visit  # the kept record becomes the new prev
                else:
                    # Keep prev (already specific or both same), drop this one
                    to_delete.append(visit["id"])
                    # prev stays

                duplicates_found += 1

                if len(to_delete) >= BATCH_SIZE and not dry_run:
                    CrawlerVisit.objects.filter(id__in=to_delete).delete()
                    to_delete = []
            else:
                prev = visit

        # Final batch
        if to_delete and not dry_run:
            CrawlerVisit.objects.filter(id__in=to_delete).delete()

        if dry_run:
            self.stdout.write(
                self.style.WARNING(
                    f"DRY RUN — would delete {duplicates_found:,} duplicate records "
                    f"({total - duplicates_found:,} would remain)"
                )
            )
        else:
            self.stdout.write(
                self.style.SUCCESS(
                    f"Deleted {duplicates_found:,} duplicate records "
                    f"({total - duplicates_found:,} remain)"
                )
            )
