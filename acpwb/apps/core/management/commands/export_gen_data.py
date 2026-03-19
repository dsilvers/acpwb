import json

from django.conf import settings
from django.core.management.base import BaseCommand

from apps.honeypot.report_generator import (
    REPORT_CATALOG,
    _generate_synthetic,
    _rng_from_seed,
)
from apps.projects.generators import INDUSTRIES


class Command(BaseCommand):
    help = "Exports data needed for external generator scripts to a JSON file."

    def handle(self, *args, **options):
        self.stdout.write("Exporting data for image generation...")

        # --- Gather report data ---
        # We take the catalog and add some synthetic ones to get good coverage.
        report_entries = list(REPORT_CATALOG)
        for i in range(74):  # Same number as in the original generate_covers.py
            seed = f"synthetic_cover_item{i}"
            rng = _rng_from_seed(seed)
            entry = _generate_synthetic(rng, seed)
            report_entries.append(entry)

        # --- Assemble data structure ---
        data = {
            "reports": report_entries,
            "projects": {"industries": INDUSTRIES},
        }

        # --- Write to file in the git repo root ---
        output_path = settings.BASE_DIR.parent / "gen_data.json"
        output_path.write_text(json.dumps(data, indent=2))

        self.stdout.write(self.style.SUCCESS(f"Successfully exported data to {output_path}"))