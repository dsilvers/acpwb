"""
Refresh the local Tor exit-node list from the public Tor Project bulk list.

The list changes daily, so this is scheduled once a day via cron:
    17 3 * * * cd /home/acpwb/acpwb && bash -c 'set -a; source .env; set +a; \
        exec /home/acpwb/acpwb/.direnv/python-3.14/bin/python \
        /home/acpwb/acpwb/acpwb/manage.py refresh_tor_exit_list' \
        >> /var/log/acpwb-tor-refresh.log 2>&1

Writes atomically (temp file + os.replace) so apps.core.tor_exit_list never
sees a half-written file mid-refresh.
"""
import fcntl
import os
import urllib.request

from django.conf import settings
from django.core.management.base import BaseCommand

_LOCK_FILE = '/tmp/acpwb-tor-refresh.lock'


class Command(BaseCommand):
    help = 'Download the current Tor bulk exit-node list to TOR_EXIT_LIST_PATH.'

    def add_arguments(self, parser):
        parser.add_argument('--url', default=None, help='Override TOR_EXIT_LIST_URL')
        parser.add_argument('--dry-run', action='store_true', help='Fetch and count lines, do not write')

    def handle(self, *args, **options):
        with open(_LOCK_FILE, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self.stdout.write('Another refresh is already running — exiting.')
                return
            self._refresh(options)

    def _refresh(self, options):
        url = options['url'] or settings.TOR_EXIT_LIST_URL
        req = urllib.request.Request(url, headers={'User-Agent': 'acpwb.com-tor-refresh/1.0'})
        with urllib.request.urlopen(req, timeout=30) as resp:
            body = resp.read().decode('utf-8', errors='replace')

        ips = [line.strip() for line in body.splitlines() if line.strip() and not line.startswith('#')]

        if options['dry_run']:
            self.stdout.write(f'Fetched {len(ips)} exit-node IPs from {url} (dry run, not written).')
            return

        path = settings.TOR_EXIT_LIST_PATH
        os.makedirs(os.path.dirname(path), exist_ok=True)
        tmp_path = path + '.tmp'
        with open(tmp_path, 'w') as f:
            f.write('\n'.join(ips) + '\n')
        os.replace(tmp_path, path)

        self.stdout.write(self.style.SUCCESS(f'Wrote {len(ips)} exit-node IPs to {path}.'))
