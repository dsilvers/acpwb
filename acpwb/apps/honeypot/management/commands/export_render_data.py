"""
One-time (repeatable) export of the literal Python data pools that back the
archive/policy content generators, to JSON, so the acpwb_go service can embed
them at build time instead of hand-porting ~16,500 lines of literal data.

Exports exactly the names actually imported by policy_generator.py and
views.py's archive content builders — nothing more, nothing inferred.

Usage (inside the web container, where only ./acpwb is mounted — the
sibling acpwb_go/ directory is NOT reachable from in here):
    python manage.py export_render_data
    docker cp <web-container>:/tmp/render_data_export/. ../acpwb_go/data/
"""
import json
from pathlib import Path

from django.core.management.base import BaseCommand

from apps.honeypot import archive_data, archive_data_compliance, archive_data_minutes, policy_data
from apps.people import generators as people_generators

# (module, [(source_name, export_name), ...])
EXPORTS = [
    (archive_data, [
        ('_ARCHIVE_SLUGS', 'ARCHIVE_SLUGS'),
        ('_ARCHIVE_ORGS', 'ARCHIVE_ORGS'),
        ('ARCHIVE_INDUSTRIES', 'ARCHIVE_INDUSTRIES'),
        ('_ARCHIVE_PHASES', 'ARCHIVE_PHASES'),
        ('_ARCHIVE_PARA_TEMPLATES', 'ARCHIVE_PARA_TEMPLATES'),
        ('_ARCHIVE_METRIC_NAMES', 'ARCHIVE_METRIC_NAMES'),
        ('_ARCHIVE_FINDING_TEMPLATES', 'ARCHIVE_FINDING_TEMPLATES'),
        ('_ARCHIVE_METRIC_LABELS', 'ARCHIVE_METRIC_LABELS'),
        ('_ARCHIVE_TITLE_PREFIXES', 'ARCHIVE_TITLE_PREFIXES'),
        ('_ARCHIVE_YEAR_DATA', 'ARCHIVE_YEAR_DATA'),
        ('_ARCHIVE_WORDS', 'ARCHIVE_WORDS'),
        ('_CONSULTANT_TITLES', 'CONSULTANT_TITLES'),
        ('_EXEC_SUMMARY_BULLETS', 'EXEC_SUMMARY_BULLETS'),
        ('_FOOTNOTE_TEMPLATES', 'ARCHIVE_FOOTNOTE_TEMPLATES'),
        ('_REVISION_TYPES', 'REVISION_TYPES'),
        ('_DISTRIBUTION_CLASSES', 'DISTRIBUTION_CLASSES'),
        ('_ENGAGEMENT_CODES', 'ENGAGEMENT_CODES'),
        ('_BENCH_METRICS', 'BENCH_METRICS'),
        ('_PEER_GROUPS', 'PEER_GROUPS'),
        ('_DOC_VERSIONS', 'ARCHIVE_DOC_VERSIONS'),
    ]),
    (archive_data_compliance, [
        ('_AUDIT_REF_PREFIXES', 'AUDIT_REF_PREFIXES'),
        ('_COMPLIANCE_FRAMEWORKS', 'COMPLIANCE_FRAMEWORKS'),
        ('_COMPLIANCE_FINDING_TYPES', 'COMPLIANCE_FINDING_TYPES'),
        ('_COMPLIANCE_RISK_LEVELS', 'COMPLIANCE_RISK_LEVELS'),
        ('_COMPLIANCE_STATUSES', 'COMPLIANCE_STATUSES'),
        ('_COMPLIANCE_SCOPE_TEMPLATES', 'COMPLIANCE_SCOPE_TEMPLATES'),
        ('_COMPLIANCE_METHODOLOGY_TEMPLATES', 'COMPLIANCE_METHODOLOGY_TEMPLATES'),
        ('_CORRECTIVE_ACTION_TEMPLATES', 'CORRECTIVE_ACTION_TEMPLATES'),
        ('_MGMT_RESPONSE_TEMPLATES', 'MGMT_RESPONSE_TEMPLATES'),
        ('_PROJECT_NAMES', 'PROJECT_NAMES'),
        ('_DOC_VERSIONS', 'COMPLIANCE_DOC_VERSIONS'),
        ('_COMPLIANCE_TITLE_PREFIXES', 'COMPLIANCE_TITLE_PREFIXES'),
    ]),
    (archive_data_minutes, [
        ('_COMMITTEE_NAMES', 'COMMITTEE_NAMES'),
        ('_MEETING_LOCATIONS', 'MEETING_LOCATIONS'),
        ('_COMMITTEE_ROLES', 'COMMITTEE_ROLES'),
        ('_AGENDA_ITEM_TITLES', 'AGENDA_ITEM_TITLES'),
        ('_AGENDA_DISCUSSION_TEMPLATES', 'AGENDA_DISCUSSION_TEMPLATES'),
        ('_RESOLUTION_TEMPLATES', 'RESOLUTION_TEMPLATES'),
        ('_MOTION_VERBS', 'MOTION_VERBS'),
        ('_ACTION_ITEM_TEMPLATES', 'ACTION_ITEM_TEMPLATES'),
    ]),
    (policy_data, [
        ('AGENCIES', 'AGENCIES'),
        ('POLICY_SLUGS', 'POLICY_SLUGS'),
        ('DOCUMENT_TYPES', 'DOCUMENT_TYPES'),
        ('SIGNATORY_TITLES', 'SIGNATORY_TITLES'),
        ('CREDENTIALS', 'CREDENTIALS'),
        ('LEGISLATION', 'LEGISLATION'),
        ('SUMMARY_TEMPLATES', 'SUMMARY_TEMPLATES'),
        ('SECTION_HEADINGS', 'SECTION_HEADINGS'),
        ('_OPTIONAL_SECTION_POOL', 'OPTIONAL_SECTION_POOL'),
        ('PARAGRAPH_TEMPLATES', 'PARAGRAPH_TEMPLATES'),
        ('RECOMMENDATION_TEMPLATES', 'RECOMMENDATION_TEMPLATES'),
        ('POSITIONS', 'POSITIONS'),
        ('_MONTHS_LONG', 'MONTHS_LONG'),
        ('FOOTNOTE_TEMPLATES', 'POLICY_FOOTNOTE_TEMPLATES'),
        ('_STUB_TITLE_PREFIXES', 'STUB_TITLE_PREFIXES'),
        ('_FEATURED_SEEDS', 'FEATURED_SEEDS'),
        ('_CEO_NAMES', 'CEO_NAMES'),
        ('_YEAR_ERA_THEMES', 'YEAR_ERA_THEMES'),
        ('_CEO_MESSAGE_TEMPLATES', 'CEO_MESSAGE_TEMPLATES'),
        ('_YEAR_ANNUAL_LETTERS', 'YEAR_ANNUAL_LETTERS'),
        ('_EXPERT_TYPES', 'EXPERT_TYPES'),
        ('_INDUSTRY_SECTORS', 'INDUSTRY_SECTORS'),
        ('_TIMEFRAMES', 'TIMEFRAMES'),
        ('_COMPARISON_GROUPS', 'COMPARISON_GROUPS'),
        ('_FINDINGS_BRIEF', 'FINDINGS_BRIEF'),
    ]),
    (people_generators, [
        ('FIRST_NAMES', 'FIRST_NAMES'),
        ('LAST_NAMES', 'LAST_NAMES'),
        ('TITLES', 'PEOPLE_TITLES'),
    ]),
]


def _to_jsonable(value):
    """Preserve tuple-vs-list shape info: tuples become {"__tuple__": [...]}
    so the Go side can distinguish fixed-arity records from lists if needed."""
    if isinstance(value, dict):
        return {str(k): _to_jsonable(v) for k, v in value.items()}
    if isinstance(value, tuple):
        return {'__tuple__': [_to_jsonable(v) for v in value]}
    if isinstance(value, list):
        return [_to_jsonable(v) for v in value]
    return value


class Command(BaseCommand):
    help = 'Export literal archive/policy data pools to JSON for the Go service to embed.'

    def add_arguments(self, parser):
        parser.add_argument('--out', default=None, help='Output directory (default: acpwb_go/data relative to repo root)')

    def handle(self, *args, **options):
        # Only ./acpwb is bind-mounted into the web container, so the
        # sibling acpwb_go/ directory on the host is never reachable from
        # in here — write somewhere in-container and docker cp it out.
        out_dir = Path(options['out'] or '/tmp/render_data_export')
        out_dir.mkdir(parents=True, exist_ok=True)

        manifest = {}
        for module, names in EXPORTS:
            for source_name, export_name in names:
                value = getattr(module, source_name)
                jsonable = _to_jsonable(value)
                filename = f'{export_name}.json'
                (out_dir / filename).write_text(json.dumps(jsonable, ensure_ascii=False, indent=None))
                manifest[export_name] = filename
                self.stdout.write(f'  {module.__name__}.{source_name} -> {export_name} ({len(jsonable) if hasattr(jsonable, "__len__") else "?"} entries)')

        (out_dir / '_manifest.json').write_text(json.dumps(manifest, indent=2, sort_keys=True))
        self.stdout.write(self.style.SUCCESS(
            f'Exported {len(manifest)} data pools to {out_dir}'
        ))
