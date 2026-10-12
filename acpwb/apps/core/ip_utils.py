def get_client_ip(request):
    """Best-effort real client IP behind acpwb_edge.

    The edge is the only proxy hop (no CDN in front) and replaces any
    client-sent X-Forwarded-For with the single address it accepted the
    connection from. Taking the last entry is still the safe choice: it's
    the one the proxy writes, so it also holds behind a proxy that appends
    instead (the old nginx $proxy_add_x_forwarded_for setup), whereas the
    first entry — as this codebase used to take everywhere — lets a client
    spoof the logged IP with its own X-Forwarded-For header.
    """
    x_forwarded = request.META.get('HTTP_X_FORWARDED_FOR')
    if x_forwarded:
        parts = [p.strip() for p in x_forwarded.split(',') if p.strip()]
        if parts:
            return parts[-1]
    return request.META.get('REMOTE_ADDR') or '0.0.0.0'
