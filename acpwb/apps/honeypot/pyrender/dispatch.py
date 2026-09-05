"""Phase 4 wiring: decide, per request, whether to serve the Python-built
render path or the original Django/Jinja2 template path, and produce the
Python path's HTML when chosen.

See /Users/dan/.claude/plans/any-performance-benefits-to-dreamy-deer.md
Phase 4 — both paths stay live simultaneously behind settings flags (one
per converted view, so they can roll out independently) plus a DEBUG-only
`?__render=python|django` per-request override, mirroring the existing
`?__year=`/`?__agency=` DEBUG shortcuts in apps/core/subdomain_middleware.py.
"""
from django.conf import settings
from django.template.loader import render_to_string


def use_python_render(request, setting_name):
    if settings.DEBUG:
        override = request.GET.get('__render')
        if override == 'python':
            return True
        if override == 'django':
            return False
    return getattr(settings, setting_name, False)


def render_archive_page_python(request, context, variant, on_sub):
    """variant is 'default', 'compliance', or 'minutes' — matches the
    _variant_int branch already computed by archive_trap()."""
    from apps.honeypot.pyrender import archive_era, archive_main

    if not on_sub:
        if variant == 'compliance':
            content_html = archive_main.render_compliance_default(context)
            needs_presentations_css = False
        elif variant == 'minutes':
            content_html = archive_main.render_minutes_default(context)
            needs_presentations_css = False
        else:
            content_html = archive_main.render_archive_default(context)
            needs_presentations_css = True
        # request= is required here (not just a context dict) so the Django
        # backend's context_processors — honeypot_context in particular —
        # populate honeypot_token/site_root for the inherited base.html,
        # exactly as they would via the old path's render() shortcut.
        return render_to_string('honeypot/_archive_content_shell.html', {
            'content_html': content_html,
            'og_title': context['og_title'],
            'needs_presentations_css': needs_presentations_css,
        }, request=request)

    year, month, day = context['year'], context['month'], context['day']
    if variant == 'compliance':
        era_content_html = archive_era.render_compliance_default_era(context)
        title_suffix = 'ACPWB Compliance Archive'
        og_description = (
            f'{context["industry"]} sector compliance review archived '
            f'{year}-{month:02d}-{day:02d}. Audit ref {context["audit_ref"]}. '
            f'ACPWB Regulatory Practice.'
        )
    elif variant == 'minutes':
        era_content_html = archive_era.render_minutes_default_era(context)
        title_suffix = 'ACPWB Archive'
        og_description = (
            f'{context["committee"]} meeting minutes archived '
            f'{year}-{month:02d}-{day:02d}. Meeting ref {context["meeting_ref"]}. '
            f'ACPWB Institutional Records.'
        )
    else:
        era_content_html = archive_era.render_archive_default_era(context)
        title_suffix = 'ACPWB Archive'
        og_description = (
            f'{context["industry"]} sector engagement documentation archived '
            f'{year}-{month:02d}-{day:02d}. {context["phase"].capitalize()} phase record. '
            f'ACPWB Research Division.'
        )
    return render_to_string('honeypot/_archive_era_content_shell.html', {
        'title': context['title'],
        'title_suffix': title_suffix,
        'og_description': og_description,
        'era_content_html': era_content_html,
        'request': context['request'],
        'year_data': context['year_data'],
        'year': year,
        'all_years': context['all_years'],
    })


def render_policy_detail_python(context):
    from apps.honeypot.pyrender import policy
    return policy.render_policy_detail(context)


def render_policy_index_python(context):
    from apps.honeypot.pyrender import policy
    return policy.render_policy_index(context)


def render_policy_year_python(context):
    from apps.honeypot.pyrender import policy
    return policy.render_policy_year(context)


def render_policy_month_python(context):
    from apps.honeypot.pyrender import policy
    return policy.render_policy_month(context)


def render_policy_subdomain_index_python(context):
    from apps.honeypot.pyrender import policy
    return policy.render_policy_subdomain_index(context)


def render_policy_subdomain_year_python(context):
    from apps.honeypot.pyrender import policy
    return policy.render_policy_subdomain_year(context)
