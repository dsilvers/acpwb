from django.utils.text import slugify

from apps.honeypot.archive_data import _ARCHIVE_ORGS

# Deduplicate by slug while preserving order
_seen_slugs: set[str] = set()
ORGANIZATIONS: list[str] = []
for _name in _ARCHIVE_ORGS:
    _slug = slugify(_name)
    if _slug not in _seen_slugs:
        _seen_slugs.add(_slug)
        ORGANIZATIONS.append(_name)

ORG_SLUG_MAP = {slugify(name): name for name in ORGANIZATIONS}
ORG_SLUGS = list(ORG_SLUG_MAP.keys())
