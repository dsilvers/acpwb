"""
ACPWB Activity Dashboard — requires staff login.
"""
from collections import Counter, defaultdict
from datetime import datetime, timedelta

from django.contrib.admin.views.decorators import staff_member_required
from django.db.models import Count, Max
from django.shortcuts import render
from django.utils import timezone

from apps.honeypot.models import CrawlerVisit, ArchiveVisit
from apps.people.models import PeoplePageVisit
from apps.projects.models import ProjectPageVisit
from apps.webhooks.models import InboundEmail


# ── Bot classification ────────────────────────────────────────────────────────

BOT_PATTERNS = [
    # AI crawlers — most interesting
    ('GPTBot',              'OpenAI GPTBot'),
    ('OAI-SearchBot',       'OpenAI SearchBot'),
    ('ChatGPT-User',        'OpenAI ChatGPT'),
    ('ClaudeBot',           'Anthropic ClaudeBot'),
    ('Claude-Web',          'Anthropic Claude'),
    ('anthropic-ai',        'Anthropic'),
    ('PerplexityBot',       'Perplexity'),
    ('Google-Extended',     'Google-Extended (AI)'),
    ('FacebookBot',         'Meta FacebookBot'),
    ('Applebot-Extended',   'Apple Applebot-Extended'),
    ('Bytespider',          'ByteDance Bytespider'),
    # Search engines
    ('Googlebot',           'Googlebot'),
    ('bingbot',             'Bingbot'),
    ('BingPreview',         'Bing Preview'),
    ('msnbot',              'MSN Bot'),
    ('Baiduspider',         'Baiduspider'),
    ('YandexBot',           'YandexBot'),
    ('Slurp',               'Yahoo Slurp'),
    ('DuckDuckBot',         'DuckDuckBot'),
    ('Applebot',            'Applebot'),
    ('sogou',               'Sogou'),
    ('360Spider',           '360Spider'),
    ('SeznamBot',           'Seznam'),
    # SEO/marketing crawlers
    ('SemrushBot',          'SemrushBot'),
    ('AhrefsBot',           'AhrefsBot'),
    ('MJ12bot',             'Majestic MJ12'),
    ('DotBot',              'DotBot'),
    ('DataForSeoBot',       'DataForSEO'),
    ('PetalBot',            'Huawei PetalBot'),
    ('PiplBot',             'Pipl'),
    # Social
    ('Twitterbot',          'Twitterbot'),
    ('facebookexternalhit', 'Facebook Scraper'),
    ('LinkedInBot',         'LinkedIn'),
    # Archives
    ('ia_archiver',         'Internet Archive'),
    ('archive.org_bot',     'Internet Archive'),
    # Generic HTTP clients (likely scrapers/bots)
    ('python-requests',     'Python Requests'),
    ('curl/',               'cURL'),
    ('wget',                'Wget'),
    ('scrapy',              'Scrapy'),
    ('Go-http-client',      'Go HTTP Client'),
    ('Java/',               'Java HTTP Client'),
    ('libwww-perl',         'libwww-perl'),
    ('axios',               'axios'),
    ('node-fetch',          'node-fetch'),
    ('okhttp',              'OkHttp'),
    ('httpx',               'httpx'),
    ('aiohttp',             'aiohttp'),
    ('Faraday',             'Faraday (Ruby)'),
]


def classify_ua(ua):
    if not ua or not ua.strip():
        return '(empty user agent)'
    for pattern, name in BOT_PATTERNS:
        if pattern.lower() in ua.lower():
            return name
    return 'Other / Browser'


def classify_ua_group(ua):
    """Coarser grouping for overview charts."""
    label = classify_ua(ua)
    if label in ('Other / Browser', '(empty user agent)'):
        return label
    ai_bots = {'OpenAI GPTBot', 'OpenAI SearchBot', 'OpenAI ChatGPT',
                'Anthropic ClaudeBot', 'Anthropic Claude', 'Anthropic',
                'Perplexity', 'Google-Extended (AI)', 'ByteDance Bytespider',
                'Meta FacebookBot', 'Apple Applebot-Extended'}
    if label in ai_bots:
        return 'AI Crawlers'
    search_bots = {'Googlebot', 'Bingbot', 'Bing Preview', 'MSN Bot',
                   'Baiduspider', 'YandexBot', 'Yahoo Slurp', 'DuckDuckBot',
                   'Applebot', 'Sogou', '360Spider', 'Seznam'}
    if label in search_bots:
        return 'Search Engines'
    scraper_bots = {'Python Requests', 'cURL', 'Wget', 'Scrapy', 'Go HTTP Client',
                    'Java HTTP Client', 'libwww-perl', 'axios', 'node-fetch',
                    'OkHttp', 'httpx', 'aiohttp', 'Faraday (Ruby)'}
    if label in scraper_bots:
        return 'Generic Scrapers'
    return 'SEO / Other Bots'


# ── Date range helpers ────────────────────────────────────────────────────────

PRESETS = [
    ('today',  'Today'),
    ('7d',     'Last 7 Days'),
    ('30d',    'Last 30 Days'),
    ('90d',    'Last 90 Days'),
    ('ytd',    'Year to Date'),
    ('all',    'All Time'),
    ('custom', 'Custom Range'),
]


def _parse_date_range(request):
    preset = request.GET.get('range', '30d')
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
    elif preset == 'custom':
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


def _apply_range(qs, date_range, field='timestamp'):
    if date_range['start']:
        qs = qs.filter(**{f'{field}__gte': date_range['start']})
    if date_range['end']:
        qs = qs.filter(**{f'{field}__lte': date_range['end']})
    return qs


def _bot_breakdown(qs, ua_field='user_agent', limit=20):
    """Classify user agents from a queryset, return sorted list of (name, count)."""
    counts = Counter()
    for ua in qs.values_list(ua_field, flat=True):
        counts[classify_ua(ua or '')] += 1
    top = counts.most_common(limit)
    total = sum(counts.values())
    result = []
    for name, count in top:
        result.append({
            'name': name,
            'count': count,
            'pct': round(count * 100 / total) if total else 0,
        })
    return result, total


def _daily_chart(qs, days=30, field='timestamp'):
    """Return dict with bars list and start/end date strings."""
    now = timezone.now()
    day_counts = defaultdict(int)
    start = now - timedelta(days=days)
    for ts in qs.filter(**{f'{field}__gte': start}).values_list(field, flat=True):
        day_counts[ts.date().isoformat()] += 1
    bars = []
    for i in range(days - 1, -1, -1):
        d = (now - timedelta(days=i)).date().isoformat()
        bars.append({'date': d, 'count': day_counts[d]})
    peak = max((r['count'] for r in bars), default=1) or 1
    for r in bars:
        r['pct'] = round(r['count'] * 100 / peak)
    return {
        'bars': bars,
        'start': bars[0]['date'] if bars else '',
        'end': bars[-1]['date'] if bars else '',
    }


# ── Views ─────────────────────────────────────────────────────────────────────

@staff_member_required(login_url='/django-admin/login/')
def overview(request):
    dr = _parse_date_range(request)

    crawler_qs  = _apply_range(CrawlerVisit.objects.all(), dr)
    archive_qs  = _apply_range(ArchiveVisit.objects.all(), dr)
    email_qs    = _apply_range(InboundEmail.objects.all(), dr, field='received_at')
    people_qs   = _apply_range(PeoplePageVisit.objects.all(), dr)
    project_qs  = _apply_range(ProjectPageVisit.objects.all(), dr)

    # Top bots across all crawler visits
    top_bots, _ = _bot_breakdown(crawler_qs, limit=15)

    # Trap type breakdown
    trap_counts = list(
        crawler_qs.values('trap_type')
        .annotate(count=Count('id'))
        .order_by('-count')
    )
    trap_total = sum(t['count'] for t in trap_counts) or 1
    for t in trap_counts:
        t['pct'] = round(t['count'] * 100 / trap_total)
        t['label'] = dict(CrawlerVisit.TRAP_CHOICES).get(t['trap_type'], t['trap_type'])

    # Daily chart (always 30 days regardless of filter, for context)
    daily = _daily_chart(CrawlerVisit.objects.all(), days=30)

    # Bot group breakdown
    group_counts = Counter()
    for ua in crawler_qs.values_list('user_agent', flat=True):
        group_counts[classify_ua_group(ua or '')] += 1
    group_total = sum(group_counts.values()) or 1
    bot_groups = [
        {'name': k, 'count': v, 'pct': round(v * 100 / group_total)}
        for k, v in sorted(group_counts.items(), key=lambda x: -x[1])
    ]

    context = {
        **dr,
        'counts': {
            'crawler':  crawler_qs.count(),
            'archive':  archive_qs.count(),
            'email':    email_qs.count(),
            'people':   people_qs.count(),
            'projects': project_qs.count(),
        },
        'top_bots':   top_bots,
        'trap_counts': trap_counts,
        'bot_groups': bot_groups,
        'daily':      daily,
        'recent_crawlers': crawler_qs.order_by('-timestamp')[:10],
        'recent_emails':   email_qs.order_by('-received_at')[:5],
    }
    return render(request, 'dashboard/overview.html', context)


@staff_member_required(login_url='/django-admin/login/')
def crawlers(request):
    dr = _parse_date_range(request)
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

    top_ips = list(
        qs.values('ip_address')
        .annotate(count=Count('id'))
        .order_by('-count')[:20]
    )

    top_paths = list(
        qs.values('path')
        .annotate(count=Count('id'))
        .order_by('-count')[:20]
    )

    daily = _daily_chart(CrawlerVisit.objects.all(), days=60)

    context = {
        **dr,
        'total': total,
        'top_bots': top_bots,
        'trap_counts': trap_counts,
        'top_ips': top_ips,
        'top_paths': top_paths,
        'daily': daily,
        'recent': qs.order_by('-timestamp').select_related()[:50],
    }
    return render(request, 'dashboard/crawlers.html', context)


@staff_member_required(login_url='/django-admin/login/')
def archive(request):
    dr = _parse_date_range(request)
    qs = _apply_range(ArchiveVisit.objects.all(), dr)

    top_bots, total = _bot_breakdown(qs, limit=20)

    depth_counts = list(
        qs.values('depth')
        .annotate(count=Count('id'))
        .order_by('depth')[:15]
    )
    depth_total = sum(d['count'] for d in depth_counts) or 1
    for d in depth_counts:
        d['pct'] = round(d['count'] * 100 / depth_total)

    top_ips = list(
        qs.values('ip_address')
        .annotate(count=Count('id'), max_depth=Max('depth'))
        .order_by('-count')[:20]
    )

    top_roots = list(
        qs.values('slug')
        .annotate(count=Count('id'))
        .order_by('-count')[:20]
    )

    daily = _daily_chart(ArchiveVisit.objects.all(), days=30)

    context = {
        **dr,
        'total': total,
        'top_bots': top_bots,
        'depth_counts': depth_counts,
        'top_ips': top_ips,
        'top_roots': top_roots,
        'daily': daily,
        'recent': qs.order_by('-timestamp')[:50],
    }
    return render(request, 'dashboard/archive.html', context)


@staff_member_required(login_url='/django-admin/login/')
def emails(request):
    dr = _parse_date_range(request)
    qs = _apply_range(InboundEmail.objects.all(), dr, field='received_at')

    total = qs.count()

    # Top sender domains
    domain_counts = Counter()
    for sender in qs.values_list('sender', flat=True):
        domain = sender.split('@')[-1].lower() if '@' in sender else sender
        domain_counts[domain] += 1
    top_domains = [
        {'domain': d, 'count': c, 'pct': round(c * 100 / total) if total else 0}
        for d, c in domain_counts.most_common(20)
    ]

    # Top recipients (which fake addresses are getting hit)
    top_recipients = list(
        qs.values('recipient')
        .annotate(count=Count('id'))
        .order_by('-count')[:20]
    )

    daily = _daily_chart(InboundEmail.objects.all(), days=30, field='received_at')

    context = {
        **dr,
        'total': total,
        'top_domains': top_domains,
        'top_recipients': top_recipients,
        'daily': daily,
        'recent': qs.order_by('-received_at').prefetch_related('matches')[:50],
    }
    return render(request, 'dashboard/emails.html', context)


@staff_member_required(login_url='/django-admin/login/')
def people(request):
    dr = _parse_date_range(request)
    people_qs   = _apply_range(PeoplePageVisit.objects.all(), dr)
    project_qs  = _apply_range(ProjectPageVisit.objects.all(), dr)

    people_bots, people_total = _bot_breakdown(people_qs, limit=20)
    project_bots, project_total = _bot_breakdown(project_qs, limit=20)

    top_people_ips = list(
        people_qs.values('ip_address')
        .annotate(count=Count('id'))
        .order_by('-count')[:15]
    )
    top_project_ips = list(
        project_qs.values('ip_address')
        .annotate(count=Count('id'))
        .order_by('-count')[:15]
    )

    people_daily = _daily_chart(PeoplePageVisit.objects.all(), days=30)
    project_daily = _daily_chart(ProjectPageVisit.objects.all(), days=30)

    context = {
        **dr,
        'people_total': people_total,
        'project_total': project_total,
        'people_bots': people_bots,
        'project_bots': project_bots,
        'top_people_ips': top_people_ips,
        'top_project_ips': top_project_ips,
        'people_daily': people_daily,
        'project_daily': project_daily,
        'recent_people':  people_qs.order_by('-timestamp')[:30],
        'recent_projects': project_qs.order_by('-timestamp')[:30],
    }
    return render(request, 'dashboard/people.html', context)
