import pytest
from unittest.mock import patch
from django.test import override_settings
from apps.honeypot.models import CrawlerVisit, CanaryToken, InternalLoginAttempt

_no_redis_crawler = patch('apps.core.crawler_queue.push_crawler_visit', return_value=False)


# ── .env probe ─────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_env_probe_returns_200_with_credentials(client):
    response = client.get('/.env')
    assert response.status_code == 200
    assert b'AWS_ACCESS_KEY_ID' in response.content


@pytest.mark.django_db
def test_env_probe_creates_canary_token(client):
    client.get('/.env')
    assert CanaryToken.objects.filter(token_type='env_url').exists()


@pytest.mark.django_db
def test_env_probe_logs_crawlervisit(client):
    with _no_redis_crawler:
        client.get('/.env')
    assert CrawlerVisit.objects.filter(trap_type='env_probe').exists()


# ── wp-config.php probe ────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_wp_config_returns_200_with_db_password(client):
    response = client.get('/wp-config.php')
    assert response.status_code == 200
    assert b'DB_PASSWORD' in response.content


@pytest.mark.django_db
def test_wp_config_logs_crawlervisit(client):
    with _no_redis_crawler:
        client.get('/wp-config.php')
    assert CrawlerVisit.objects.filter(trap_type='wp_probe').exists()


# ── wp-login.php ───────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_wp_login_get_returns_form(client):
    response = client.get('/wp-login.php')
    assert response.status_code == 200
    assert b'<form' in response.content
    assert b'log' in response.content  # username field name


@pytest.mark.django_db
def test_wp_login_post_logs_credentials(client):
    client.post('/wp-login.php', {'log': 'admin', 'pwd': 'hunter2'})
    attempt = InternalLoginAttempt.objects.last()
    assert attempt is not None
    assert attempt.username == 'admin'
    assert attempt.password == 'hunter2'


@pytest.mark.django_db
def test_wp_login_post_redirects(client):
    response = client.post('/wp-login.php', {'log': 'admin', 'pwd': 'secret'})
    assert response.status_code == 302


# ── xmlrpc.php ─────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_xmlrpc_get_returns_plaintext(client):
    response = client.get('/xmlrpc.php')
    assert response.status_code == 200
    assert b'POST' in response.content


@pytest.mark.django_db
def test_xmlrpc_post_returns_fault(client):
    body = (
        '<?xml version="1.0"?>'
        '<methodCall><methodName>wp.getUsersBlogs</methodName>'
        '<params><param><value>admin</value></param>'
        '<param><value>password123</value></param></params>'
        '</methodCall>'
    )
    response = client.post('/xmlrpc.php', data=body, content_type='text/xml')
    assert response.status_code == 200
    assert b'faultCode' in response.content


@pytest.mark.django_db
def test_xmlrpc_post_logs_credentials(client):
    body = (
        '<?xml version="1.0"?>'
        '<methodCall><methodName>wp.getUsersBlogs</methodName>'
        '<params><param><value><string>admin</string></value></param>'
        '<param><value><string>s3cr3t</string></value></param></params>'
        '</methodCall>'
    )
    client.post('/xmlrpc.php', data=body, content_type='text/xml')
    attempt = InternalLoginAttempt.objects.filter(username='admin').last()
    assert attempt is not None
    assert attempt.password == 's3cr3t'


# ── webshell *.php ─────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_webshell_with_cmd_param_returns_fake_output(client):
    response = client.get('/shell.php?cmd=id')
    assert response.status_code == 200
    assert b'www-data' in response.content


@pytest.mark.django_db
def test_webshell_with_cmd_logs_query_string(client):
    with _no_redis_crawler:
        client.get('/shell.php?cmd=whoami')
    visit = CrawlerVisit.objects.filter(trap_type='webshell_probe').last()
    assert visit is not None
    assert 'cmd=whoami' in visit.query_string


@pytest.mark.django_db
def test_webshell_without_cmd_returns_php_error(client):
    response = client.get('/notreal.php')
    assert response.status_code == 200
    assert b'Fatal error' in response.content or b'Parse error' in response.content or b'Warning' in response.content


# ── .git/config ────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_git_config_returns_200(client):
    response = client.get('/.git/config')
    assert response.status_code == 200
    assert b'[core]' in response.content


# ── .htpasswd ──────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_htpasswd_returns_200(client):
    response = client.get('/.htpasswd')
    assert response.status_code == 200
    assert b'admin' in response.content


# ── canary ping ────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_canary_ping_marks_triggered(client):
    token = CanaryToken.objects.create(
        token='testtoken123',
        token_type='env_url',
    )
    response = client.get(f'/.well-known/tokens/testtoken123/ping')
    assert response.status_code == 200
    token.refresh_from_db()
    assert token.triggered is True


@pytest.mark.django_db
def test_canary_ping_unknown_token_returns_404(client):
    response = client.get('/.well-known/tokens/doesnotexist/ping')
    assert response.status_code == 404


@pytest.mark.django_db
def test_canary_ping_logs_crawlervisit(client):
    CanaryToken.objects.create(token='pingtoken', token_type='env_url')
    with _no_redis_crawler:
        client.get('/.well-known/tokens/pingtoken/ping')
    assert CrawlerVisit.objects.filter(trap_type='canary_trigger').exists()


# ── handler404 ─────────────────────────────────────────────────────────────────

@pytest.mark.django_db
@override_settings(DEBUG=False)
def test_scanner_probe_404_logs_visit(client):
    before = CrawlerVisit.objects.filter(trap_type='scanner_probe').count()
    with _no_redis_crawler:
        client.get('/backup.sql')
    after = CrawlerVisit.objects.filter(trap_type='scanner_probe').count()
    assert after == before + 1

