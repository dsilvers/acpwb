import io

import pytest
from django.core.management import call_command
from django.utils import timezone

from apps.core.reputation_scoring import compute_score, is_old_tls_protocol, looks_like_modern_browser
from apps.honeypot.models import CrawlerVisit, IPIntelligence, IPReputationScore, RequestFingerprint

CHROME_UA = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36'
SAFARI_UA = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15'


# ── looks_like_modern_browser ───────────────────────────────────────────────

@pytest.mark.parametrize('ua', [CHROME_UA, SAFARI_UA, 'Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0'])
def test_looks_like_modern_browser_true(ua):
    assert looks_like_modern_browser(ua) is True


@pytest.mark.parametrize('ua', ['python-requests/2.28.0', 'curl/7.88.1', 'Go-http-client/1.1', 'GPTBot/1.0', '', None])
def test_looks_like_modern_browser_false(ua):
    assert looks_like_modern_browser(ua) is False


# ── is_old_tls_protocol ─────────────────────────────────────────────────────

def test_is_old_tls_protocol():
    assert is_old_tls_protocol('TLSv1') is True
    assert is_old_tls_protocol('TLSv1.1') is True
    assert is_old_tls_protocol('TLSv1.3') is False
    assert is_old_tls_protocol('') is False


# ── compute_score ────────────────────────────────────────────────────────────

def test_compute_score_no_signals_is_human():
    score, classification, evidence = compute_score(
        trap_tainted=False, header_mismatch=False, protocol_mismatch=False,
        tls_mismatch=False, new_ip_no_referrer=False, is_hosting=None,
    )
    assert score == 0
    assert classification == 'human'
    assert evidence == []


def test_compute_score_trap_tainted_alone_crosses_threshold():
    score, classification, evidence = compute_score(
        trap_tainted=True, header_mismatch=False, protocol_mismatch=False,
        tls_mismatch=False, new_ip_no_referrer=False, is_hosting=None,
    )
    assert score == 40
    assert classification == 'unknown'  # score high, but ASN not yet known
    assert evidence == ['trap_tainted']


def test_compute_score_classifies_datacenter_when_hosting_true():
    score, classification, evidence = compute_score(
        trap_tainted=True, header_mismatch=True, protocol_mismatch=False,
        tls_mismatch=False, new_ip_no_referrer=False, is_hosting=True,
    )
    assert classification == 'datacenter_bot'
    assert score == 65


def test_compute_score_classifies_residential_proxy_when_hosting_false():
    score, classification, evidence = compute_score(
        trap_tainted=True, header_mismatch=True, protocol_mismatch=True,
        tls_mismatch=False, new_ip_no_referrer=False, is_hosting=False,
    )
    assert classification == 'residential_proxy_bot'
    assert score == 85


def test_compute_score_caps_at_100():
    score, _, _ = compute_score(
        trap_tainted=True, header_mismatch=True, protocol_mismatch=True,
        tls_mismatch=True, new_ip_no_referrer=True, is_hosting=False,
    )
    assert score == 100


def test_compute_score_below_threshold_stays_human_even_with_asn_data():
    score, classification, evidence = compute_score(
        trap_tainted=False, header_mismatch=False, protocol_mismatch=False,
        tls_mismatch=False, new_ip_no_referrer=True, is_hosting=False,
    )
    assert score == 10
    assert classification == 'human'


# ── score_ip_reputation command ─────────────────────────────────────────────

@pytest.mark.django_db
def test_score_ip_reputation_flags_trap_tainted_ip():
    now = timezone.now()
    CrawlerVisit.objects.create(ip_address='203.0.113.50', path='/internal/', trap_type='ghost_link', timestamp=now)
    RequestFingerprint.objects.create(
        ip_address='203.0.113.50', user_agent=CHROME_UA, timestamp=now,
        client_protocol='HTTP/1.1', browser_headers_present='',
    )

    call_command('score_ip_reputation', stdout=io.StringIO())

    score = IPReputationScore.objects.get(ip_address='203.0.113.50')
    assert 'trap_tainted' in score.evidence
    assert 'header_mismatch' in score.evidence
    assert 'protocol_mismatch' in score.evidence
    assert score.first_flagged_at is not None


@pytest.mark.django_db
def test_score_ip_reputation_classifies_using_ip_intelligence():
    now = timezone.now()
    CrawlerVisit.objects.create(ip_address='203.0.113.60', path='/internal/', trap_type='canary_trigger', timestamp=now)
    IPIntelligence.objects.create(ip_address='203.0.113.60', is_hosting=False, lookup_ok=True)
    RequestFingerprint.objects.create(ip_address='203.0.113.60', user_agent=CHROME_UA, timestamp=now)

    call_command('score_ip_reputation', stdout=io.StringIO())

    score = IPReputationScore.objects.get(ip_address='203.0.113.60')
    assert score.classification == 'residential_proxy_bot'


@pytest.mark.django_db
def test_score_ip_reputation_is_resumable_via_watermark():
    now = timezone.now()
    RequestFingerprint.objects.create(ip_address='203.0.113.70', user_agent='curl/8.0', timestamp=now)
    call_command('score_ip_reputation', stdout=io.StringIO())
    from apps.core.models import DashboardStat
    watermark_after_first = DashboardStat.objects.get(key='ip_reputation_score_watermark').value['id']

    # A second run with nothing new must not re-scan already-processed rows.
    call_command('score_ip_reputation', stdout=io.StringIO())
    watermark_after_second = DashboardStat.objects.get(key='ip_reputation_score_watermark').value['id']
    assert watermark_after_second == watermark_after_first


@pytest.mark.django_db
def test_score_ip_reputation_plain_bot_ua_not_flagged_for_header_mismatch():
    """looks_like_modern_browser() gates the header/protocol/TLS mismatch
    checks — a UA that doesn't claim to be a browser at all shouldn't get
    penalized for missing browser-only headers; that's what bot_classify
    (UA substring matching) already handles."""
    now = timezone.now()
    RequestFingerprint.objects.create(
        ip_address='203.0.113.80', user_agent='python-requests/2.28.0', timestamp=now,
        client_protocol='HTTP/1.1', browser_headers_present='',
    )

    call_command('score_ip_reputation', stdout=io.StringIO())

    score = IPReputationScore.objects.get(ip_address='203.0.113.80')
    assert score.evidence == []
    assert score.classification == 'human'
