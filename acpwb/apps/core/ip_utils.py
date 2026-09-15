def get_client_ip(request):
    """Best-effort real client IP behind nginx.

    nginx is the only proxy hop (no CDN in front) and sets X-Forwarded-For via
    $proxy_add_x_forwarded_for, which *appends* $remote_addr to whatever
    X-Forwarded-For the client already sent rather than replacing it. Taking
    the first entry — as this codebase used to everywhere — lets a client
    spoof the logged IP by sending its own X-Forwarded-For header. The last
    entry is always the one nginx itself appended, so it's the only value
    that can't be forged by the client.
    """
    x_forwarded = request.META.get('HTTP_X_FORWARDED_FOR')
    if x_forwarded:
        parts = [p.strip() for p in x_forwarded.split(',') if p.strip()]
        if parts:
            return parts[-1]
    return request.META.get('REMOTE_ADDR') or '0.0.0.0'
