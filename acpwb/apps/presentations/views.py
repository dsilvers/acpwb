import re

from django.shortcuts import render, redirect
from django.urls import reverse
from django.http import Http404, HttpResponse
# Imported at module load (main thread, process boot) rather than lazily
# inside _render_pdf: weasyprint's first import does a one-time
# ctypes.util.find_library() that shells out via subprocess, and gevent's
# child watcher only works on the default loop. _render_pdf runs on
# gevent's threadpool (see run_in_thread below), which has no default loop,
# so a first-time import there crashes with "child watchers are only
# available on the default loop".
from weasyprint import HTML

from .image_selector import _STATIC_ROOT


def _get_ip(request):
    x_forwarded = request.META.get('HTTP_X_FORWARDED_FOR')
    if x_forwarded:
        return x_forwarded.split(',')[0].strip()
    return request.META.get('REMOTE_ADDR', '0.0.0.0')


def _log_crawler(request):
    try:
        from apps.core.crawler_queue import queue_crawler_visit
        from apps.core.bot_classify import bot_type_to_group, classify_ua_or_ip
        from django.utils import timezone
        ua = request.META.get('HTTP_USER_AGENT', '')
        ip = _get_ip(request)
        bot_type = classify_ua_or_ip(ua, ip)
        data = {
            'timestamp': timezone.now().isoformat(),
            'ip_address': ip,
            'user_agent': ua[:512],
            'host': request.get_host()[:253],
            'path': request.path[:512],
            'referrer': request.META.get('HTTP_REFERER', '')[:256],
            'trap_type': 'presentation',
            'query_string': request.META.get('QUERY_STRING', '')[:256],
            'bot_type': bot_type,
            'bot_group': bot_type_to_group(bot_type),
        }
        queue_crawler_visit(data)
    except Exception:
        pass


from .generators import (
    generate_presentation_meta,
    generate_slide,
    generate_presentations_for_context,
    generate_landing_orgs,
    generate_related,
)
from .data.organizations import ORG_SLUG_MAP

_SLUG_RE = re.compile(r'^[a-z][a-z0-9-]{4,200}-\d{4}$')


def _validate_presentation(org_slug, year, month, day, slug):
    if org_slug not in ORG_SLUG_MAP:
        raise Http404
    try:
        year, month, day = int(year), int(month), int(day)
    except (TypeError, ValueError):
        raise Http404
    if not (1990 <= year <= 2030 and 1 <= month <= 12 and 1 <= day <= 31):
        raise Http404
    if not _SLUG_RE.match(slug):
        raise Http404
    return year, month, day


def presentation_landing(request):
    _log_crawler(request)
    page = max(1, int(request.GET.get('page', 1) or 1))
    recent = generate_presentations_for_context(f'presrecent_main_p{page}', count=8)
    org_sections = generate_landing_orgs(page)
    return render(request, 'presentations/landing.html', {
        'recent': recent,
        'org_sections': org_sections,
        'page': page,
        'prev_page': page - 1 if page > 1 else None,
        'next_page': page + 1,
        'og_title': 'Research Presentations & Industry Briefings — ACPWB',
        'og_description': (
            'Access hundreds of research presentations, industry briefings, and '
            'executive slide decks from leading organizations.'
        ),
    })


def org_page(request, org_slug):
    if org_slug not in ORG_SLUG_MAP:
        raise Http404
    _log_crawler(request)
    org_name = ORG_SLUG_MAP[org_slug]
    page = max(1, int(request.GET.get('page', 1) or 1))
    presentations = generate_presentations_for_context(
        f'presorg_{org_slug}_p{page}', count=12
    )
    return render(request, 'presentations/org.html', {
        'org_name': org_name,
        'org_slug': org_slug,
        'presentations': presentations,
        'page': page,
        'prev_page': page - 1 if page > 1 else None,
        'next_page': page + 1,
        'og_title': f'{org_name} Presentations — ACPWB',
        'og_description': (
            f'Research presentations and industry briefings from {org_name}.'
        ),
    })


def presentation_detail(request, org_slug, year, month, day, slug):
    year, month, day = _validate_presentation(org_slug, year, month, day, slug)
    _log_crawler(request)
    slide_num = 1
    pres_meta = generate_presentation_meta(org_slug, year, month, day, slug)
    slide = generate_slide(pres_meta, slide_num)
    same_org, related = generate_related(pres_meta['pres_seed'], org_slug, pres_meta)
    next_url = (
        reverse('presentation-slide', args=[org_slug, year, f'{month:02d}', f'{day:02d}', slug, 2])
        if pres_meta['slide_count'] > 1 else None
    )
    present_url = reverse('presentation-present', args=[org_slug, year, f'{month:02d}', f'{day:02d}', slug, 1])
    return render(request, 'presentations/slide.html', {
        'pres': pres_meta,
        'slide': slide,
        'slide_num': slide_num,
        'prev_url': None,
        'next_url': next_url,
        'present_url': present_url,
        'same_org': same_org,
        'related': related,
        'progress_pct': round(1 / pres_meta['slide_count'] * 100),
        'og_title': pres_meta['title'],
        'og_description': pres_meta['subtitle'],
    })


def presentation_slide(request, org_slug, year, month, day, slug, slide_num):
    year, month, day = _validate_presentation(org_slug, year, month, day, slug)
    try:
        slide_num = int(slide_num)
    except (TypeError, ValueError):
        raise Http404

    _log_crawler(request)
    pres_meta = generate_presentation_meta(org_slug, year, month, day, slug)

    if slide_num < 1 or slide_num > pres_meta['slide_count']:
        raise Http404

    slide = generate_slide(pres_meta, slide_num)
    same_org, related = generate_related(pres_meta['pres_seed'], org_slug, pres_meta)

    prev_url = (
        reverse('presentation-slide', args=[org_slug, year, f'{month:02d}', f'{day:02d}', slug, slide_num - 1])
        if slide_num > 1 else None
    )
    next_url = (
        reverse('presentation-slide', args=[org_slug, year, f'{month:02d}', f'{day:02d}', slug, slide_num + 1])
        if slide_num < pres_meta['slide_count'] else None
    )
    present_url = reverse('presentation-present', args=[org_slug, year, f'{month:02d}', f'{day:02d}', slug, slide_num])

    return render(request, 'presentations/slide.html', {
        'pres': pres_meta,
        'slide': slide,
        'slide_num': slide_num,
        'prev_url': prev_url,
        'next_url': next_url,
        'present_url': present_url,
        'same_org': same_org,
        'related': related,
        'progress_pct': round(slide_num / pres_meta['slide_count'] * 100),
        'og_title': f"{pres_meta['title']} — Slide {slide_num}",
        'og_description': pres_meta['subtitle'],
    })


def _render_pdf(html_string, base_url):
    return HTML(string=html_string, base_url=base_url).write_pdf()


def presentation_download_pdf(request, org_slug, year, month, day, slug):
    year, month, day = _validate_presentation(org_slug, year, month, day, slug)
    _log_crawler(request)
    pres_meta = generate_presentation_meta(org_slug, year, month, day, slug)
    slides = [generate_slide(pres_meta, n) for n in range(1, pres_meta['slide_count'] + 1)]

    from django.template.loader import render_to_string
    html_string = render_to_string('presentations/presentation_print.html', {
        'pres': pres_meta,
        'slides': slides,
    }, request=request)
    # base_url lets <img src="img/presentations/..."> resolve straight to disk
    # instead of base64-encoding into the HTML string — avoids inflating the
    # payload WeasyPrint has to parse/decode by ~33% per embedded image.
    from apps.core.async_utils import run_in_thread
    # weasyprint's layout/render is CPU + C-library (cairo/pango) work that
    # doesn't yield on gevent's event loop — running it inline would stall
    # every other concurrent connection on this worker for the duration.
    pdf_bytes = run_in_thread(_render_pdf, html_string, f'{_STATIC_ROOT}/')
    filename = f"{slug}-{year}-{month:02d}-{day:02d}.pdf"
    resp = HttpResponse(pdf_bytes, content_type='application/pdf')
    resp['Content-Disposition'] = f'attachment; filename="{filename}"'
    return resp


def presentation_download_pptx(request, org_slug, year, month, day, slug):
    year, month, day = _validate_presentation(org_slug, year, month, day, slug)
    _log_crawler(request)
    pres_meta = generate_presentation_meta(org_slug, year, month, day, slug)
    slides = [generate_slide(pres_meta, n) for n in range(1, pres_meta['slide_count'] + 1)]
    from .pptx_export import generate_pptx_bytes
    pptx_bytes = generate_pptx_bytes(pres_meta, slides)
    filename = f"{slug}-{year}-{month:02d}-{day:02d}.pptx"
    resp = HttpResponse(pptx_bytes, content_type='application/vnd.openxmlformats-officedocument.presentationml.presentation')
    resp['Content-Disposition'] = f'attachment; filename="{filename}"'
    return resp


def presentation_present(request, org_slug, year, month, day, slug, slide_num):
    year, month, day = _validate_presentation(org_slug, year, month, day, slug)
    try:
        slide_num = int(slide_num)
    except (TypeError, ValueError):
        raise Http404

    _log_crawler(request)
    pres_meta = generate_presentation_meta(org_slug, year, month, day, slug)

    if slide_num < 1 or slide_num > pres_meta['slide_count']:
        raise Http404

    slide = generate_slide(pres_meta, slide_num)

    prev_url = (
        reverse('presentation-present', args=[org_slug, year, f'{month:02d}', f'{day:02d}', slug, slide_num - 1])
        if slide_num > 1 else None
    )
    next_url = (
        reverse('presentation-present', args=[org_slug, year, f'{month:02d}', f'{day:02d}', slug, slide_num + 1])
        if slide_num < pres_meta['slide_count'] else None
    )
    exit_url = reverse('presentation-slide', args=[org_slug, year, f'{month:02d}', f'{day:02d}', slug, slide_num])

    return render(request, 'presentations/present.html', {
        'pres': pres_meta,
        'slide': slide,
        'slide_num': slide_num,
        'prev_url': prev_url,
        'next_url': next_url,
        'exit_url': exit_url,
        'progress_pct': round(slide_num / pres_meta['slide_count'] * 100),
    })
