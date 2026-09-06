"""
Second-pass export of literal data pools needed to port the "default"
archive page render path (see export_render_data.py for the original pass
and doc-comment conventions — this follows the same pattern).

Adds: report catalog + report-summary pools (report_generator.py), and the
presentation-generator pools needed to reproduce generate_presentations_for_context
(apps/presentations/generators.py) — title/acronym text helpers, verbs/nouns/
adjectives/themes, title/subtitle/venue templates, and the organizations pool.

Usage (inside the web container):
    python manage.py export_render_data2
    docker cp <web-container>:/tmp/render_data_export2/. ../acpwb_go/data/
"""
import json
from pathlib import Path

from django.core.management.base import BaseCommand

from apps.honeypot import report_generator
from apps.people import generators as people_generators
from apps.presentations.data import verbs, nouns, adjectives, themes, text as pres_text
from apps.presentations.data import slide_templates, organizations

EXPORTS = [
    (report_generator, [
        ('REPORT_CATALOG', 'REPORT_CATALOG'),
        ('YEAR_POOL', 'REPORT_YEAR_POOL'),
        ('SUMMARY_TEMPLATES', 'REPORT_SUMMARY_TEMPLATES'),
        ('FINDING_PHRASES', 'REPORT_FINDING_PHRASES'),
        ('REPORT_ADJECTIVES', 'REPORT_ADJECTIVES'),
        ('REPORT_SUBJECTS', 'REPORT_SUBJECTS'),
        ('REPORT_SUFFIXES', 'REPORT_SUFFIXES'),
    ]),
    (people_generators, [
        ('TITLES', 'PEOPLE_TITLES'),
        ('DEPARTMENTS', 'PEOPLE_DEPARTMENTS'),
    ]),
    (verbs, [('VERBS', 'PRES_VERBS')]),
    (nouns, [('NOUNS', 'PRES_NOUNS')]),
    (adjectives, [('ADJECTIVES', 'PRES_ADJECTIVES')]),
    (themes, [('THEMES', 'PRES_THEMES')]),
    (pres_text, [
        ('TITLE_CASE_LOWER', 'PRES_TITLE_CASE_LOWER'),
        ('ACRONYMS', 'PRES_ACRONYMS'),
    ]),
    (slide_templates, [
        ('TITLE_TEMPLATES', 'PRES_TITLE_TEMPLATES'),
        ('SUBTITLES', 'PRES_SUBTITLES'),
        ('VENUES', 'PRES_VENUES'),
    ]),
    (organizations, [
        ('ORGANIZATIONS', 'PRES_ORGANIZATIONS'),
        ('ORG_SLUG_MAP', 'PRES_ORG_SLUG_MAP'),
    ]),
]


def _to_jsonable(value):
    if isinstance(value, (set, frozenset)):
        return sorted(value)
    if isinstance(value, dict):
        return {str(k): _to_jsonable(v) for k, v in value.items()}
    if isinstance(value, tuple):
        return {'__tuple__': [_to_jsonable(v) for v in value]}
    if isinstance(value, list):
        return [_to_jsonable(v) for v in value]
    return value


class Command(BaseCommand):
    help = 'Export a second pass of literal data pools (reports/presentations) to JSON for the Go service to embed.'

    def add_arguments(self, parser):
        parser.add_argument('--out', default=None)

    def handle(self, *args, **options):
        out_dir = Path(options['out'] or '/tmp/render_data_export2')
        out_dir.mkdir(parents=True, exist_ok=True)

        manifest = {}
        for module, names in EXPORTS:
            for source_name, export_name in names:
                value = getattr(module, source_name)
                jsonable = _to_jsonable(value)
                filename = f'{export_name}.json'
                (out_dir / filename).write_text(json.dumps(jsonable, ensure_ascii=False, indent=None))
                manifest[export_name] = filename
                self.stdout.write(f'  {module.__name__}.{source_name} -> {export_name}')

        (out_dir / '_manifest2.json').write_text(json.dumps(manifest, indent=2, sort_keys=True))
        self.stdout.write(self.style.SUCCESS(f'Exported {len(manifest)} data pools to {out_dir}'))
