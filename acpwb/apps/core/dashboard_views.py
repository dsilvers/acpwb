"""
ACPWB Activity Dashboard — requires staff login.

Stats are pre-computed by the `precalc_dashboard` management command and
stored as DashboardStat rows in the database.  Views read directly from those
rows; no caching layer is needed.  Run precalc_dashboard on a 30-minute cron.

Recent records (last 50 visits, last 5 emails, etc.) are always live queries.
"""
from collections import Counter
from datetime import date, timedelta

from django.conf import settings
from django.contrib import messages
from django.contrib.admin.views.decorators import staff_member_required
from django.core.exceptions import ValidationError
from django.core.paginator import Paginator
from django.db.models import Avg, Count, Max, Q
from django.db.models.functions import TruncDate
from django.http import Http404, HttpResponse
from django.shortcuts import get_object_or_404, redirect, render
from django.utils import timezone
from django.utils.http import content_disposition_header

from apps.core.models import DashboardStat
from apps.honeypot.models import (
    ArchiveVisit, CanaryToken, ComplianceCompletion, CrawlerVisit, EdgeConnection, EdgeRequest, InternalLoginAttempt,
)
from apps.people.models import PeoplePageVisit
from apps.projects.models import ProjectPageVisit
from apps.public.models import ConferenceRegistration, DataOptOutRequest, JobApplication, JobApplicationDocument
from apps.webhooks import replies
from apps.webhooks.models import CallLog, HoneypotMatch, InboundEmail, VoicemailRecording


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
        'daily_table':          list(reversed(_daily_bars(s.get('crawlers.daily', {}), 60)['bars'])),
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
    # Top IPs need max_depth — both read from precalculated stats (no live query)
    by_ip = _stat('archive.by_ip', {})
    depth_map = _stat('archive.max_depth_by_ip', {})
    top_ip_keys = [ip for ip, _ in sorted(by_ip.items(), key=lambda x: x[1], reverse=True)[:20]]
    top_ips = [
        {'ip_address': ip, 'count': by_ip[ip], 'max_depth': depth_map.get(ip, 0)}
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
        'updated_at':     _updated_at('emails.total'),
    }
    return render(request, 'dashboard/emails.html', ctx)


@staff_member_required(login_url='/django-admin/login/')
def email_inbox(request):
    q = request.GET.get('q', '').strip()
    emails = InboundEmail.objects.order_by('-received_at')
    if q:
        # One box: an address or domain matches sender/recipient, words match
        # subject or body (the table is small enough for plain icontains).
        emails = emails.filter(
            Q(sender__icontains=q) | Q(from_header__icontains=q) | Q(recipient__icontains=q) | Q(subject__icontains=q)
            | Q(body_plain__icontains=q) | Q(body_html__icontains=q)
        )
    # _match_honeypot also logs a confidence='none' row when nothing matched; that isn't a match.
    emails = emails.annotate(
        match_count=Count('matches', filter=~Q(matches__match_confidence='none'), distinct=True),
        reply_count=Count('replies', distinct=True),
        last_reply_at=Max('replies__sent_at'),
    )
    page = Paginator(emails.defer('body_plain', 'body_html', 'raw_payload'), 50).get_page(request.GET.get('page'))
    return render(request, 'dashboard/email_inbox.html', {'page': page, 'q': q})


@staff_member_required(login_url='/django-admin/login/')
def email_detail(request, pk):
    email = get_object_or_404(InboundEmail, pk=pk)
    reply = replies.reply_defaults(email)
    reply_error = None
    if request.method == 'POST':
        reply = {k: request.POST.get(k, '') for k in reply}
        try:
            replies.send_reply(email, sent_by=request.user.get_username(), **reply)
        except ValidationError as e:
            reply_error = ' '.join(e.messages)
        except Exception as e:  # SMTP / network: show it, keep the draft
            reply_error = f'Mailgun did not accept the message: {e}'
        else:
            messages.success(request, 'Reply sent.')
            return redirect('dashboard-email-detail', pk=pk)
    return render(request, 'dashboard/email_detail.html', {
        'email': email,
        'body': replies.body_text(email),
        'reply_to_header': replies.header(email, 'Reply-To'),
        'attachments': replies.attachments(email),
        'honeypot_matches': email.matches.exclude(match_confidence='none').select_related(
            'generated_employee', 'original_visit', 'source_code_email',
        ),
        'sent_replies': list(email.replies.order_by('sent_at')),  # conversation order
        'reply': reply,
        'reply_error': reply_error,
        'reply_domain': settings.REPLY_FROM_DOMAIN,
        'inline_types': INLINE_ATTACHMENT_TYPES,
    })


# Shown in the browser when asked; everything else (HTML, SVG, Office files…)
# always downloads, so nothing a sender wrote runs on the dashboard origin.
INLINE_ATTACHMENT_TYPES = {'image/png', 'image/jpeg', 'image/gif', 'image/webp', 'application/pdf', 'text/plain'}


@staff_member_required(login_url='/django-admin/login/')
def email_attachment(request, pk, index):
    email = get_object_or_404(InboundEmail, pk=pk)
    found = replies.attachment(email, index)
    if found is None:
        raise Http404
    filename, content_type, data = found
    inline = request.GET.get('inline') == '1' and content_type in INLINE_ATTACHMENT_TYPES
    response = HttpResponse(data, content_type=content_type if inline else 'application/octet-stream')
    response['Content-Disposition'] = content_disposition_header(not inline, filename)
    response['X-Content-Type-Options'] = 'nosniff'
    if content_type != 'application/pdf':  # Chrome's PDF viewer won't load sandboxed
        response['Content-Security-Policy'] = 'sandbox'
    return response


@staff_member_required(login_url='/django-admin/login/')
def honeypot_traps(request):
    by_trap = _stat('crawlers.by_trap_type', {})
    trap_totals = _trap_counts(by_trap)

    CONTENT_KEYS = {'wiki', 'report_list', 'report_download', 'dataset', 'api', 'well_known', 'pow', 'handbook', 'process_improvement', 'source_code', 'source_code_download', 'compliance', 'compliance_certificate'}
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
def edge(request):
    import time
    summary = _stat('edge.summary_24h', {}) or {}
    process = _stat('edge.process', {}) or {}
    reqs = _stat('edge.requests_24h', {}) or {}
    total = summary.get('total', 0)
    req_total = reqs.get('total', 0)
    outcomes = summary.get('by_outcome', {})
    alpn = summary.get('by_alpn', {})
    failed = sum(v for k, v in outcomes.items() if k not in ('ok', 'plain'))
    ctx = {
        'total':          total,
        'failed':         failed,
        'failed_pct':     round(failed * 100 / total, 1) if total else 0,
        'h2':             alpn.get('h2', 0),
        'h2_pct':         round(alpn.get('h2', 0) * 100 / total, 1) if total else 0,
        'by_outcome':     _top_named(outcomes),
        'by_alpn':        _top_named(alpn),
        'by_tls_version': _top_named(summary.get('by_tls_version', {})),
        'top_ja4':        summary.get('top_ja4', []),
        'top_failed_ja4': summary.get('top_failed_ja4', []),
        'top_h2':         summary.get('top_h2_fingerprint', []),
        'by_close':       _top_named(summary.get('by_close_reason', {})),
        'by_goaway':      _top_named(summary.get('by_goaway', {})),
        'top_h2_abuse':   summary.get('top_h2_abuse', []),
        'top_conn_ips':   summary.get('top_conn_ips', []),
        'aggregated':     summary.get('aggregated_conns', 0),
        'rejected':       summary.get('rejected_requests', 0),
        'process':        process,
        'uptime_hours':   round((time.time() - process['started_at']) / 3600, 1) if process.get('started_at') else None,
        'recent':         list(EdgeConnection.objects.order_by('-timestamp')[:50]),
        'req_total':      req_total,
        'req_clients':    reqs.get('clients', 0),
        'req_new_clients': reqs.get('new_clients', 0),
        'req_ips':        reqs.get('ips', 0),
        'req_failed':     reqs.get('failed', 0),
        'req_failed_pct': round(reqs.get('failed', 0) * 100 / req_total, 1) if req_total else 0,
        'req_by_route':   reqs.get('by_route', []),
        'req_by_status':  _top_named(reqs.get('by_status', {})),
        'req_by_error':   _top_named(reqs.get('by_error', {})),
        'req_by_protocol': _top_named(reqs.get('by_protocol', {})),
        'recent_requests': list(EdgeRequest.objects.order_by('-timestamp')[:50]),
        'updated_at':     _updated_at('edge.summary_24h'),
    }
    classes = _stat('edge.classes_24h', {}) or {}
    by_network = classes.get('by_network', [])
    class_total = sum(r['requests'] for r in by_network)
    # Proven: the request itself shows automation (a contradicted browser
    # claim, a self-declared bot or HTTP tool, a browser in a datacenter).
    # Unproven: a real browser engine on a home network, or no claim to
    # check; on this site that's still almost all automation, but nothing
    # in one request proves it.
    proven = sum(r.get(k, 0) for r in by_network for k in ('impersonator', 'known_bot', 'tool', 'headless'))
    unproven = sum(r.get('browser', 0) + r.get('unknown', 0) for r in by_network)
    # The headline: the last 3 rolled-up hours, proof by behaviour included
    # (precalc's 'headline'); the class-only 24h figures if it's missing.
    head = classes.get('headline', {})
    h3 = head.get('3h') or {'total': class_total, 'by_class': proven, 'by_behaviour': 0, 'unproven': unproven,
                            'bare_ips': 0}
    h24 = head.get('24h') or h3

    def pct(n, d):
        return round(n * 100 / d, 1) if d else 0
    fake = sum(r['impersonator'] for r in by_network)
    res_fake = sum(r['impersonator'] for r in by_network if r['network'] == 'residential')
    ctx.update({
        'cls_by_network':  by_network,
        'cls_total':       h3['total'],
        'cls_proven':      h3['by_class'] + h3['by_behaviour'],
        'cls_proven_pct':  pct(h3['by_class'] + h3['by_behaviour'], h3['total']),
        'cls_by_class_pct': pct(h3['by_class'], h3['total']),
        'cls_by_behaviour_pct': pct(h3['by_behaviour'], h3['total']),
        'cls_proven_24h_pct': pct(h24['by_class'] + h24['by_behaviour'], h24['total']),
        'cls_unproven':    h3['unproven'],
        'cls_unproven_pct': pct(h3['unproven'], h3['total']),
        'cls_bare_ips':    h3.get('bare_ips', 0),
        'cls_no_hello_ips': sum(r.get('no_hello_ips', 0) for r in by_network if r['network'] == 'residential'),
        'multi_ip':        classes.get('multi_ip', {}),
        'cls_fake':        fake,
        'cls_fake_pct':    round(fake * 100 / class_total, 1) if class_total else 0,
        'cls_res_fake':    res_fake,
        'cls_res_fake_ips': sum(r['fake_ips'] for r in by_network if r['network'] == 'residential'),
        'cls_hourly':      classes.get('hourly', []),
        'cls_fake_fps':    classes.get('top_fake_fps', []),
        'cls_proxy_asns':  classes.get('top_proxy_asns', []),
        'cls_fake_ips':    classes.get('top_fake_ips', []),
        'cls_exposed':     classes.get('exposed', []),
        'cls_cookie_pools': classes.get('cookie_pools', []),
        'cls_trap_fps':    classes.get('trap_fps', []),
    })
    return render(request, 'dashboard/edge.html', ctx)


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
    from django.db.models import Count, Q

    qs = JobApplication.objects.all()
    agg = qs.aggregate(
        total=Count('id'),
        with_resume=Count('id', filter=~Q(resume_filename='')),
        jobs_count=Count('job_title', distinct=True),
    )
    total = agg['total']
    with_resume = agg['with_resume']
    jobs_count = agg['jobs_count']

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


@staff_member_required(login_url='/django-admin/login/')
def conference_registration_detail(request, pk):
    from apps.public.models import ConferenceRegistration
    reg = get_object_or_404(ConferenceRegistration, pk=pk)
    return render(request, 'dashboard/conference_detail.html', {
        'reg': reg,
        'confirmation_number': f'PERCH-{reg.year}-{reg.pk:05d}',
    })


def conference_registrations(request):
    from django.db.models import Count

    qs = ConferenceRegistration.objects.all()
    total = qs.count()

    by_type = (
        qs.values('registration_type')
        .annotate(count=Count('id'))
        .order_by('-count')
    )
    type_rows = []
    labels = {'full': 'Full Conference', 'day1': 'Day 1 Only', 'day2': 'Day 2 Only'}
    for row in by_type:
        pct = round(row['count'] / total * 100) if total else 0
        type_rows.append({
            'label': labels.get(row['registration_type'], row['registration_type']),
            'count': row['count'],
            'pct': pct,
        })

    recent = qs.order_by('-created_at')[:50]

    return render(request, 'dashboard/conference.html', {
        'total': total,
        'type_rows': type_rows,
        'recent': recent,
    })


# ── /compliance/ (ComplianceCenter) ──────────────────────────────────────

# Expected score from guessing: most questions have four choices, some two
# (True/False) or five ("All of the above").
COMPLIANCE_RANDOM_SCORE = 26

_CLIENT_CLASS_LABELS = {
    'known_bot': 'Known bot', 'tool': 'Tool / script', 'browser': 'Browser',
    'headless': 'Headless / automated', 'impersonator': 'Impersonator', 'unknown': 'Unknown', '': '—',
    'unclassified': 'Unclassified',
}

_FUNNEL_KINDS = [
    ('catalog', 'Catalog'), ('agencies', 'Agency directory'), ('sponsor', 'Agency page'), ('topics', 'Topic index'),
    ('topic', 'Topic page'), ('sitemap', 'Sitemaps'), ('course', 'Course page'), ('question', 'Question pages'),
    ('results', 'Results pages'), ('certificate', 'Certificate pages'), ('certificate_pdf', 'Certificate PDFs'),
    ('verify', 'Verification pages'), ('transcript', 'Transcript'),
]


def _bucket_scores(rows):
    """[(score, n)] → 10 buckets of 10 points (100 lands in the last)."""
    buckets = [0] * 10
    for score, n in rows:
        buckets[min(score // 10, 9)] += n
    peak = max(buckets) or 1
    return [{'label': f'{i * 10}–{i * 10 + 9 if i < 9 else 100}%', 'count': c, 'pct': round(c * 100 / peak),
             'random': i * 10 <= COMPLIANCE_RANDOM_SCORE < i * 10 + 10} for i, c in enumerate(buckets)]


def _top_list(counter, limit=15):
    items = counter.most_common(limit)
    peak = items[0][1] if items else 1
    return [{'name': k, 'count': v, 'pct': round(v * 100 / peak)} for k, v in items]


@staff_member_required(login_url='/django-admin/login/')
def compliance(request):
    from apps.core import compliance_stats

    try:
        days = int(request.GET.get('days', 7))
    except ValueError:
        days = 7
    days = days if days in (1, 7, 30) else 7
    since = timezone.now() - timedelta(days=days)

    comps = ComplianceCompletion.objects.filter(timestamp__gte=since)
    completed = comps.filter(event='completed')
    summary = completed.aggregate(n=Count('id'), avg=Avg('score'), passed=Count('id', filter=Q(passed=True)))
    event_counts = dict(comps.values_list('event').annotate(n=Count('id')))

    # Score vs. random guessing, overall and per agent.
    score_hist = _bucket_scores(completed.values_list('score').annotate(n=Count('id')))
    agents = list(
        completed.values('bot_type', 'client_class')
        .annotate(attempts=Count('id'), avg=Avg('score'), passed=Count('id', filter=Q(passed=True)),
                  ips=Count('ip_address', distinct=True))
        .order_by('-attempts')[:25]
    )
    certs_by_agent = dict(
        comps.filter(event='certificate').values_list('bot_type').annotate(n=Count('id'))
    )
    for a in agents:
        a['avg'] = round(a['avg'] or 0)
        a['pass_rate'] = round(a['passed'] * 100 / a['attempts']) if a['attempts'] else 0
        a['certificates'] = certs_by_agent.get(a['bot_type'], 0)
        a['class_label'] = _CLIENT_CLASS_LABELS.get(a['client_class'], a['client_class'])
        a['reads'] = a['attempts'] >= 3 and a['avg'] >= 50

    leaderboard = list(
        comps.filter(event='certificate').values('ip_address', 'bot_type', 'client_class')
        .annotate(n=Count('id'), courses=Count('course', distinct=True), last=Max('timestamp'))
        .order_by('-n')[:15]
    )
    for row in leaderboard:
        row['class_label'] = _CLIENT_CLASS_LABELS.get(row['client_class'], row['client_class'])

    daily_rows = (
        ComplianceCompletion.objects.filter(timestamp__gte=timezone.now() - timedelta(days=30), event__in=['certificate', 'completed'])
        .annotate(day=TruncDate('timestamp')).values('day', 'event').annotate(n=Count('id'))
    )
    certs_daily = {r['day'].isoformat(): r['n'] for r in daily_rows if r['event'] == 'certificate'}

    identified = list(comps.filter(event='identified').order_by('-timestamp')[:100])
    for row in identified:
        row.class_label = _CLIENT_CLASS_LABELS.get(row.client_class, row.client_class)

    matches_qs = HoneypotMatch.objects.filter(source_code_email__path__startswith='/compliance/')
    matches = list(matches_qs.select_related('inbound_email', 'source_code_email').order_by('-matched_at')[:20])

    stats = compliance_stats.read(days)
    funnel, class_cols, class_rows, fanout = [], [], [], []
    top = {}
    if stats:
        attempts_started = stats['fanout'][0]['attempts'] if stats['fanout'] and stats['fanout'][0]['depth'] == 0 else 0
        reached5 = next((f['attempts'] for f in stats['fanout'] if f['depth'] == 4), 0)
        steps = [
            ('Course page views', stats['pages'].get('course', 0)),
            ('Attempts started (distinct)', attempts_started),
            ('Attempts reaching question 5', reached5),
            ('Attempts completed', summary['n'] or 0),
            ('Attempts passed', summary['passed'] or 0),
            ('Certificates viewed', event_counts.get('certificate', 0)),
            ('Certificate PDFs downloaded', event_counts.get('certificate_pdf', 0)),
        ]
        peak = max((v for _, v in steps), default=1) or 1
        funnel = [{'label': k, 'count': v, 'pct': max(1, round(v * 100 / peak)) if v else 0} for k, v in steps]
        classes = Counter()
        for per in stats['by_class'].values():
            classes.update(per)
        class_cols = [c for c, _ in classes.most_common()]
        for kind, label in _FUNNEL_KINDS:
            total = stats['pages'].get(kind, 0)
            if total:
                per = stats['by_class'].get(kind, {})
                class_rows.append({'label': label, 'total': total, 'cells': [per.get(c, 0) for c in class_cols]})
        class_cols = [_CLIENT_CLASS_LABELS.get(c, c) for c in class_cols]
        fanout = stats['fanout']
        fpeak = max((f['ratio'] for f in fanout), default=1) or 1
        for f in fanout:
            f['pct'] = round(f['ratio'] * 100 / fpeak)
        top = {
            'topics': _top_list(stats['topics']),
            'agencies': _top_list(stats['agencies']),
            'exhibits': _top_list(stats['exhibits']),
            'bots': _top_list(stats['bots']),
        }

    ctx = {
        'days': days,
        'random_score': COMPLIANCE_RANDOM_SCORE,
        'completed_n': summary['n'] or 0,
        'avg_score': round(summary['avg'] or 0),
        'pass_rate': round((summary['passed'] or 0) * 100 / summary['n']) if summary['n'] else 0,
        'certificates': event_counts.get('certificate', 0),
        'pdfs': event_counts.get('certificate_pdf', 0),
        'identified_n': event_counts.get('identified', 0),
        'learner_ips': stats['ips_total'] if stats else None,
        'page_total': sum(stats['pages'].values()) if stats else None,
        'score_hist': score_hist,
        'agents': agents,
        'leaderboard': leaderboard,
        'certs_daily': _daily_bars(certs_daily, 30),
        'pages_daily': stats['daily'] if stats else [],
        'identified': identified,
        'match_count': matches_qs.count(),
        'matches': matches,
        'funnel': funnel,
        'class_cols': class_cols,
        'class_rows': class_rows,
        'fanout': fanout,
        'top': top,
        'redis_ok': stats is not None,
    }
    return render(request, 'dashboard/compliance.html', ctx)
