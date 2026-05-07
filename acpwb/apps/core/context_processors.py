import hashlib
import time


def honeypot_context(request):
    """Inject per-request context used across all templates."""
    token = hashlib.md5(
        f"{request.path}{time.time()}{request.META.get('REMOTE_ADDR', '')}".encode()
    ).hexdigest()[:8]
    # On archive subdomains, header/footer links must be absolute so they go to
    # the main domain rather than resolving against the subdomain.
    on_sub = (getattr(request, 'on_archive_subdomain', False)
              or getattr(request, 'on_policy_subdomain', False))
    site_root = 'https://acpwb.com' if on_sub else ''
    return {
        'honeypot_token': token,
        'site_root': site_root,
    }
