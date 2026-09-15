from apps.core.signal_capture import BROWSER_SIGNAL_HEADERS, browser_headers_present


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
