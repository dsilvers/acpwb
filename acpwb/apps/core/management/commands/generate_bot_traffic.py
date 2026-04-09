"""
Generate synthetic bot traffic for local testing of the botseed frontend.

Publishes fake request events directly to Redis 'request_stream', bypassing HTTP.
Simulates a realistic mix of bot types, paths, and IPs.

Usage:
    python manage.py generate_bot_traffic             # 5 req/s until Ctrl-C
    python manage.py generate_bot_traffic --rps 20    # faster
    python manage.py generate_bot_traffic --rps 1 --count 50  # fixed count
"""

import json
import random
import signal
import sys
import time
from datetime import datetime, timezone

import redis as redis_lib
from django.core.management.base import BaseCommand

from apps.core.bot_classify import bot_type_to_group, classify_ua_or_ip

_SAMPLE_UAS = [
    # AI crawlers
    'Mozilla/5.0 (compatible; GPTBot/1.0; +https://openai.com/gptbot)',
    'Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; ClaudeBot/1.0; +https://www.anthropic.com/claude-bot)',
    'PerplexityBot/1.0 (+https://perplexity.ai/perplexitybot)',
    'Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html) Google-Extended',
    'Bytespider; spider-feedback@bytedance.com',
    'Mozilla/5.0 (compatible; AmazonBot/1.0; +https://developer.amazon.com/support/amazonbot)',
    'CCBot/2.0 (https://commoncrawl.org/faq/)',
    'cohere-ai/1.0',
    # Search engines
    'Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)',
    'Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)',
    'Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)',
    'Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)',
    'DuckDuckBot/1.1; (+http://duckduckgo.com/duckduckbot.html)',
    'Applebot/0.1 (+http://www.apple.com/go/applebot)',
    # SEO/marketing
    'Mozilla/5.0 (compatible; SemrushBot/7~bl; +http://www.semrush.com/bot.html)',
    'Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)',
    'Mozilla/5.0 (compatible; MJ12bot/v1.4.8; http://mj12bot.com/)',
    'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot)',
    # Generic HTTP clients
    'python-requests/2.31.0',
    'curl/7.88.1',
    'Go-http-client/1.1',
    'axios/1.4.0',
    'python-httpx/0.24.0',
    'aiohttp/3.9.1',
    # Browser-like (may be Alibaba Qwen or real browsers)
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0',
    '',  # empty UA
]

_SAMPLE_IPS = [
    # Alibaba Qwen ranges
    '47.79.100.50', '47.79.201.33', '47.79.18.7',
    '8.219.45.12', '8.219.88.99', '8.219.176.204',
    '47.82.61.5', '47.82.60.200',
    # Googlebot / Bingbot typical ranges
    '66.249.66.1', '66.249.70.15', '66.249.79.200',
    '157.55.39.50', '157.55.39.100',
    # Generic bot hosting
    '52.167.144.100', '13.66.8.77', '40.77.189.220',
    '5.188.210.45', '194.165.16.78', '89.248.165.50',
    # Generic
    '192.0.2.1', '203.0.113.45', '198.51.100.22', '198.51.100.88',
]

_SAMPLE_PATHS = [
    '/', '/our-people/', '/mission/', '/projects/', '/partners/',
    '/reports/', '/reports/2023-compensation-survey/',
    '/wiki/corporate-governance/', '/wiki/esg-framework/', '/wiki/sec-compliance/',
    '/archive/2018/03/15/quarterly-review/', '/archive/2015/07/22/annual-outlook/',
    '/careers/', '/faq/', '/privacy/', '/site-map/',
    '/.env', '/wp-login.php', '/xmlrpc.php', '/wp-config.php',
    '/.well-known/ai-agent.json', '/.well-known/robots.txt',
    '/api/v1/private-data', '/api/v1/openapi.json',
    '/sitemap-wiki.xml', '/sitemap-publications.xml', '/sitemap-archive.xml',
    '/internal/', '/internal/salary-database/', '/internal/employee-records/',
    '/datasets/', '/datasets/employee-compensation/data.jsonl',
    '/feeds/reports.xml', '/feeds/archive.xml',
    '/nonexistent-page/', '/admin/', '/.git/config', '/.htpasswd',
]

_SAMPLE_HOSTS = [
    'acpwb.com', 'acpwb.com', 'acpwb.com', 'acpwb.com',  # weighted toward main domain
    'archives-2018.acpwb.com', 'archives-2015.acpwb.com',
    'archives-2020.acpwb.com', 'archives-2007.acpwb.com',
    'archives-2022.acpwb.com',
]


class Command(BaseCommand):
    help = 'Generate synthetic bot traffic for local botseed testing.'

    def add_arguments(self, parser):
        parser.add_argument(
            '--rps', type=float, default=5.0,
            help='Requests per second (default: 5)',
        )
        parser.add_argument(
            '--count', type=int, default=0,
            help='Number of requests to send, then stop (default: 0 = unlimited)',
        )

    def handle(self, *args, **options):
        import os
        rps = options['rps']
        count = options['count']
        interval = 1.0 / rps

        redis_url = os.environ.get('REDIS_URL', 'redis://redis:6379/0')
        r = redis_lib.from_url(redis_url, decode_responses=True)

        self.stdout.write(
            self.style.SUCCESS(f'Generating traffic at {rps} req/s → {redis_url}')
        )
        if count:
            self.stdout.write(f'Sending {count} requests then stopping.')
        else:
            self.stdout.write('Running indefinitely. Press Ctrl-C to stop.')

        shutdown = [False]

        def handle_signal(signum, frame):
            self.stdout.write('\nStopped.')
            shutdown[0] = True
            sys.exit(0)

        signal.signal(signal.SIGINT, handle_signal)
        signal.signal(signal.SIGTERM, handle_signal)

        sent = 0
        while not shutdown[0]:
            if count and sent >= count:
                break

            ua = random.choice(_SAMPLE_UAS)
            ip_raw = random.choice(_SAMPLE_IPS)
            parts = ip_raw.split('.')
            ip_censored = '.'.join(parts[:-1] + ['xxx']) if len(parts) == 4 else ip_raw

            host = random.choice(_SAMPLE_HOSTS)
            path = random.choice(_SAMPLE_PATHS)
            method = random.choices(['GET', 'GET', 'GET', 'POST'], weights=[7, 7, 7, 1])[0]
            status = random.choices([200, 200, 301, 302, 404, 403], weights=[6, 6, 1, 1, 2, 1])[0]

            bot_type = classify_ua_or_ip(ua, ip_raw)
            bot_group = bot_type_to_group(bot_type)

            payload = json.dumps({
                'ip': ip_censored,
                'host': host,
                'path': path,
                'timestamp': datetime.now(timezone.utc).isoformat(),
                'response_ms': random.randint(50, 500),
                'response_bytes': random.randint(2000, 80000),
                'method': method,
                'status': status,
                'user_agent': ua,
                'bot_type': bot_type,
                'bot_group': bot_group,
            })

            r.publish('request_stream', payload)
            sent += 1

            if sent % 100 == 0:
                self.stdout.write(f'  sent {sent:,} requests...')

            time.sleep(interval)

        self.stdout.write(self.style.SUCCESS(f'Done — sent {sent:,} requests.'))
