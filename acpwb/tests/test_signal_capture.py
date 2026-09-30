import pytest

from apps.core.signal_capture import BROWSER_SIGNAL_HEADERS, browser_headers_present, fingerprint_sampled


def test_browser_headers_present_sorted_and_filtered():
    present = ['sec-fetch-mode', 'accept-language', 'x-some-junk-header']
    result = browser_headers_present(present)
    assert result == 'accept-language,sec-fetch-mode'


def test_browser_headers_present_empty():
    assert browser_headers_present([]) == ''


def test_browser_headers_present_none_match():
    assert browser_headers_present(['x-custom', 'x-other']) == ''


def test_browser_headers_present_all_candidates():
    result = browser_headers_present(BROWSER_SIGNAL_HEADERS)
    assert result == ','.join(sorted(BROWSER_SIGNAL_HEADERS))


@pytest.mark.parametrize("ip,rate,expected", [
    # Same cases as acpwb_go/visitqueue/visitqueue_test.go — both backends
    # must sample the same IPs.
    ('203.0.113.9', 0.51, True),   # bucket fraction 0.5084
    ('203.0.113.9', 0.50, False),
    ('2001:db8::1', 0.12, True),   # 0.1148
    ('2001:db8::1', 0.11, False),
    ('8.8.8.8', 0.26, True),       # 0.2539
    ('8.8.8.8', 0.25, False),
    ('45.148.10.19', 1.0, True),
    ('45.148.10.19', 0.0, False),
])
def test_fingerprint_sampled_matches_go(ip, rate, expected):
    assert fingerprint_sampled(ip, rate) is expected


def test_fingerprint_sampled_rate_is_roughly_proportional():
    ips = [f'10.{i // 256}.{i % 256}.1' for i in range(20000)]
    share = sum(fingerprint_sampled(ip, 0.1) for ip in ips) / len(ips)
    assert 0.09 < share < 0.11
