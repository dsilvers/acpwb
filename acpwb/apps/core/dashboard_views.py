"""
ACPWB Activity Dashboard — requires staff login.

Stats are pre-computed by the `precalc_dashboard` management command and
stored as DashboardStat rows in the database.  Views read directly from those
rows; no caching layer is needed.  Run precalc_dashboard on a 30-minute cron.

Recent records (last 50 visits, last 5 emails, etc.) are always live queries.
"""
from datetime import date, timedelta

from django.contrib.admin.views.decorators import staff_member_required
from django.db.models import Max
from django.http import Http404, HttpResponse
from django.shortcuts import get_object_or_404, render

from apps.core.models import DashboardStat
from apps.honeypot.models import ArchiveVisit, CanaryToken, CrawlerVisit, InternalLoginAttempt
from apps.people.models import PeoplePageVisit
from apps.projects.models import ProjectPageVisit
from apps.public.models import DataOptOutRequest, JobApplication, JobApplicationDocument
from apps.webhooks.models import CallLog, InboundEmail, VoicemailRecording


# ── Stat helpers ──────────────────────────────────────────────────────────────

def _stat(key, default=None):
    try:
        return DashboardStat.objects.get(key=key).value
    except DashboardStat.DoesNotExist:
        return default


def _stats(*keys):
    rows = DashboardStat.objects.filter(key__in=keys).values('key', 'value')
    return {r['key']: r['value'] for r in rows}


def _updated_at(key='crawlers.total'):
    return DashboardStat.objects.filter(key=key).values_list('updated_at', flat=True).first()


def _top_named(d, limit=20):
    """Convert {name: count} dict → [{name, count, pct}] sorted by count desc."""
    if not d:
        return []
    items = sorted(d.items(), key=lambda x: x[1], reverse=True)[:limit]
    total = sum(v for _, v in items) or 1
    return [{'name': k, 'count': v, 'pct': round(v * 100 / total)} for k, v in items]


def _top_field(d, field, limit=20):
    """Convert {k: count} dict → [{field: k, count: c}] sorted by count desc."""
    if not d:
        return []
    return [
        {field: k, 'count': v}
        for k, v in sorted(d.items(), key=lambda x: x[1], reverse=True)[:limit]
    ]


def _daily_bars(d, days=60):
    """Convert {date_str: count} dict → {bars, start, end} chart structure."""
    if d is None:
        d = {}
    today = date.today()
    bars = []
    for i in range(days - 1, -1, -1):
        day = today - timedelta(days=i)
        ds = day.isoformat()
        bars.append({'date': ds, 'count': d.get(ds, 0)})
    peak = max((r['count'] for r in bars), default=1) or 1
    for r in bars:
        r['pct'] = round(r['count'] * 100 / peak)
    return {
        'bars': bars,
        'start': bars[0]['date'] if bars else '',
        'end': bars[-1]['date'] if bars else '',
    }


def _trap_counts(by_trap_type_dict):
    """Convert {trap_type: count} dict → [{label, trap_type, count, pct}] sorted."""
    from apps.honeypot.models import CrawlerVisit as CV
    label_map = dict(CV.TRAP_CHOICES)
    if not by_trap_type_dict:
        return []
    total = sum(by_trap_type_dict.values()) or 1
    items = sorted(by_trap_type_dict.items(), key=lambda x: x[1], reverse=True)
    return [
        {
            'trap_type': k,
            'label': label_map.get(k, k),
            'count': v,
            'pct': round(v * 100 / total),
        }
        for k, v in items
    ]


def _depth_counts(by_depth_dict):
    """Convert {depth_str: count} dict → [{depth, count, pct}] sorted by depth."""
    if not by_depth_dict:
        return []
    total = sum(by_depth_dict.values()) or 1
    items = sorted(
        ((k, v) for k, v in by_depth_dict.items() if k != ''),
        key=lambda x: int(x[0]),
    )
    return [
        {'depth': int(k), 'count': v, 'pct': round(v * 100 / total)}
        for k, v in items
    ]


def _top_domains(by_domain_dict, limit=20):
    """Convert {domain: count} dict → [{domain, count, pct}] sorted."""
    if not by_domain_dict:
        return []
    items = sorted(by_domain_dict.items(), key=lambda x: x[1], reverse=True)[:limit]
    total = sum(v for _, v in items) or 1
    return [{'domain': k, 'count': v, 'pct': round(v * 100 / total)} for k, v in items]


# ── Bot breakdown helpers (still used for archive top_bots display) ───────────

def _bot_breakdown_from_dict(d, limit=20):
    """Convert {bot_type: count} dict → [{name, count, pct}]."""
    return _top_named(d, limit)


# ── Views ─────────────────────────────────────────────────────────────────────

@staff_member_required(login_url='/django-admin/login/')
def overview(request):
    s = _stats(
        'crawlers.total', 'archive.total', 'emails.total', 'people.total',
        'projects.total', 'login_attempts.total', 'optouts.total',
        'crawlers.by_bot_type', 'crawlers.by_trap_type', 'crawlers.by_bot_group',
        'crawlers.daily',
    )
    ctx = {
        'counts': {
            'crawler':  s.get('crawlers.total', 0),
            'archive':  s.get('archive.total', 0),
            'email':    s.get('emails.total', 0),
            'people':   s.get('people.total', 0),
            'projects': s.get('projects.total', 0),
            'logins':   s.get('login_attempts.total', 0),
            'optouts':  s.get('optouts.total', 0),
            'job_applications': JobApplication.objects.count(),
        },
        'top_bots':        _top_named(s.get('crawlers.by_bot_type', {}), 15),
        'trap_counts':     _trap_counts(s.get('crawlers.by_trap_type', {})),
        'bot_groups':      _top_named(s.get('crawlers.by_bot_group', {})),
        'daily':           _daily_bars(s.get('crawlers.daily', {}), 30),
        'recent_crawlers': list(CrawlerVisit.objects.order_by('-timestamp')[:10]),
        'recent_emails':   list(InboundEmail.objects.order_by('-received_at')[:5]),
        'recent_optouts':  list(DataOptOutRequest.objects.order_by('-created_at')[:10]),
        'updated_at':      _updated_at('crawlers.total'),
    }
    return render(request, 'dashboard/overview.html', ctx)


@staff_member_required(login_url='/django-admin/login/')
def crawlers(request):
    import time
    s = _stats(
        'crawlers.total', 'crawlers.by_bot_type', 'crawlers.by_trap_type',
        'crawlers.by_ip', 'crawlers.by_path', 'crawlers.by_host', 'crawlers.daily',
        'crawlers.probe_by_path', 'crawlers.webshell_cmds', 'canary.triggered_count',
    )
    canary_latest = CanaryToken.objects.filter(triggered=True).order_by('-triggered_at').first()
    ctx = {
        'total':                s.get('crawlers.total', 0),
        'top_bots':             _top_named(s.get('crawlers.by_bot_type', {}), 30),
        'trap_counts':          _trap_counts(s.get('crawlers.by_trap_type', {})),
        'top_ips':              _top_field(s.get('crawlers.by_ip', {}), 'ip_address', 20),
        'top_paths':            _top_field(s.get('crawlers.by_path', {}), 'path', 20),
        'top_hosts':            _top_field(s.get('crawlers.by_host', {}), 'host', 20),
        'daily':                _daily_bars(s.get('crawlers.daily', {}), 60),
        'recent':               list(CrawlerVisit.objects.order_by('-timestamp')[:50]),
        'top_probe_paths':      _top_field(s.get('crawlers.probe_by_path', {}), 'path', 20),
        'webshell_commands':    _top_field(s.get('crawlers.webshell_cmds', {}), 'query_string', 10),
        'canary_trigger_count': s.get('canary.triggered_count', 0),
        'canary_latest':        canary_latest,
        'updated_at':           _updated_at('crawlers.total'),
        'graphs_ts':            int(time.time() // 60),
    }
    return render(request, 'dashboard/crawlers.html', ctx)


@staff_member_required(login_url='/django-admin/login/')
def archive(request):
    # Top IPs need max_depth — read top 20 from stored dict, then query max_depth
    by_ip = _stat('archive.by_ip', {})
    top_ip_keys = [ip for ip, _ in sorted(by_ip.items(), key=lambda x: x[1], reverse=True)[:20]]
    ip_count_map = {ip: by_ip[ip] for ip in top_ip_keys}
    max_depths = (
        ArchiveVisit.objects
        .filter(ip_address__in=top_ip_keys)
        .values('ip_address')
        .annotate(max_depth=Max('depth'))
    )
    depth_map = {r['ip_address']: r['max_depth'] for r in max_depths}
    top_ips = [
        {'ip_address': ip, 'count': ip_count_map[ip], 'max_depth': depth_map.get(ip, 0)}
        for ip in top_ip_keys
    ]

    ctx = {
        'total':        _stat('archive.total', 0),
        'top_bots':     _bot_breakdown_from_dict(_stat('archive.by_bot_type', {}), 20),
        'depth_counts': _depth_counts(_stat('archive.by_depth', {})),
        'top_ips':      top_ips,
        'top_roots':    _top_field(_stat('archive.by_slug', {}), 'slug', 20),
        'daily':        _daily_bars(_stat('archive.daily', {}), 30),
        'recent':       list(ArchiveVisit.objects.order_by('-timestamp')[:50]),
        'updated_at':   _updated_at('archive.total'),
    }
    return render(request, 'dashboard/archive.html', ctx)


@staff_member_required(login_url='/django-admin/login/')
def emails(request):
    ctx = {
        'total':          _stat('emails.total', 0),
        'top_domains':    _top_domains(_stat('emails.by_domain', {}), 20),
        'top_recipients': _top_field(_stat('emails.by_recipient', {}), 'recipient', 20),
        'daily':          _daily_bars(_stat('emails.daily', {}), 30),
        'recent':         list(InboundEmail.objects.order_by('-received_at').prefetch_related('matches')[:50]),
        'updated_at':     _updated_at('emails.total'),
    }
    return render(request, 'dashboard/emails.html', ctx)


@staff_member_required(login_url='/django-admin/login/')
def email_detail(request, pk):
    from django.shortcuts import get_object_or_404
    from django.utils.html import strip_tags
    email = get_object_or_404(InboundEmail, pk=pk)
    body = email.body_plain or strip_tags(email.body_html)
    return render(request, 'dashboard/email_detail.html', {
        'email': email,
        'body': body,
    })


@staff_member_required(login_url='/django-admin/login/')
def honeypot_traps(request):
    by_trap = _stat('crawlers.by_trap_type', {})
    trap_totals = _trap_counts(by_trap)

    CONTENT_KEYS = {'wiki', 'report_list', 'report_download', 'dataset', 'api', 'well_known', 'pow'}
    SCANNER_KEYS = {'scanner_probe', 'env_probe', 'wp_probe', 'webshell_probe'}

    content_total = sum(by_trap.get(k, 0) for k in CONTENT_KEYS)
    ghost_total = by_trap.get('ghost_link', 0)
    scanner_total = sum(by_trap.get(k, 0) for k in SCANNER_KEYS)
    policy_total = by_trap.get('policy', 0)

    canary_latest = CanaryToken.objects.filter(triggered=True).order_by('-triggered_at').first()

    ctx = {
        'trap_totals':          trap_totals,
        'content_total':        content_total,
        'ghost_total':          ghost_total,
        'scanner_total':        scanner_total,
        'policy_total':         policy_total,
        'canary_trigger_count': _stat('canary.triggered_count', 0),
        'canary_latest':        canary_latest,
        'top_probe_paths':      _top_field(_stat('crawlers.probe_by_path', {}), 'path', 20),
        'webshell_commands':    _top_field(_stat('crawlers.webshell_cmds', {}), 'query_string', 10),
        'updated_at':           _updated_at('crawlers.total'),
    }
    return render(request, 'dashboard/honeypot.html', ctx)


@staff_member_required(login_url='/django-admin/login/')
def people(request):
    ctx = {
        'people_total':    _stat('people.total', 0),
        'project_total':   _stat('projects.total', 0),
        'people_bots':     _top_named(_stat('people.by_bot_type', {}), 20),
        'project_bots':    _top_named(_stat('projects.by_bot_type', {}), 20),
        'top_people_ips':  _top_field(_stat('people.by_ip', {}), 'ip_address', 15),
        'top_project_ips': _top_field(_stat('projects.by_ip', {}), 'ip_address', 15),
        'people_daily':    _daily_bars(_stat('people.daily', {}), 30),
        'project_daily':   _daily_bars(_stat('projects.daily', {}), 30),
        'recent_people':   list(PeoplePageVisit.objects.order_by('-timestamp')[:30]),
        'recent_projects': list(ProjectPageVisit.objects.order_by('-timestamp')[:30]),
        'updated_at':      _updated_at('people.total'),
    }
    return render(request, 'dashboard/people.html', ctx)


@staff_member_required(login_url='/django-admin/login/')
def logins(request):
    ctx = {
        'total':        _stat('login_attempts.total', 0),
        'top_usernames': _top_field(_stat('login_attempts.by_username', {}), 'username', 30),
        'top_ips':       _top_field(_stat('login_attempts.by_ip', {}), 'ip_address', 20),
        'by_source':     _top_named(_stat('login_attempts.by_source', {})),
        'daily':         _daily_bars(_stat('login_attempts.daily', {}), 30),
        'recent':        list(InternalLoginAttempt.objects.order_by('-created_at')[:50]),
        'updated_at':    _updated_at('login_attempts.total'),
    }
    return render(request, 'dashboard/logins.html', ctx)


@staff_member_required(login_url='/django-admin/login/')
def live_stream(request):
    from django.conf import settings
    return render(request, 'dashboard/live_stream.html', {
        'ws_token': getattr(settings, 'STREAM_WS_TOKEN', ''),
    })


@staff_member_required(login_url='/django-admin/login/')
def voicemails(request):
    # Annotate call log with whether a voicemail was left
    vm_sids = set(VoicemailRecording.objects.values_list('call_sid', flat=True))
    calls = list(CallLog.objects.order_by('-received_at')[:100])
    for call in calls:
        call.has_voicemail = call.call_sid in vm_sids

    ctx = {
        'voicemails':   list(VoicemailRecording.objects.order_by('-received_at')[:100]),
        'vm_total':     VoicemailRecording.objects.count(),
        'vm_pending':   VoicemailRecording.objects.filter(transcription_status='pending').count(),
        'vm_completed': VoicemailRecording.objects.filter(transcription_status='completed').count(),
        'vm_failed':    VoicemailRecording.objects.filter(transcription_status='failed').count(),
        'calls':        calls,
        'call_total':   CallLog.objects.count(),
    }
    return render(request, 'dashboard/voicemails.html', ctx)


@staff_member_required(login_url='/django-admin/login/')
def voicemail_audio(request, recording_sid):
    import base64
    from urllib.error import URLError
    from urllib.request import Request, urlopen

    from django.conf import settings
    from django.http import HttpResponse, HttpResponseNotFound, StreamingHttpResponse

    try:
        vm = VoicemailRecording.objects.get(recording_sid=recording_sid)
    except VoicemailRecording.DoesNotExist:
        return HttpResponseNotFound()

    audio_url = vm.recording_url + '.mp3'
    credentials = base64.b64encode(
        f"{settings.TWILIO_ACCOUNT_SID}:{settings.TWILIO_AUTH_TOKEN}".encode()
    ).decode()
    req = Request(audio_url, headers={'Authorization': f'Basic {credentials}'})

    try:
        response = urlopen(req, timeout=10)
    except URLError:
        return HttpResponse(status=502)

    return StreamingHttpResponse(
        response,
        content_type='audio/mpeg',
        headers={'Content-Disposition': f'inline; filename="{recording_sid}.mp3"'},
    )


@staff_member_required(login_url='/django-admin/login/')
def careers_applications(request):
    from django.db.models import Count

    qs = JobApplication.objects.all()
    total = qs.count()
    with_resume = qs.exclude(resume_filename='').count()
    jobs_count = qs.values('job_title').distinct().count()

    by_job = (
        qs.values('job_title')
        .annotate(count=Count('id'))
        .order_by('-count')[:20]
    )
    job_rows = []
    for row in by_job:
        pct = round(row['count'] / total * 100) if total else 0
        job_rows.append({'title': row['job_title'], 'count': row['count'], 'pct': pct})

    recent = qs.order_by('-created_at')[:50]

    return render(request, 'dashboard/careers.html', {
        'total': total,
        'with_resume': with_resume,
        'jobs_count': jobs_count,
        'job_rows': job_rows,
        'recent': recent,
        'updated_at': None,
    })


@staff_member_required(login_url='/django-admin/login/')
def career_application_detail(request, pk):
    from apps.public.career_generator import generate_job, is_valid_job_id
    app = get_object_or_404(JobApplication, pk=pk)
    docs = app.documents.all()
    job = generate_job(app.job_id) if is_valid_job_id(app.job_id) else None
    return render(request, 'dashboard/career_application_detail.html', {'app': app, 'docs': docs, 'job': job})


@staff_member_required(login_url='/django-admin/login/')
def career_download_resume(request, pk):
    app = get_object_or_404(JobApplication, pk=pk)
    if not app.resume_data:
        raise Http404
    content_type = app.resume_content_type or 'application/octet-stream'
    response = HttpResponse(bytes(app.resume_data), content_type=content_type)
    response['Content-Disposition'] = f'attachment; filename="{app.resume_filename}"'
    return response


@staff_member_required(login_url='/django-admin/login/')
def career_download_document(request, doc_pk):
    doc = get_object_or_404(JobApplicationDocument, pk=doc_pk)
    content_type = doc.content_type or 'application/octet-stream'
    response = HttpResponse(bytes(doc.data), content_type=content_type)
    response['Content-Disposition'] = f'attachment; filename="{doc.filename}"'
    return response
