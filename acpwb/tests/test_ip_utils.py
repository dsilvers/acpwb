from django.test import RequestFactory

from apps.core.ip_utils import get_client_ip

rf = RequestFactory()


def test_get_client_ip_no_xff_uses_remote_addr():
    request = rf.get('/', REMOTE_ADDR='10.0.0.5')
    assert get_client_ip(request) == '10.0.0.5'


def test_get_client_ip_uses_last_xff_entry_not_first():
    """nginx appends via $proxy_add_x_forwarded_for rather than replacing —
    the last entry is always the one nginx itself added, so it's the only
    value a client can't forge by sending its own X-Forwarded-For header."""
    request = rf.get('/', HTTP_X_FORWARDED_FOR='9.9.9.9, 203.0.113.7', REMOTE_ADDR='127.0.0.1')
    assert get_client_ip(request) == '203.0.113.7'


def test_get_client_ip_single_xff_entry():
    request = rf.get('/', HTTP_X_FORWARDED_FOR='203.0.113.7', REMOTE_ADDR='127.0.0.1')
    assert get_client_ip(request) == '203.0.113.7'


def test_get_client_ip_falls_back_when_xff_empty():
    request = rf.get('/', HTTP_X_FORWARDED_FOR='', REMOTE_ADDR='10.0.0.5')
    assert get_client_ip(request) == '10.0.0.5'


def test_get_client_ip_no_remote_addr_defaults():
    request = rf.get('/')
    del request.META['REMOTE_ADDR']
    assert get_client_ip(request) == '0.0.0.0'
