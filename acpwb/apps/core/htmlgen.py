"""
Shared helpers for the hand-written Python HTML builders that replace
Django/Jinja2 template rendering for archive/policy pages (the
"raw-templates" performance work — see
/Users/dan/.claude/plans/any-performance-benefits-to-dreamy-deer.md for the
full plan and the profiling/benchmarks that motivated it).

Two patterns used throughout:

  - Fully static or near-static partials (a handful of {{ var }}
    substitutions, no loops — honeypot/_archive_seal.html,
    partials/_jsonld_garbage.html, partials/_ghost_links.html,
    partials/_prompt_injection.html) are rendered ONCE per process via the
    real template engine with sentinel placeholder tokens standing in for
    the few variables that ever change, cached as a plain string, then
    filled in per-request via cheap str.replace() calls. This is
    byte-identical to the original template's output (it IS that output,
    literally) while avoiding Django's per-node dispatch cost on every
    request — the whole point, since these partials are included on every
    archive/policy page and none of their content is actually loop-driven.

  - Genuinely per-item content (loops, conditionals, custom-tag calls —
    presentations/_pres_card.html) is built directly via f-strings/joins,
    calling the real underlying Python functions behind any custom
    template tags directly (they're plain @register.simple_tag functions)
    rather than reimplementing their logic by hand.
"""
from django.template.defaultfilters import truncatechars as _dj_truncatechars
from django.template.defaultfilters import truncatewords as _dj_truncatewords
from django.template.loader import render_to_string
from django.utils.html import escape

__all__ = [
    'escape', 'truncatewords', 'truncatechars',
    'get_archive_seal', 'get_jsonld_garbage', 'get_ghost_links', 'get_prompt_injection',
    'render_pres_card',
]


def truncatewords(text, n):
    return _dj_truncatewords(text, n)


def truncatechars(text, n):
    return _dj_truncatechars(text, n)


def _cached_static_partial(cache, template_name, sentinel_context, real_context):
    """Render `template_name` once (with sentinel placeholder values standing
    in for the keys in `real_context`), cache the result in `cache` (a
    single-entry dict used as a mutable cell), then substitute the real
    values via str.replace() on every subsequent call."""
    if 'template' not in cache:
        cache['template'] = render_to_string(template_name, sentinel_context)
    out = cache['template']
    for key, real_value in real_context.items():
        out = out.replace(sentinel_context[key], str(real_value))
    return out


_SEAL_CACHE = {}
_SEAL_SENTINELS = {'year': '__HTMLGEN_SEAL_YEAR__', 'record_id': '__HTMLGEN_SEAL_RECORD_ID__'}


def get_archive_seal(year, record_id):
    """Byte-identical to rendering honeypot/_archive_seal.html (338 lines,
    only `year`/`record_id` ever vary) without Django's per-node cost."""
    return _cached_static_partial(
        _SEAL_CACHE, 'honeypot/_archive_seal.html', _SEAL_SENTINELS,
        {'year': year, 'record_id': record_id},
    )


_JSONLD_CACHE = {}
_JSONLD_SENTINELS = {'honeypot_token': '__HTMLGEN_JSONLD_TOKEN__'}


def get_jsonld_garbage(honeypot_token):
    return _cached_static_partial(
        _JSONLD_CACHE, 'partials/_jsonld_garbage.html', _JSONLD_SENTINELS,
        {'honeypot_token': honeypot_token},
    )


_GHOST_LINKS_CACHE = {}


def get_ghost_links():
    """Fully static — no context variables at all."""
    if 'template' not in _GHOST_LINKS_CACHE:
        _GHOST_LINKS_CACHE['template'] = render_to_string('partials/_ghost_links.html', {})
    return _GHOST_LINKS_CACHE['template']


_PROMPT_INJECTION_CACHE = {}
_PROMPT_INJECTION_SENTINELS = {'honeypot_token': '__HTMLGEN_PROMPT_TOKEN__'}


def get_prompt_injection(honeypot_token):
    return _cached_static_partial(
        _PROMPT_INJECTION_CACHE, 'partials/_prompt_injection.html', _PROMPT_INJECTION_SENTINELS,
        {'honeypot_token': honeypot_token},
    )


def render_pres_card(pres, url_prefix=''):
    """Direct Python equivalent of presentations/_pres_card.html. Calls the
    real org_logo/headshot_or_avatar template-tag functions directly (plain
    @register.simple_tag functions under apps/core/templatetags/acpwb_tags.py)
    so their output is exactly what the template would have produced."""
    from django.templatetags.static import static

    from apps.core.templatetags.acpwb_tags import headshot_or_avatar, org_logo

    e = escape
    pres_url = f'{url_prefix}{pres["pres_url"]}'
    if pres.get('thumb_bg'):
        thumb_style = (
            f"background-image:url('{static(pres['thumb_bg'])}');"
            f"background-size:cover;background-position:center;"
        )
    else:
        thumb_style = f"background-color:{pres['theme']['bg']};"

    authors = pres.get('authors', [])
    authors_avatars = ''.join(
        str(headshot_or_avatar(a['avatar_seed'], a['initials'], 24)) for a in authors
    )
    authors_names = ', '.join(e(a['full_name']) for a in authors)

    return (
        f'<div class="pres-card">'
        f'<a href="{e(pres_url)}" style="text-decoration:none">'
        f'<div class="pres-card-thumb" style="{thumb_style}">'
        f'<div style="position:absolute;top:0.6em;right:0.7em;z-index:2;'
        f'background:{pres["theme"]["accent"]};color:{pres["theme"]["bg"]};font-size:.62rem;'
        f'padding:.2em .5em;border-radius:2px;font-weight:800;font-family:system-ui,sans-serif">'
        f'{pres["slide_count"]} slides</div></div></a>'
        f'<div class="pres-card-body">'
        f'<div class="pres-card-org" style="display:flex;align-items:center;gap:.4em">'
        f'{str(org_logo(pres["org_slug"], 22))}{e(pres["org_name"])}</div>'
        f'<a href="{e(pres_url)}" class="pres-card-title" style="text-decoration:none">'
        f'{e(truncatewords(pres["title"], 12))}</a>'
        f'<div class="pres-card-meta">{e(pres["pub_date_display"])} &mdash; {e(pres["industry"])}</div>'
        f'<div class="pres-card-authors">{authors_avatars}'
        f'<span style="font-size:.7rem;color:#555;margin-left:.3em">{authors_names}</span></div>'
        f'</div></div>'
    )
