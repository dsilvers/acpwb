"""/compliance/ is served by acpwb_go (tests in acpwb_go/compliance); Django
only drains its milestones and links to it."""
from unittest.mock import patch

import pytest

from apps.honeypot.models import ComplianceCompletion, CrawlerVisit

ITEM = {
    'event': 'completed', 'course': 'sec/insider-trading-awareness-48213', 'attempt': 2,
    'answers': 'abcdabcdabcdabc', 'score': 87, 'passed': True, 'cert_id': '0123abcd',
    'learner_name': 'Ada Lovelace', 'learner_email': 'ada@example.com', 'learner_employee_id': 'E123',
    'ip_address': '10.1.2.3', 'user_agent': 'GPTBot/1.0', 'host': 'acpwb.com',
    'request_id': '0f8fad5b-d9cb-469f-a165-70867728950e', 'timestamp': '2026-10-10T05:00:00+00:00',
}


def test_trap_type_choices():
    choices = dict(CrawlerVisit.TRAP_CHOICES)
    assert 'compliance' in choices and 'compliance_certificate' in choices


def test_build_completion():
    from apps.core.management.commands.drain_compliance_queue import build_completion
    row = build_completion(ITEM)
    assert row.event == 'completed' and row.passed and row.score == 87 and row.attempt == 2
    assert str(row.request_id) == ITEM['request_id']
    assert row.timestamp.isoformat() == '2026-10-10T05:00:00+00:00'
    assert build_completion({**ITEM, 'event': 'bogus'}) is None
    assert build_completion({**ITEM, 'ip_address': ''}) is None
    assert build_completion({**ITEM, 'request_id': 'not-a-uuid'}).request_id is None
    assert build_completion({**ITEM, 'score': 'x'}).score == 0


@pytest.mark.django_db
def test_drain_inserts_rows():
    from apps.core.management.commands.drain_compliance_queue import Command
    batches = iter([([ITEM, {**ITEM, 'event': 'certificate'}, {'event': 'nope'}], 'batch-1'), ([], None)])
    with patch('apps.core.management.commands.drain_compliance_queue.recover_compliance_completions', return_value=[]), \
         patch('apps.core.management.commands.drain_compliance_queue.pop_compliance_completions', side_effect=lambda n: next(batches)), \
         patch('apps.core.management.commands.drain_compliance_queue.finalize_batch') as fin, \
         patch('apps.core.management.commands.drain_compliance_queue.compliance_queue_length', return_value=0):
        Command()._drain({'batch': 500, 'max_seconds': 0})
    assert sorted(ComplianceCompletion.objects.values_list('event', flat=True)) == ['certificate', 'completed']
    fin.assert_called_once_with('batch-1')


def test_robots_lists_compliance(client):
    body = client.get('/robots.txt').content.decode()
    assert 'Allow: /compliance/' in body
    assert 'Sitemap: https://acpwb.com/compliance/sitemap.xml' in body


@pytest.mark.django_db
def test_header_and_footer_link(client):
    html = client.get('/').content.decode()
    assert html.count('href="/compliance/">Compliance</a>') == 2


def test_slashless_redirect(client):
    resp = client.get('/compliance')
    assert resp.status_code == 301 and resp['Location'] == '/compliance/'


def test_build_completion_classification():
    from apps.core.management.commands.drain_compliance_queue import build_completion
    row = build_completion({**ITEM, 'bot_type': 'OpenAI GPTBot', 'client_class': 'known_bot'})
    assert row.bot_type == 'OpenAI GPTBot' and row.client_class == 'known_bot'
    assert build_completion(ITEM).bot_type == ''


class _FakeRedis:
    """Just enough Redis for compliance_stats.read: hashes, sorted sets, and
    HyperLogLogs as exact sets."""
    def __init__(self):
        self.h, self.z, self.s = {}, {}, {}

    def hset(self, k, f=None, v=None, mapping=None):
        self.h.setdefault(k, {}).update(mapping or {f: v})

    def zincrby(self, k, n, m):
        self.z.setdefault(k, {})[m] = self.z.get(k, {}).get(m, 0) + n

    def pfadd(self, k, *vals):
        self.s.setdefault(k, set()).update(vals)

    def pipeline(self, transaction=False):
        fake, calls = self, []

        class P:
            def __getattr__(self, name):
                return lambda *a, **kw: calls.append((name, a, kw))

            def execute(self):
                out = []
                for name, a, kw in calls:
                    if name == 'hgetall':
                        out.append({k: str(v) for k, v in fake.h.get(a[0], {}).items()})
                    elif name == 'zrevrange':
                        out.append(sorted(fake.z.get(a[0], {}).items(), key=lambda x: -x[1]))
                    elif name == 'pfcount':
                        out.append(len(set().union(*[fake.s.get(k, set()) for k in a])))
                return out
        return P()


def test_compliance_stats_aggregates_days():
    """apps.core.compliance_stats folds acpwb_go's per-day Redis counters."""
    from datetime import datetime, timedelta, timezone as tz

    from apps.core import compliance_stats

    r = _FakeRedis()
    today = datetime.now(tz.utc).date()
    for d, n in ((today, 3), (today - timedelta(days=1), 2)):
        r.hset(compliance_stats.key(d, 'pages'), mapping={'question': n, 'question|known_bot': n, 'course': 1})
        r.hset(compliance_stats.key(d, 'topics'), 'cpr-and-aed-awareness', n)
        r.zincrby(compliance_stats.key(d, 'agencies'), n, 'osha')
        r.pfadd(compliance_stats.key(d, 'ips'), f'10.0.0.{n}')
        r.pfadd(compliance_stats.key(d, 'att:0'), f'10.0.0.{n}|osha/x|1')
        r.pfadd(compliance_stats.key(d, 'fan:0'), f'10.0.0.{n}|osha/x|1|')
        r.pfadd(compliance_stats.key(d, 'att:1'), f'10.0.0.{n}|osha/x|1')
        r.pfadd(compliance_stats.key(d, 'fan:1'), *[f'10.0.0.{n}|osha/x|1|{c}' for c in 'abcd'])
    with patch.object(compliance_stats, '_redis', return_value=r):
        s = compliance_stats.read(7)
    assert s['pages']['question'] == 5 and s['by_class']['question']['known_bot'] == 5
    assert s['topics']['cpr-and-aed-awareness'] == 5 and s['agencies']['osha'] == 5
    assert s['ips_total'] == 2
    assert [(f['question'], f['ratio']) for f in s['fanout']] == [(1, 1.0), (2, 4.0)]


@pytest.mark.django_db
def test_compliance_dashboard_renders(admin_client):
    ComplianceCompletion.objects.create(event='completed', ip_address='10.0.0.1', course='osha/cpr-and-aed-awareness-12345',
                                        attempt=1, score=92, passed=True, bot_type='OpenAI GPTBot', client_class='known_bot')
    for i in range(3):
        ComplianceCompletion.objects.create(event='completed', ip_address='10.0.0.1', course='osha/x-1', attempt=i + 2,
                                            score=88, passed=True, bot_type='OpenAI GPTBot', client_class='known_bot')
    ComplianceCompletion.objects.create(event='certificate', ip_address='10.0.0.1', course='osha/x-1', attempt=2,
                                        score=88, passed=True, cert_id='0123abcd', bot_type='OpenAI GPTBot')
    ComplianceCompletion.objects.create(event='identified', ip_address='10.0.0.2', course='osha/x-1',
                                        learner_name='<b>Robo Tester</b>', learner_email='robo@example.com')
    with patch('apps.core.compliance_stats.read', return_value=None):
        resp = admin_client.get('/acpwb-dashboard/compliance/?days=7')
    assert resp.status_code == 200
    body = resp.content.decode()
    assert 'OpenAI GPTBot' in body and 'reads questions' in body
    assert 'robo@example.com' in body and '&lt;b&gt;Robo Tester&lt;/b&gt;' in body
    assert 'Redis unavailable' in body

    from collections import Counter, defaultdict
    stats = {
        'pages': Counter({'course': 4, 'question': 20}), 'by_class': defaultdict(Counter, {'question': Counter({'known_bot': 20})}),
        'bots': Counter({'OpenAI GPTBot': 20}), 'topics': Counter({'cpr-and-aed-awareness': 9}),
        'exhibits': Counter({'aedpads': 2}), 'depth': Counter({0: 4}), 'agencies': Counter({'osha': 9}),
        'daily': [{'date': '2026-10-22', 'count': 24, 'ips': 2}], 'ips_total': 2,
        'fanout': [{'depth': 0, 'question': 1, 'attempts': 2, 'paths': 2, 'ratio': 1.0},
                   {'depth': 1, 'question': 2, 'attempts': 2, 'paths': 8, 'ratio': 4.0}],
    }
    with patch('apps.core.compliance_stats.read', return_value=stats):
        body = admin_client.get('/acpwb-dashboard/compliance/').content.decode()
    assert 'Attempts started (distinct)' in body and 'aedpads' in body and 'Known bot' in body


@pytest.mark.django_db
def test_compliance_dashboard_requires_staff(client):
    assert client.get('/acpwb-dashboard/compliance/').status_code == 302
