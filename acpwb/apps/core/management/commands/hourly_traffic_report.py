from collections import defaultdict

from django.core.management.base import BaseCommand
from django.db.models.functions import TruncHour
from django.db.models import Count
from django.utils import timezone

from apps.honeypot.models import CrawlerVisit


class Command(BaseCommand):
    help = (
        "Hourly breakdown (last 24h) of pages served and unique visitor IPs, "
        "drawn from CrawlerVisit. Note: HTTP response codes are NOT stored in "
        "the database anywhere (only briefly published on the ephemeral Redis "
        "request_stream for the live dashboard) — that data only exists in the "
        "nginx access logs."
    )

    def add_arguments(self, parser):
        parser.add_argument(
            "--hours", type=int, default=24, help="Lookback window in hours (default 24)"
        )

    def handle(self, *args, **options):
        hours = options["hours"]
        since = timezone.now() - timezone.timedelta(hours=hours)

        base = CrawlerVisit.objects.filter(timestamp__gte=since)

        # Single GROUP BY bucket query — DB does the distinct-IP counting,
        # not Python (pulling every row's IP client-side was the slow part).
        stats_qs = (
            base.annotate(bucket=TruncHour("timestamp"))
            .values("bucket")
            .annotate(unique_ips=Count("ip_address", distinct=True), pages=Count("id"))
            .order_by("bucket")
        )
        stats_by_bucket = {row["bucket"]: row for row in stats_qs}

        trap_qs = (
            base.annotate(bucket=TruncHour("timestamp"))
            .values("bucket", "trap_type")
            .annotate(n=Count("id"))
        )
        trap_by_bucket = defaultdict(lambda: defaultdict(int))
        for row in trap_qs:
            trap_by_bucket[row["bucket"]][row["trap_type"]] += row["n"]

        buckets = sorted(stats_by_bucket)

        if not buckets:
            self.stdout.write(self.style.WARNING(f"No visits logged in the last {hours}h."))
            return

        totals = base.aggregate(
            total_ips=Count("ip_address", distinct=True), total_pages=Count("id")
        )

        self.stdout.write(
            self.style.WARNING(
                "Response codes are not logged in the database (nowhere on CrawlerVisit, "
                "ArchiveVisit, or PeoplePageVisit) — they only pass through Redis "
                "momentarily for the live dashboard stream and aren't persisted. "
                "Skipping that column; nginx access logs are the only place status "
                "codes are recorded.\n"
            )
        )

        header = f"{'Hour (local)':<20} {'Unique IPs':>10} {'Pages Served':>13}"
        self.stdout.write(header)
        self.stdout.write("-" * len(header))

        for bucket in buckets:
            local_bucket = timezone.localtime(bucket)
            row = stats_by_bucket[bucket]

            self.stdout.write(
                f"{local_bucket.strftime('%Y-%m-%d %H:00'):<20} {row['unique_ips']:>10} {row['pages']:>13}"
            )

            traps = trap_by_bucket.get(bucket)
            if traps:
                trap_str = ", ".join(
                    f"{t}={n}" for t, n in sorted(traps.items(), key=lambda kv: -kv[1])
                )
                self.stdout.write(f"{'':<20} {'':>10}   {trap_str}")

        self.stdout.write("-" * len(header))
        self.stdout.write(
            f"{'TOTAL (dedup IPs)':<20} {totals['total_ips']:>10} {totals['total_pages']:>13}"
        )
