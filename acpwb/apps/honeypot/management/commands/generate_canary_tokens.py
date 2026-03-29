import time

import requests
from django.conf import settings
from django.core.management.base import BaseCommand, CommandError

from apps.honeypot.models import CanaryToken


class Command(BaseCommand):
    help = 'Pre-generate canarytokens.org AWS key tokens into the pool'

    def add_arguments(self, parser):
        parser.add_argument(
            '--count',
            type=int,
            default=50,
            help='Number of tokens to generate (default: 50)',
        )
        parser.add_argument(
            '--dry-run',
            action='store_true',
            help='Show what would be generated without calling the API',
        )

    def handle(self, *args, **options):
        count = options['count']
        dry_run = options['dry_run']

        webhook_url = getattr(settings, 'CANARYTOKENS_WEBHOOK_URL', '')
        if not webhook_url:
            raise CommandError(
                'CANARYTOKENS_WEBHOOK_URL is not set in settings. '
                'Add it to your .env file.'
            )

        existing = CanaryToken.objects.filter(token_type='aws_keys', served_at__isnull=True).count()
        self.stdout.write(f'Current unserved pool size: {existing}')

        if dry_run:
            self.stdout.write(
                f'[dry-run] Would generate {count} AWS key token(s) '
                f'via canarytokens.org → {webhook_url}'
            )
            return

        generated = 0
        failed = 0
        for i in range(count):
            memo = f'acpwb-honeypot-{int(time.time())}-{i}'
            try:
                resp = requests.post(
                    'https://canarytokens.org/generate',
                    data={
                        'type': 'aws-keys',
                        'webhook_url': webhook_url,
                        'memo': memo,
                    },
                    timeout=15,
                )
                resp.raise_for_status()
                data = resp.json()
            except requests.RequestException as e:
                self.stderr.write(f'  [{i+1}/{count}] API error: {e}')
                failed += 1
                continue

            token_id = data.get('token', '')
            access_key_id = data.get('access_key_id', '')
            if not token_id or not access_key_id:
                self.stderr.write(f'  [{i+1}/{count}] Unexpected response: {data}')
                failed += 1
                continue

            CanaryToken.objects.create(
                token=token_id,
                token_type='aws_keys',
                canarytoken_token=token_id,
                aws_access_key_id=access_key_id,
                notes=memo,
            )
            generated += 1
            self.stdout.write(f'  [{i+1}/{count}] Created: {access_key_id}')

            # Be polite to canarytokens.org — small delay between requests
            if i < count - 1:
                time.sleep(0.5)

        new_pool = CanaryToken.objects.filter(token_type='aws_keys', served_at__isnull=True).count()
        self.stdout.write(
            self.style.SUCCESS(
                f'Done. Generated: {generated}, Failed: {failed}. '
                f'New pool size: {new_pool}'
            )
        )
