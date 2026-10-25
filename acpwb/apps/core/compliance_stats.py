"""
Read the /compliance/ dashboard counters that acpwb_go keeps in Redis
(visitqueue.PushComplianceStats): one set of keys per UTC day,
``acpwb:cstat:<YYYYMMDD>:<name>``, kept 45 days.

    pages      hash  page kind, and "kind|client_class" → hits
    bots       hash  bot type → question-page hits
    topics     hash  course-type slug → hits
    exhibits   hash  exhibit kind → question-page hits
    depth      hash  answers so far → question/results hits
    agencies   zset  agency slug → hits
    ips        HLL   distinct IPs
    att:<d>    HLL   distinct (ip, course, attempt) reaching depth d
    fan:<d>    HLL   distinct (ip, course, attempt, answers) at depth d

Everything here is a cheap read of a few small keys; nothing touches the
CrawlerVisit hypertable.
"""
from collections import Counter, defaultdict
from datetime import datetime, timedelta, timezone

MAX_DEPTH = 25
_client = None


def _redis():
    global _client
    if _client is None:
        import redis as redis_lib
        from django.conf import settings
        _client = redis_lib.from_url(
            getattr(settings, 'REDIS_URL', 'redis://redis:6379/0'),
            socket_connect_timeout=1, socket_timeout=3, decode_responses=True,
        )
    return _client


def _days(n):
    today = datetime.now(timezone.utc).date()
    return [today - timedelta(days=i) for i in range(n - 1, -1, -1)]


def key(day, name):
    return f'acpwb:cstat:{day:%Y%m%d}:{name}'


def read(days=7):
    """Aggregate the last ``days`` days of counters. Returns None if Redis
    is unavailable."""
    try:
        r = _redis()
        dlist = _days(days)
        pipe = r.pipeline(transaction=False)
        for d in dlist:
            for name in ('pages', 'bots', 'topics', 'exhibits', 'depth'):
                pipe.hgetall(key(d, name))
            pipe.zrevrange(key(d, 'agencies'), 0, 49, withscores=True)
            pipe.pfcount(key(d, 'ips'))
        pipe.pfcount(*[key(d, 'ips') for d in dlist])
        for depth in range(MAX_DEPTH + 1):
            pipe.pfcount(*[key(d, f'att:{depth}') for d in dlist])
            pipe.pfcount(*[key(d, f'fan:{depth}') for d in dlist])
        res = pipe.execute()
    except Exception:
        return None

    pages, by_class = Counter(), defaultdict(Counter)
    bots, topics, exhibits, depth, agencies = Counter(), Counter(), Counter(), Counter(), Counter()
    daily = []
    i = 0
    for d in dlist:
        p, b, t, e, dp, ag, ips = res[i:i + 7]
        i += 7
        day_total = 0
        for k, v in p.items():
            v = int(v)
            if '|' in k:
                kind, cls = k.split('|', 1)
                by_class[kind][cls] += v
            else:
                pages[k] += v
                day_total += v
        bots.update({k: int(v) for k, v in b.items()})
        topics.update({k: int(v) for k, v in t.items()})
        exhibits.update({k: int(v) for k, v in e.items()})
        depth.update({int(k): int(v) for k, v in dp.items()})
        agencies.update({k: int(v) for k, v in ag})
        daily.append({'date': d.isoformat(), 'count': day_total, 'ips': ips})
    ips_total = res[i]
    i += 1
    fanout = []
    for dd in range(MAX_DEPTH + 1):
        att, fan = res[i], res[i + 1]
        i += 2
        if att:
            fanout.append({'depth': dd, 'question': dd + 1, 'attempts': att, 'paths': fan,
                           'ratio': round(fan / att, 2) if att else 0})
    return {
        'pages': pages, 'by_class': by_class, 'bots': bots, 'topics': topics,
        'exhibits': exhibits, 'depth': depth, 'agencies': agencies,
        'daily': daily, 'ips_total': ips_total, 'fanout': fanout,
    }
