import io

import pytest
from django.core.management import call_command

from apps.honeypot.models import IPReputationScore, PublishedIPReputation


@pytest.mark.django_db
def test_lookup_requires_api_key(client, settings):
    settings.IP_REPUTATION_API_KEY = 'test-key'
    response = client.get('/reputation-api/ip/203.0.113.5/')
    assert response.status_code == 401


@pytest.mark.django_db
def test_lookup_rejects_wrong_key(client, settings):
    settings.IP_REPUTATION_API_KEY = 'test-key'
    response = client.get('/reputation-api/ip/203.0.113.5/', HTTP_X_API_KEY='wrong')
    assert response.status_code == 401


@pytest.mark.django_db
def test_lookup_fails_closed_when_unconfigured(client, settings):
    settings.IP_REPUTATION_API_KEY = ''
    response = client.get('/reputation-api/ip/203.0.113.5/', HTTP_X_API_KEY='')
    assert response.status_code == 401


@pytest.mark.django_db
def test_lookup_unscored_ip(client, settings):
    settings.IP_REPUTATION_API_KEY = 'test-key'
    response = client.get('/reputation-api/ip/203.0.113.5/', HTTP_X_API_KEY='test-key')
    assert response.status_code == 200
    assert response.json()['classification'] == 'unscored'


@pytest.mark.django_db
def test_lookup_returns_published_classification_not_internal_evidence(client, settings):
    settings.IP_REPUTATION_API_KEY = 'test-key'
    PublishedIPReputation.objects.create(ip_address='203.0.113.9', classification='residential_proxy_bot', confidence=85)

    response = client.get('/reputation-api/ip/203.0.113.9/', HTTP_X_API_KEY='test-key')
    data = response.json()
    assert data['classification'] == 'residential_proxy_bot'
    assert data['confidence'] == 85
    assert 'evidence' not in data


@pytest.mark.django_db
def test_lookup_rejects_invalid_ip(client, settings):
    settings.IP_REPUTATION_API_KEY = 'test-key'
    response = client.get('/reputation-api/ip/not-an-ip/', HTTP_X_API_KEY='test-key')
    assert response.status_code == 400


@pytest.mark.django_db
def test_publish_ip_reputation_syncs_without_evidence():
    IPReputationScore.objects.create(
        ip_address='203.0.113.10', classification='datacenter_bot',
        residential_proxy_score=70, evidence=['trap_tainted', 'header_mismatch'],
    )

    call_command('publish_ip_reputation', stdout=io.StringIO())

    published = PublishedIPReputation.objects.get(ip_address='203.0.113.10')
    assert published.classification == 'datacenter_bot'
    assert published.confidence == 70
    assert not hasattr(published, 'evidence')


@pytest.mark.django_db
def test_publish_ip_reputation_is_resumable_via_watermark():
    IPReputationScore.objects.create(ip_address='203.0.113.11', classification='human', residential_proxy_score=0)
    call_command('publish_ip_reputation', stdout=io.StringIO())
    from apps.core.models import DashboardStat
    watermark_after_first = DashboardStat.objects.get(key='ip_reputation_publish_watermark').value['id']

    call_command('publish_ip_reputation', stdout=io.StringIO())
    watermark_after_second = DashboardStat.objects.get(key='ip_reputation_publish_watermark').value['id']
    assert watermark_after_second == watermark_after_first
