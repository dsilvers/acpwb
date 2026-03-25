"""
ACPWB Activity Dashboard — requires staff login.

Stats are pre-computed hourly by the `precalc_dashboard` management command and
stored in Redis.  Views serve from cache; on a cache miss they compute live and
re-populate the cache.  Custom date ranges always run live.

Cache keys:  dashboard:{view}:{preset}  (30 keys: 5 views × 6 presets)
TTL:         5400 s (90 minutes)
"""
from collections import Counter
from datetime import datetime, timedelta

from django.contrib.admin.views.decorators import staff_member_required
from django.core.cache import cache
from django.db.models import Count, Max
from django.db.models.functions import TruncDate
from django.shortcuts import render
from django.utils import timezone


from apps.honeypot.models import CrawlerVisit, ArchiveVisit, InternalLoginAttempt
from apps.people.models import PeoplePageVisit
from apps.projects.models import ProjectPageVisit
from apps.public.models import DataOptOutRequest
from apps.webhooks.models import InboundEmail

_DASH_CACHE_TTL = 5400  # 90 minutes

PRESETS = [
    ('today',  'Today'),
    ('7d',     'Last 7 Days'),
    ('30d',    'Last 30 Days'),
    ('90d',    'Last 90 Days'),
    ('ytd',    'Year to Date'),
    ('all',    'All Time'),
    ('custom', 'Custom Range'),
]

# Presets that can be pre-computed (no request context required)
PRECALC_PRESETS = ['today', '7d', '30d', '90d', 'ytd', 'all']


# ── Date range helpers ────────────────────────────────────────────────────────

def _build_date_range(preset):
    """Build a date range dict from a preset string without a request object.

    Used by the precalc_dashboard management command and by _parse_date_range.
    """
    now = timezone.now()

    if preset == 'today':
        start = now.replace(hour=0, minute=0, second=0, microsecond=0)
        end = now
    elif preset == '7d':
        start = now - timedelta(days=7)
        end = now
    elif preset == '30d':
        start = now - timedelta(days=30)
        end = now
    elif preset == '90d':
        start = now - timedelta(days=90)
        end = now
    elif preset == 'ytd':
        start = now.replace(month=1, day=1, hour=0, minute=0, second=0, microsecond=0)
        end = now
    elif preset == 'all':
        start = None
        end = None
    else:
        start = now - timedelta(days=30)
        end = now
        preset = '30d'

    return {
        'preset': preset,
        'start': start,
        'end': end,
        'from_str': start.strftime('%Y-%m-%d') if start else '',
        'to_str': end.strftime('%Y-%m-%d') if end else '',
        'presets': PRESETS,
    }


def _parse_date_range(request):
    preset = request.GET.get('range', '30d')

    if preset == 'custom':
        now = timezone.now()
        from_str = request.GET.get('from', '')
        to_str = request.GET.get('to', '')
        try:
            start = datetime.strptime(from_str, '%Y-%m-%d').replace(
                hour=0, minute=0, second=0, tzinfo=timezone.utc)
            end = datetime.strptime(to_str, '%Y-%m-%d').replace(
                hour=23, minute=59, second=59, tzinfo=timezone.utc)
        except (ValueError, TypeError):
            start = now - timedelta(days=30)
            end = now
            preset = '30d'
        return {
            'preset': preset,
            'start': start,
            'end': end,
            'from_str': start.strftime('%Y-%m-%d') if start else '',
            'to_str': end.strftime('%Y-%m-%d') if end else '',
            'presets': PRESETS,
        }

    return _build_date_range(preset)


def _apply_range(qs, date_range, field='timestamp'):
    if date_range['start']:
        qs = qs.filter(**{f'{field}__gte': date_range['start']})
    if date_range['end']:
        qs = qs.filter(**{f'{field}__lte': date_range['end']})
    return qs


def _bot_breakdown(qs, limit=20):
    """GROUP BY bot_type in SQL — O(1). Only for CrawlerVisit querysets."""
    rows = (qs.values('bot_type')
              .annotate(count=Count('id'))
              .order_by('-count')[:limit])
    total = sum(r['count'] for r in rows) or 1
    result = [
        {
            'name': r['bot_type'] or 'Unknown',
            'count': r['count'],
            'pct': round(r['count'] * 100 / total),
        }
        for r in rows
    ]
    return result, total


def _bot_breakdown_ua(qs, limit=20):
    """Python-loop fallback for models that only have user_agent (no bot_type column)."""
    from apps.core.bot_classify import classify_ua
    counts = Counter()
    for ua in qs.values_list('user_agent', flat=True):
        counts[classify_ua(ua or '')] += 1
    top = counts.most_common(limit)
    total = sum(counts.values()) or 1
    return [
        {'name': name, 'count': count, 'pct': round(count * 100 / total)}
        for name, count in top
    ], total


def _bot_group_breakdown(qs):
    """GROUP BY bot_group in SQL — O(1) instead of O(n) Python loop."""
    rows = (qs.values('bot_group')
              .annotate(count=Count('id'))
              .order_by('-count'))
    total = sum(r['count'] for r in rows) or 1
    return [
        {
            'name': r['bot_group'] or 'Unknown',
            'count': r['count'],
            'pct': round(r['count'] * 100 / total),
        }
        for r in rows
    ]


def _daily_chart(qs, days=30, field='timestamp'):
    """Single GROUP BY date SQL query instead of fetching all timestamps."""
    now = timezone.now()
    start = now - timedelta(days=days)
    rows = (qs.filter(**{f'{field}__gte': start})
              .annotate(date=TruncDate(field))
              .values('date')
              .annotate(count=Count('id'))
              .order_by('date'))
    counts = {r['date'].isoformat(): r['count'] for r in rows}
    bars = []
    for i in range(days - 1, -1, -1):
        d = (now - timedelta(days=i)).date().isoformat()
        bars.append({'date': d, 'count': counts.get(d, 0)})
    peak = max((r['count'] for r in bars), default=1) or 1
    for r in bars:
        r['pct'] = round(r['count'] * 100 / peak)
    return {
        'bars': bars,
        'start': bars[0]['date'] if bars else '',
        'end': bars[-1]['date'] if bars else '',
    }


def _get_daily_chart(redis_key, qs_callable, days, field='timestamp'):
    """Return a cached daily chart, computing and caching it on miss.

    redis_key   — e.g. 'dashboard:daily:crawlers:60'
    qs_callable — zero-arg callable returning the base queryset (avoids passing
                  a live queryset that may be evaluated before it's needed)
    """
    result = cache.get(redis_key)
    if result is None:
        result = _daily_chart(qs_callable(), days=days, field=field)
        cache.set(redis_key, result, _DASH_CACHE_TTL)
    return result


# ── Compute helpers (called by views on cache miss and by precalc_dashboard) ──

def _compute_overview(dr, daily=None):
    crawler_qs = _apply_range(CrawlerVisit.objects.all(), dr)
    archive_qs = _apply_range(ArchiveVisit.objects.all(), dr)
    email_qs   = _apply_range(InboundEmail.objects.all(), dr, field='received_at')
    people_qs  = _apply_range(PeoplePageVisit.objects.all(), dr)
    project_qs = _apply_range(ProjectPageVisit.objects.all(), dr)
    login_qs   = _apply_range(InternalLoginAttempt.objects.all(), dr, field='created_at')
    optout_qs  = _apply_range(DataOptOutRequest.objects.all(), dr, field='created_at')

    top_bots, _ = _bot_breakdown(crawler_qs, limit=15)

    trap_counts = list(
        crawler_qs.values('trap_type')
        .annotate(count=Count('id'))
        .order_by('-count')
    )
    trap_total = sum(t['count'] for t in trap_counts) or 1
    for t in trap_counts:
        t['pct'] = round(t['count'] * 100 / trap_total)
        t['label'] = dict(CrawlerVisit.TRAP_CHOICES).get(t['trap_type'], t['trap_type'])

    if daily is None:
        daily = _get_daily_chart(
            'dashboard:daily:overview:30',
            CrawlerVisit.objects.all, days=30,
        )

    return {
        **dr,
        'counts': {
            'crawler':  crawler_qs.count(),
            'archive':  archive_qs.count(),
            'email':    email_qs.count(),
            'people':   people_qs.count(),
            'projects': project_qs.count(),
            'logins':   login_qs.count(),
            'optouts':  optout_qs.count(),
        },
        'top_bots':        top_bots,
        'trap_counts':     trap_counts,
        'bot_groups':      _bot_group_breakdown(crawler_qs),
        'daily':           daily,
        'recent_crawlers': list(crawler_qs.order_by('-timestamp')[:10]),
        'recent_emails':   list(email_qs.order_by('-received_at')[:5]),
        'recent_optouts':  list(DataOptOutRequest.objects.order_by('-created_at')[:10]),
        'cached_at':       timezone.now(),
    }


def _compute_crawlers(dr, daily=None):
    qs = _apply_range(CrawlerVisit.objects.all(), dr)
    top_bots, total = _bot_breakdown(qs, limit=30)

    trap_counts = list(
        qs.values('trap_type')
        .annotate(count=Count('id'))
        .order_by('-count')
    )
    trap_total = sum(t['count'] for t in trap_counts) or 1
    for t in trap_counts:
        t['pct'] = round(t['count'] * 100 / trap_total)
        t['label'] = dict(CrawlerVisit.TRAP_CHOICES).get(t['trap_type'], t['trap_type'])

    if daily is None:
        daily = _get_daily_chart(
            'dashboard:daily:crawlers:60',
            CrawlerVisit.objects.all, days=60,
        )

    return {
        **dr,
        'total':       total,
        'top_bots':    top_bots,
        'trap_counts': trap_counts,
        'top_ips':     list(qs.values('ip_address').annotate(count=Count('id')).order_by('-count')[:20]),
        'top_paths':   list(qs.values('path').annotate(count=Count('id')).order_by('-count')[:20]),
        'top_hosts':   list(qs.exclude(host='').values('host').annotate(count=Count('id')).order_by('-count')[:20]),
        'daily':       daily,
        'recent':      list(qs.order_by('-timestamp').select_related()[:50]),
        'cached_at':   timezone.now(),
    }


def _compute_archive(dr, daily=None):
    qs = _apply_range(ArchiveVisit.objects.all(), dr)
    top_bots, total = _bot_breakdown_ua(qs, limit=20)

    depth_counts = list(
        qs.values('depth')
        .annotate(count=Count('id'))
        .order_by('depth')[:15]
    )
    depth_total = sum(d['count'] for d in depth_counts) or 1
    for d in depth_counts:
        d['pct'] = round(d['count'] * 100 / depth_total)

    if daily is None:
        daily = _get_daily_chart(
            'dashboard:daily:archive:30',
            ArchiveVisit.objects.all, days=30,
        )

    return {
        **dr,
        'total':        total,
        'top_bots':     top_bots,
        'depth_counts': depth_counts,
        'top_ips':      list(qs.values('ip_address').annotate(count=Count('id'), max_depth=Max('depth')).order_by('-count')[:20]),
        'top_roots':    list(qs.values('slug').annotate(count=Count('id')).order_by('-count')[:20]),
        'daily':        daily,
        'recent':       list(qs.order_by('-timestamp')[:50]),
        'cached_at':    timezone.now(),
    }


def _compute_emails(dr, daily=None):
    qs = _apply_range(InboundEmail.objects.all(), dr, field='received_at')
    total = qs.count()

    domain_counts = Counter()
    for sender in qs.values_list('sender', flat=True):
        domain = sender.split('@')[-1].lower() if '@' in sender else sender
        domain_counts[domain] += 1
    top_domains = [
        {'domain': d, 'count': c, 'pct': round(c * 100 / total) if total else 0}
        for d, c in domain_counts.most_common(20)
    ]

    if daily is None:
        daily = _get_daily_chart(
            'dashboard:daily:emails:30',
            InboundEmail.objects.all, days=30, field='received_at',
        )

    return {
        **dr,
        'total':           total,
        'top_domains':     top_domains,
        'top_recipients':  list(qs.values('recipient').annotate(count=Count('id')).order_by('-count')[:20]),
        'daily':           daily,
        'recent':          list(qs.order_by('-received_at').prefetch_related('matches')[:50]),
        'cached_at':       timezone.now(),
    }


def _compute_people(dr, people_daily=None, project_daily=None):
    people_qs  = _apply_range(PeoplePageVisit.objects.all(), dr)
    project_qs = _apply_range(ProjectPageVisit.objects.all(), dr)

    people_bots, people_total   = _bot_breakdown_ua(people_qs, limit=20)
    project_bots, project_total = _bot_breakdown_ua(project_qs, limit=20)

    if people_daily is None:
        people_daily = _get_daily_chart(
            'dashboard:daily:people:30',
            PeoplePageVisit.objects.all, days=30,
        )
    if project_daily is None:
        project_daily = _get_daily_chart(
            'dashboard:daily:projects:30',
            ProjectPageVisit.objects.all, days=30,
        )

    return {
        **dr,
        'people_total':    people_total,
        'project_total':   project_total,
        'people_bots':     people_bots,
        'project_bots':    project_bots,
        'top_people_ips':  list(people_qs.values('ip_address').annotate(count=Count('id')).order_by('-count')[:15]),
        'top_project_ips': list(project_qs.values('ip_address').annotate(count=Count('id')).order_by('-count')[:15]),
        'people_daily':    people_daily,
        'project_daily':   project_daily,
        'recent_people':   list(people_qs.order_by('-timestamp')[:30]),
        'recent_projects': list(project_qs.order_by('-timestamp')[:30]),
        'cached_at':       timezone.now(),
    }


# ── Views ─────────────────────────────────────────────────────────────────────

def _cached_view(request, view_name, compute_fn, template):
    dr = _parse_date_range(request)
    preset = dr['preset']
    if preset != 'custom':
        cache_key = f'dashboard:{view_name}:{preset}'
        ctx = cache.get(cache_key)
        if ctx is not None:
            return render(request, template, ctx)
        ctx = compute_fn(dr)
        cache.set(cache_key, ctx, _DASH_CACHE_TTL)
    else:
        ctx = compute_fn(dr)
    return render(request, template, ctx)


@staff_member_required(login_url='/django-admin/login/')
def overview(request):
    return _cached_view(request, 'overview', _compute_overview, 'dashboard/overview.html')


@staff_member_required(login_url='/django-admin/login/')
def crawlers(request):
    return _cached_view(request, 'crawlers', _compute_crawlers, 'dashboard/crawlers.html')


@staff_member_required(login_url='/django-admin/login/')
def archive(request):
    return _cached_view(request, 'archive', _compute_archive, 'dashboard/archive.html')


@staff_member_required(login_url='/django-admin/login/')
def emails(request):
    return _cached_view(request, 'emails', _compute_emails, 'dashboard/emails.html')


@staff_member_required(login_url='/django-admin/login/')
def people(request):
    return _cached_view(request, 'people', _compute_people, 'dashboard/people.html')
