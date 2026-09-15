"""The external-facing surface of the traffic-intelligence product — reads
only PublishedIPReputation (classification + confidence), never
IPReputationScore's internal evidence. See PublishedIPReputation's docstring
for why the two are kept separate.

Deliberately minimal for a first cut: single-IP lookup, static API-key
header auth. No rate limiting, billing, or bulk export yet — those are real
product decisions for once there's an actual customer, not something to
guess at now.
"""
import ipaddress

from django.conf import settings
from django.http import JsonResponse

from apps.honeypot.models import PublishedIPReputation


def _authorized(request):
    expected = getattr(settings, 'IP_REPUTATION_API_KEY', '')
    if not expected:
        return False  # unconfigured — fail closed, not open
    return request.META.get('HTTP_X_API_KEY', '') == expected


def ip_reputation_lookup(request, ip_address):
    if not _authorized(request):
        return JsonResponse({'error': 'unauthorized'}, status=401)

    try:
        ipaddress.ip_address(ip_address)
    except ValueError:
        return JsonResponse({'error': 'invalid IP address'}, status=400)

    row = PublishedIPReputation.objects.filter(ip_address=ip_address).first()
    if row is None:
        return JsonResponse({
            'ip_address': ip_address,
            'classification': 'unscored',
            'confidence': 0,
        })

    return JsonResponse({
        'ip_address': row.ip_address,
        'classification': row.classification,
        'confidence': row.confidence,
        'published_at': row.published_at.isoformat(),
    })
