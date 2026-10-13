import io
from unittest.mock import MagicMock, patch

import pytest
from django.core.management import call_command

from apps.core.crawler_queue import push_fingerprint_signal
from apps.honeypot.models import RequestFingerprint


@pytest.fixture
def capture_on(settings):
    settings.FINGERPRINT_CAPTURE_ENABLED = True
    settings.FINGERPRINT_SAMPLE_RATE = 1.0


@pytest.mark.django_db
def test_request_stream_middleware_queues_fingerprint_signal(client, capture_on):
    with patch('apps.core.crawler_queue.check_first_seen_ip', return_value=True) as mock_first_seen, \
         patch('apps.core.crawler_queue.queue_fingerprint_signal') as mock_queue:
        client.get(
            '/careers/',
            HTTP_USER_AGENT='Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0 Safari/537.36',
            HTTP_ACCEPT_LANGUAGE='en-US',
            HTTP_SEC_FETCH_SITE='none',
            HTTP_X_CLIENT_PROTOCOL='HTTP/1.1',
        )

    assert mock_queue.called
    data = mock_queue.call_args[0][0]
    assert data['first_seen_ip'] is True
    assert data['client_protocol'] == 'HTTP/1.1'
    assert 'accept-language' in data['browser_headers_present']
    assert 'sec-fetch-site' in data['browser_headers_present']
    mock_first_seen.assert_called_once()


@pytest.mark.django_db
def test_request_stream_middleware_runs_on_every_path_not_just_bot_uas(client, capture_on):
    """Residential-proxy traffic is exactly the traffic that looks like a
    normal browser — signal capture must not be gated behind
    BotTrackingMiddleware's bot-UA pattern the way CrawlerVisit logging is."""
    with patch('apps.core.crawler_queue.check_first_seen_ip', return_value=False), \
         patch('apps.core.crawler_queue.queue_fingerprint_signal') as mock_queue:
        client.get('/careers/', HTTP_USER_AGENT='Mozilla/5.0 (a normal-looking browser)')

    assert mock_queue.called


@pytest.mark.django_db
def test_drain_fingerprint_queue_persists_rows():
    push_fingerprint_signal({
        'timestamp': '2026-01-01T00:00:00.000000+00:00',
        'ip_address': '203.0.113.9',
        'host': 'acpwb.com',
        'user_agent': 'curl/8.0',
        'referrer_present': False,
        'first_seen_ip': True,
        'client_protocol': 'HTTP/1.1',
        'tls_protocol': '',
        'tls_cipher': '',
        'browser_headers_present': '',
    })

    call_command('drain_fingerprint_queue', stdout=io.StringIO())

    row = RequestFingerprint.objects.get(ip_address='203.0.113.9')
    assert row.user_agent == 'curl/8.0'
    assert row.first_seen_ip is True
    assert row.referrer_present is False


@pytest.mark.django_db
def test_drain_fingerprint_queue_dedupes_by_idempotency_key():
    """RequestFingerprint is a brand-new table (not an existing hypertable),
    so — unlike CrawlerVisit — idempotency_key gets a real unique constraint
    and bulk_create(ignore_conflicts=True) can dedupe on its own."""
    RequestFingerprint.objects.create(
        ip_address='198.51.100.4', idempotency_key='11111111-1111-1111-1111-111111111111',
    )
    push_fingerprint_signal({
        'timestamp': '2026-01-01T00:00:00.000000+00:00',
        'ip_address': '198.51.100.4',
        'idempotency_key': '11111111-1111-1111-1111-111111111111',
    })
    call_command('drain_fingerprint_queue', stdout=io.StringIO())

    assert RequestFingerprint.objects.filter(ip_address='198.51.100.4').count() == 1


@pytest.mark.django_db
def test_fingerprint_capture_is_off_by_default(client):
    """Default-off: nothing touches Redis — not even the per-IP first-seen
    key — unless FINGERPRINT_CAPTURE_ENABLED is set."""
    with patch('apps.core.crawler_queue.check_first_seen_ip') as mock_first_seen, \
         patch('apps.core.crawler_queue.queue_fingerprint_signal') as mock_queue:
        client.get('/careers/', HTTP_USER_AGENT='Mozilla/5.0 (a normal-looking browser)')

    assert not mock_queue.called
    assert not mock_first_seen.called


@pytest.mark.django_db
def test_fingerprint_capture_skips_unsampled_ips(client, settings):
    settings.FINGERPRINT_CAPTURE_ENABLED = True
    settings.FINGERPRINT_SAMPLE_RATE = 0.0
    with patch('apps.core.crawler_queue.check_first_seen_ip') as mock_first_seen, \
         patch('apps.core.crawler_queue.queue_fingerprint_signal') as mock_queue:
        client.get('/careers/', HTTP_USER_AGENT='Mozilla/5.0 (a normal-looking browser)')

    assert not mock_queue.called
    assert not mock_first_seen.called


def test_push_fingerprint_signal_caps_queue_length(settings):
    settings.FINGERPRINT_QUEUE_MAX = 5
    pipe = MagicMock()
    fake_redis = MagicMock()
    fake_redis.pipeline.return_value = pipe
    with patch('apps.core.crawler_queue._get_client', return_value=fake_redis):
        assert push_fingerprint_signal({'ip_address': '203.0.113.9'}) is True

    pipe.rpush.assert_called_once()
    pipe.ltrim.assert_called_once_with('acpwb:fingerprint_queue', -5, -1)
    pipe.execute.assert_called_once()


@pytest.mark.django_db
def test_middleware_forwards_edge_fingerprint_headers(client, capture_on):
    with patch('apps.core.crawler_queue.check_first_seen_ip', return_value=False), \
         patch('apps.core.crawler_queue.queue_fingerprint_signal') as mock_queue:
        client.get(
            '/careers/',
            HTTP_X_CLIENT_PROTOCOL='HTTP/2.0',
            HTTP_X_TLS_PROTOCOL='TLSv1.3',
            HTTP_X_JA4='t13d1516h2_8daaf6152771_e5627efa2ab1',
            HTTP_X_JA3='0123456789abcdef0123456789abcdef',
            HTTP_X_JA4H='ge20nn09enus_' + 'a' * 12 + '_000000000000_000000000000',
            HTTP_X_H2_FINGERPRINT='1:65536;2:0;4:6291456;6:262144|15663105|0|m,a,s,p',
        )

    data = mock_queue.call_args[0][0]
    assert data['request_id'] is None  # no X-Request-Id: no edge in front
    assert data['ja4'] == 't13d1516h2_8daaf6152771_e5627efa2ab1'
    assert data['ja3'] == '0123456789abcdef0123456789abcdef'
    assert data['ja4h'].startswith('ge20nn09enus_')
    assert data['h2_fingerprint'].endswith('|m,a,s,p')


@pytest.mark.django_db
def test_drain_fingerprint_queue_drops_unknown_keys_instead_of_rows():
    """A producer shipping a new payload key before this side's migration
    must not lose the whole row."""
    push_fingerprint_signal({
        'timestamp': '2026-01-01T00:00:00.000000+00:00',
        'ip_address': '203.0.113.77',
        'ja4': 't13d1516h2_8daaf6152771_e5627efa2ab1',
        'some_future_field': 'x',
    })
    call_command('drain_fingerprint_queue', stdout=io.StringIO())

    row = RequestFingerprint.objects.get(ip_address='203.0.113.77')
    assert row.ja4 == 't13d1516h2_8daaf6152771_e5627efa2ab1'


def _push_edge_conn(record):
    import json
    from apps.core import crawler_queue
    crawler_queue._get_client().rpush(crawler_queue._EDGE_CONN_QUEUE_KEY, json.dumps(record))


@pytest.fixture
def private_edge_queue(monkeypatch):
    """The dev stack's edge pushes to the same Redis: give the test a queue
    of its own so it neither drains nor depends on live records."""
    import uuid
    from apps.core import crawler_queue
    monkeypatch.setattr(crawler_queue, '_EDGE_CONN_QUEUE_KEY', f'test:edge_conn_queue:{uuid.uuid4()}')
    monkeypatch.setattr(crawler_queue, '_EDGE_REQUEST_QUEUE_KEY', f'test:edge_request_queue:{uuid.uuid4()}')


@pytest.mark.django_db
def test_drain_edge_conn_queue_persists_and_dedupes(private_edge_queue):
    from apps.honeypot.models import EdgeConnection
    record = {
        'timestamp': '2026-01-01T00:00:00.000000+00:00',
        'ip_address': '192.0.2.10',
        'outcome': 'handshake_failed',
        'ja4': 't10i0607ht_aaaaaaaaaaaa_bbbbbbbbbbbb',
        'ja3_str': '769,47-53,0-10-11,23-24,0',
        'bytes_in': 517,
        'idempotency_key': '22222222-2222-2222-2222-222222222222',
        'unknown_future_key': 1,
    }
    _push_edge_conn(record)
    _push_edge_conn(record)  # a replayed batch must not duplicate
    call_command('drain_edge_conn_queue', stdout=io.StringIO())

    rows = EdgeConnection.objects.filter(ip_address='192.0.2.10')
    assert rows.count() == 1
    assert rows[0].outcome == 'handshake_failed'
    assert rows[0].bytes_in == 517


@pytest.mark.django_db
def test_drain_edge_request_queue_persists_and_dedupes(private_edge_queue):
    import json
    from apps.core import crawler_queue
    from apps.honeypot.models import EdgeRequest
    record = {
        'timestamp': '2026-01-01T00:00:00.000000+00:00',
        'request_id': '0f8fad5b-d9cb-469f-a165-70867728950e',
        'conn_id': '7c9e6679-7425-40de-944b-e07fc1f90ae7',
        'client_id': 'AAAAAAAAAAAAAAAAAAAAAA',
        'seq': 2,
        'ip_address': '192.0.2.11',
        'protocol': 'HTTP/2.0',
        'method': 'GET', 'host': 'acpwb.com', 'path': '/static/css/site.css', 'query': '',
        'status': 200, 'route': 'static', 'error': '',
        'response_bytes': 1234, 'duration_us': 900,
        'unknown_future_key': 1,
    }
    r = crawler_queue._get_client()
    r.rpush(crawler_queue._EDGE_REQUEST_QUEUE_KEY, json.dumps(record))
    r.rpush(crawler_queue._EDGE_REQUEST_QUEUE_KEY, json.dumps(record))  # replay
    call_command('drain_edge_request_queue', stdout=io.StringIO())

    rows = EdgeRequest.objects.filter(ip_address='192.0.2.11')
    assert rows.count() == 1
    assert str(rows[0].conn_id) == '7c9e6679-7425-40de-944b-e07fc1f90ae7'
    assert rows[0].route == 'static'
    assert rows[0].seq == 2


@pytest.mark.django_db
def test_crawler_visit_carries_edge_request_id(client):
    """The edge's X-Request-Id lands on the CrawlerVisit, joining it to
    EdgeRequest."""
    from apps.honeypot.models import CrawlerVisit
    rid = '0f8fad5b-d9cb-469f-a165-70867728950e'
    with patch('apps.core.crawler_queue.push_crawler_visit', return_value=False):
        client.get('/.env', HTTP_X_REQUEST_ID=rid, HTTP_USER_AGENT='python-requests/2.31')
    visits = CrawlerVisit.objects.filter(path='/.env')  # the probe view and the bot middleware both log
    assert visits.exists()
    assert {str(v.request_id) for v in visits} == {rid}


@pytest.mark.django_db
def test_malformed_request_id_is_ignored(client):
    from apps.honeypot.models import CrawlerVisit
    with patch('apps.core.crawler_queue.push_crawler_visit', return_value=False):
        client.get('/.env', HTTP_X_REQUEST_ID="nope'; drop table", HTTP_USER_AGENT='python-requests/2.31')
    assert set(CrawlerVisit.objects.filter(path='/.env').values_list('request_id', flat=True)) == {None}


def test_interrupted_batch_is_recovered_without_scanning(private_edge_queue):
    """A drain that dies between pop and finalize leaves its batch in the
    queue's in-flight set; recovery finds it there (no keyspace SCAN), and
    finalizing forgets it."""
    import json
    from apps.core import crawler_queue
    r = crawler_queue._get_client()
    key = crawler_queue._EDGE_REQUEST_QUEUE_KEY
    r.rpush(key, json.dumps({'n': 1}), json.dumps({'n': 2}))
    items, batch_key = crawler_queue.pop_edge_requests(10)
    assert [i['n'] for i in items] == [1, 2]

    recovered = crawler_queue.recover_edge_requests()  # "crashed" before finalize
    assert [(sorted(i['n'] for i in its), k) for its, k in recovered] == [([1, 2], batch_key)]

    crawler_queue.finalize_batch(batch_key)
    assert crawler_queue.recover_edge_requests() == []
    assert not r.exists(batch_key)
    r.delete(f'{key}:processing_batches', f'{key}:processing_scanned')
