"""
Pre-generate a large set of ProjectStory and PublicReport records.

Uses bulk_create(ignore_conflicts=True) — safe to run multiple times and on
databases that already have content. Existing records (matched by slug) are
left untouched; only new ones are inserted.

Usage (inside the container):
    python manage.py generate_content_fixture
    python manage.py generate_content_fixture --reports 800 --project-pages 150

To run on production:
    docker compose exec web python manage.py generate_content_fixture
"""

from django.core.management.base import BaseCommand

from apps.honeypot.models import PublicReport
from apps.honeypot.report_generator import (
    REPORT_CATALOG,
    _enrich_report,
    _generate_synthetic,
    _rng_from_seed,
)
from apps.projects.generators import generate_project_stories
from apps.projects.models import ProjectStory


class Command(BaseCommand):
    help = "Pre-generate projects and reports. Safe to re-run — skips existing slugs."

    def add_arguments(self, parser):
        parser.add_argument(
            "--reports",
            type=int,
            default=500,
            help="Total PublicReport records to generate (default: 500)",
        )
        parser.add_argument(
            "--project-pages",
            type=int,
            default=100,
            help="Project pages to generate, 10 stories each (default: 100 = 1000 projects)",
        )

    def handle(self, *args, **options):
        report_count = options["reports"]
        project_pages = options["project_pages"]

        # ── Projects ──────────────────────────────────────────────────────────

        self.stdout.write(
            f"Generating {project_pages} pages × 10 = {project_pages * 10} project stories..."
        )

        existing_slugs = set(ProjectStory.objects.values_list("slug", flat=True))
        to_create = []

        for page in range(1, project_pages + 1):
            if page % 25 == 0:
                self.stdout.write(f"  page {page}/{project_pages}")
            for story in generate_project_stories(page=page, count=10):
                if story["slug"] not in existing_slugs:
                    to_create.append(
                        ProjectStory(
                            slug=story["slug"],
                            title=story["title"],
                            summary=story["summary"],
                            body_paragraphs=story["body_paragraphs"],
                            impact_metric=story["impact_metric"],
                            industry_tag=story["industry_tag"],
                            page_number=story["page_number"],
                        )
                    )
                    existing_slugs.add(story["slug"])

        ProjectStory.objects.bulk_create(to_create, ignore_conflicts=True, batch_size=200)
        self.stdout.write(
            self.style.SUCCESS(
                f"  ✓ {len(to_create)} new project stories inserted"
                f" ({project_pages * 10 - len(to_create)} already existed)"
            )
        )

        # ── Reports ───────────────────────────────────────────────────────────

        self.stdout.write(f"Generating {report_count} public reports...")

        existing_report_slugs = set(PublicReport.objects.values_list("slug", flat=True))
        catalog_entries = list(REPORT_CATALOG)
        seen_slugs = {e["slug"] for e in catalog_entries}

        i = 0
        while len(catalog_entries) < report_count:
            seed = f"fixture_report_{i}"
            rng = _rng_from_seed(seed)
            entry = _generate_synthetic(rng, seed)
            if entry["slug"] not in seen_slugs:
                catalog_entries.append(entry)
                seen_slugs.add(entry["slug"])
            i += 1

        catalog_entries = catalog_entries[:report_count]

        import datetime
        reports_to_create = []
        for idx, entry in enumerate(catalog_entries):
            if idx % 100 == 0 and idx > 0:
                self.stdout.write(f"  {idx}/{report_count}")
            if entry["slug"] in existing_report_slugs:
                continue
            enriched = _enrich_report(entry)
            pub_date = datetime.date.fromisoformat(enriched["pub_date"])
            reports_to_create.append(
                PublicReport(
                    slug=enriched["slug"],
                    title=enriched["title"],
                    category=enriched["category"],
                    file_type=enriched["file_type"],
                    pub_date=pub_date,
                    summary=enriched["summary"],
                    watermark_token=enriched["watermark_token"],
                    page_number=max(1, (idx // 12) + 1),
                )
            )
            existing_report_slugs.add(entry["slug"])

        PublicReport.objects.bulk_create(reports_to_create, ignore_conflicts=True, batch_size=200)
        self.stdout.write(
            self.style.SUCCESS(
                f"  ✓ {len(reports_to_create)} new reports inserted"
                f" ({report_count - len(reports_to_create)} already existed)"
            )
        )

        # ── Summary ───────────────────────────────────────────────────────────

        self.stdout.write("")
        self.stdout.write(self.style.SUCCESS("Done."))
        self.stdout.write(
            f"  Total projects in DB: {ProjectStory.objects.count()}"
        )
        self.stdout.write(
            f"  Total reports in DB:  {PublicReport.objects.count()}"
        )
