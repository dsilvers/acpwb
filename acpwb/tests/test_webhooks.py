import hashlib
import hmac
import json
import time

import pytest
from apps.webhooks.models import InboundEmail, HoneypotMatch
from apps.webhooks.views import _verify_mailgun_signature
from apps.people.models import PeoplePageVisit, GeneratedEmployee


# ── Signature verification ─────────────────────────────────────────────────────

def test_verify_valid_signature():
    key = 'my-signing-key'
    token = 'abc123'
    timestamp = '1700000000'
    value = f"{timestamp}{token}".encode('utf-8')
    sig = hmac.new(key.encode('utf-8'), value, hashlib.sha256).hexdigest()
    assert _verify_mailgun_signature(key, token, timestamp, sig) is True


def test_verify_invalid_signature():
    assert _verify_mailgun_signature('key', 'token', '123', 'badsig') is False


def test_verify_empty_key_allows_through():
    # When key is not configured, should allow (dev mode)
    assert _verify_mailgun_signature('', 'token', '123', 'anysig') is True


def test_verify_none_key_allows_through():
    assert _verify_mailgun_signature(None, 'token', '123', 'anysig') is True


# ── Webhook view ───────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_mailgun_inbound_creates_email_record(mailgun_post):
    assert InboundEmail.objects.count() == 0
    response = mailgun_post('test@acpwb.com')
    assert response.status_code == 200
    assert InboundEmail.objects.count() == 1


@pytest.mark.django_db
def test_mailgun_inbound_rejects_bad_signature(client, settings):
    settings.MAILGUN_WEBHOOK_SIGNING_KEY = 'real-key'
    response = client.post('/webhooks/mailgun/inbound/', {
        'timestamp': '1700000000',
        'token': 'abc',
        'signature': 'WRONG',
        'sender': 'x@x.com',
        'recipient': 'y@acpwb.com',
    })
    assert response.status_code == 406
    assert InboundEmail.objects.count() == 0


@pytest.mark.django_db
def test_mailgun_inbound_creates_no_match_when_unknown_recipient(mailgun_post):
    response = mailgun_post('nobody@acpwb.com')
    assert response.status_code == 200
    match = HoneypotMatch.objects.first()
    assert match is not None
    assert match.match_confidence == 'none'


@pytest.mark.django_db
def test_mailgun_inbound_creates_exact_match(mailgun_post):
    # Set up a known generated employee
    visit = PeoplePageVisit.objects.create(
        ip_address='1.2.3.4',
        user_agent='Mozilla/5.0',
        referrer='',
        session_key='session-xyz',
    )
    GeneratedEmployee.objects.create(
        visit=visit,
        first_name='Alice',
        last_name='Walker',
        email='alice.walker@acpwb.com',
        title='Director',
        department='Operations',
        avatar_seed='seed123',
    )

    response = mailgun_post('alice.walker@acpwb.com')
    assert response.status_code == 200

    match = HoneypotMatch.objects.filter(match_confidence='exact').first()
    assert match is not None
    assert match.generated_employee.email == 'alice.walker@acpwb.com'
    assert match.original_visit == visit


@pytest.mark.django_db
def test_mailgun_inbound_match_case_insensitive(mailgun_post):
    visit = PeoplePageVisit.objects.create(
        ip_address='1.2.3.4',
        user_agent='test',
        referrer='',
        session_key='sess',
    )
    GeneratedEmployee.objects.create(
        visit=visit,
        first_name='Bob',
        last_name='Jones',
        email='bob.jones@acpwb.com',
        title='Manager',
        department='HR',
        avatar_seed='seed456',
    )

    response = mailgun_post('BOB.JONES@ACPWB.COM')
    assert response.status_code == 200

    match = HoneypotMatch.objects.filter(match_confidence='exact').first()
    assert match is not None


@pytest.mark.django_db
def test_mailgun_post_only(client):
    response = client.get('/webhooks/mailgun/inbound/')
    assert response.status_code == 405


# ── Cloudflare pipe webhook ─────────────────────────────────────────────────────

PIPE_SECRET = 'test-pipe-secret'
PIPE_RAW = "From: spammer@evil.com\r\nTo: alice@acpwb.com\r\nSubject: Hello\r\n\r\nBody text"


@pytest.fixture
def pipe_post(client, settings):
    settings.PIPE_WEBHOOK_SECRET = PIPE_SECRET

    def _post(body):
        return client.post(
            '/webhooks/pipe/inbound/',
            data=json.dumps(body),
            content_type='application/json',
            HTTP_X_WEBHOOK_SECRET=PIPE_SECRET,
        )

    return _post


@pytest.mark.django_db
def test_pipe_inbound_creates_email(pipe_post):
    resp = pipe_post({
        'sender': 'spammer@evil.com',
        'recipient': 'alice@acpwb.com',
        'subject': 'Hello',
        'raw': PIPE_RAW,
    })
    assert resp.status_code == 200
    assert InboundEmail.objects.filter(sender='spammer@evil.com').exists()


@pytest.mark.django_db
def test_pipe_inbound_rejects_bad_secret(client, settings):
    settings.PIPE_WEBHOOK_SECRET = PIPE_SECRET
    resp = client.post(
        '/webhooks/pipe/inbound/',
        data=json.dumps({'sender': 'x@x.com', 'recipient': 'y@acpwb.com', 'subject': '', 'raw': ''}),
        content_type='application/json',
        HTTP_X_WEBHOOK_SECRET='wrong-secret',
    )
    assert resp.status_code == 403
    assert InboundEmail.objects.count() == 0


@pytest.mark.django_db
def test_pipe_inbound_rejects_get(client):
    resp = client.get('/webhooks/pipe/inbound/')
    assert resp.status_code == 405


@pytest.mark.django_db
def test_pipe_inbound_rejects_bad_json(client, settings):
    settings.PIPE_WEBHOOK_SECRET = PIPE_SECRET
    resp = client.post(
        '/webhooks/pipe/inbound/',
        data='not-json',
        content_type='application/json',
        HTTP_X_WEBHOOK_SECRET=PIPE_SECRET,
    )
    assert resp.status_code == 400


@pytest.mark.django_db
def test_pipe_inbound_creates_match_for_known_email(pipe_post):
    visit = PeoplePageVisit.objects.create(
        ip_address='1.2.3.4',
        user_agent='Mozilla/5.0',
        referrer='',
        session_key='session-pipe',
    )
    GeneratedEmployee.objects.create(
        visit=visit,
        first_name='Carol',
        last_name='Price',
        email='carol.price@acpwb.com',
        title='Analyst',
        department='Finance',
        avatar_seed='seedpipe',
    )
    resp = pipe_post({
        'sender': 'spammer@evil.com',
        'recipient': 'carol.price@acpwb.com',
        'subject': 'Hi',
        'raw': PIPE_RAW,
    })
    assert resp.status_code == 200
    match = HoneypotMatch.objects.filter(match_confidence='exact').first()
    assert match is not None
    assert match.original_visit == visit
