"""
Generate a large Django fixture of ProjectStory and PublicReport records.

Usage (inside the container):
    python manage.py generate_content_fixture
    python manage.py generate_content_fixture --reports 800 --project-pages 150
    python manage.py generate_content_fixture --output /tmp/content.json

The fixture is written to acpwb/fixtures/content.json by default, which
Django's loaddata command can read from any app's fixtures/ directory.

To load on production:
    python manage.py loaddata content
"""

import json
from datetime import date, datetime, timezone
from pathlib import Path

from django.core.management.base import BaseCommand

from apps.honeypot.report_generator import (
    REPORT_CATALOG,
    REPORT_CATEGORIES,
    REPORT_ADJECTIVES,
    REPORT_SUBJECTS,
    REPORT_SUFFIXES,
    YEAR_POOL,
    _rng_from_seed,
    _enrich_report,
    _generate_synthetic,
)
from apps.projects.generators import generate_project_stories

FIXTURE_TIMESTAMP = "2025-01-15T00:00:00+00:00"
DEFAULT_OUTPUT = Path(__file__).resolve().parents[5] / "fixtures" / "content.json"


class Command(BaseCommand):
    help = "Generate a large fixture of projects and reports for pre-loading on production."

    def add_arguments(self, parser):
        parser.add_argument(
            "--reports",
            type=int,
            default=500,
            help="Total number of PublicReport records to generate (default: 500)",
        )
        parser.add_argument(
            "--project-pages",
            type=int,
            default=100,
            help="Number of project pages to generate, 10 stories each (default: 100 = 1000 projects)",
        )
        parser.add_argument(
            "--output",
            type=str,
            default=str(DEFAULT_OUTPUT),
            help=f"Output path for the fixture JSON (default: {DEFAULT_OUTPUT})",
        )

    def handle(self, *args, **options):
        report_count = options["reports"]
        project_pages = options["project_pages"]
        output_path = Path(options["output"])

        output_path.parent.mkdir(parents=True, exist_ok=True)

        fixture = []
        pk = 1

        # ── Projects ──────────────────────────────────────────────────────────

        self.stdout.write(f"Generating {project_pages} pages of project stories ({project_pages * 10} total)...")
        for page in range(1, project_pages + 1):
            if page % 10 == 0:
                self.stdout.write(f"  page {page}/{project_pages}")
            for story in generate_project_stories(page=page, count=10):
                fixture.append({
                    "model": "projects.projectstory",
                    "pk": pk,
                    "fields": {
                        "slug": story["slug"],
                        "title": story["title"],
                        "summary": story["summary"],
                        "body_paragraphs": story["body_paragraphs"],
                        "impact_metric": story["impact_metric"],
                        "industry_tag": story["industry_tag"],
                        "page_number": story["page_number"],
                        "generated_at": FIXTURE_TIMESTAMP,
                    },
                })
                pk += 1

        project_count = pk - 1
        self.stdout.write(self.style.SUCCESS(f"  ✓ {project_count} project stories"))

        # ── Reports ───────────────────────────────────────────────────────────

        pk = 1
        self.stdout.write(f"Generating {report_count} public reports...")

        # Start with the full catalog
        catalog_entries = list(REPORT_CATALOG)
        seen_slugs = {e["slug"] for e in catalog_entries}

        # Generate synthetic entries to reach the target count
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

        for idx, entry in enumerate(catalog_entries):
            if idx % 50 == 0 and idx > 0:
                self.stdout.write(f"  {idx}/{report_count}")
            enriched = _enrich_report(entry)
            pub_date = date.fromisoformat(enriched["pub_date"])
            page_num = max(1, (idx // 12) + 1)
            fixture.append({
                "model": "honeypot.publicreport",
                "pk": pk,
                "fields": {
                    "slug": enriched["slug"],
                    "title": enriched["title"],
                    "category": enriched["category"],
                    "file_type": enriched["file_type"],
                    "pub_date": pub_date.isoformat(),
                    "summary": enriched["summary"],
                    "watermark_token": enriched["watermark_token"],
                    "page_number": page_num,
                    "generated_at": FIXTURE_TIMESTAMP,
                },
            })
            pk += 1

        self.stdout.write(self.style.SUCCESS(f"  ✓ {report_count} public reports"))

        # ── Write fixture ─────────────────────────────────────────────────────

        output_path.write_text(json.dumps(fixture, indent=2))
        size_kb = output_path.stat().st_size // 1024

        self.stdout.write("")
        self.stdout.write(self.style.SUCCESS(f"Fixture written to: {output_path}"))
        self.stdout.write(f"  {project_count} ProjectStory records")
        self.stdout.write(f"  {report_count} PublicReport records")
        self.stdout.write(f"  File size: {size_kb} KB")
        self.stdout.write("")
        self.stdout.write("To load on any environment:")
        self.stdout.write("  python manage.py loaddata content")
