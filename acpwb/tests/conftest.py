import hashlib
import hmac
import time

import pytest
from django.test import Client


@pytest.fixture
def client():
    return Client()


@pytest.fixture
def bot_client():
    """Client that sends a bot user-agent."""
    c = Client()
    c.defaults['HTTP_USER_AGENT'] = 'Googlebot/2.1 (+http://www.google.com/bot.html)'
    return c


def make_mailgun_sig(signing_key, token, timestamp):
    """Produce a valid Mailgun HMAC-SHA256 signature."""
    value = f"{timestamp}{token}".encode('utf-8')
    return hmac.new(
        signing_key.encode('utf-8'),
        value,
        hashlib.sha256,
    ).hexdigest()


@pytest.fixture
def mailgun_post(client, settings):
    """Return a helper that POSTs a signed Mailgun payload."""
    settings.MAILGUN_WEBHOOK_SIGNING_KEY = 'test-signing-key'

    def _post(recipient, sender='spammer@example.com', subject='Buy now'):
        timestamp = str(int(time.time()))
        token = 'abc123token'
        sig = make_mailgun_sig('test-signing-key', token, timestamp)
        return client.post('/webhooks/mailgun/inbound/', {
            'timestamp': timestamp,
            'token': token,
            'signature': sig,
            'sender': sender,
            'recipient': recipient,
            'subject': subject,
            'body-plain': 'Spam body text',
            'body-html': '<p>Spam body text</p>',
            'Message-Id': '<msg-id@mailgun>',
        })

    return _post
