"""
Download the MaxMind GeoLite2-City and GeoLite2-ASN databases used by
enrich_ip_intelligence, using the account's official permalinks with HTTP
Basic Auth (MAXMIND_ACCOUNT_ID / MAXMIND_LICENSE_KEY).

Verifies the accompanying .sha256 checksum before replacing the live .mmdb
file, and writes atomically (temp file + os.replace) so a failed/partial
download never leaves enrich_ip_intelligence reading a corrupt database.

Run manually once to provision a box, then monthly via cron (MaxMind
publishes GeoLite2 updates roughly monthly):
    0 4 1 * * cd /home/acpwb/acpwb && bash -c 'set -a; source .env; set +a; \
        exec /home/acpwb/acpwb/.direnv/python-3.14/bin/python \
        /home/acpwb/acpwb/acpwb/manage.py download_geoip_db' \
        >> /var/log/acpwb-geoip-download.log 2>&1
"""
import base64
import hashlib
import io
import os
import tarfile
import urllib.request

from django.conf import settings
from django.core.management.base import BaseCommand, CommandError

_EDITIONS = [
    ('GeoLite2-City', 'GEOIP2_CITY_DB_PATH'),
    ('GeoLite2-ASN', 'GEOIP2_ASN_DB_PATH'),
]
_DOWNLOAD_URL = 'https://download.maxmind.com/geoip/databases/{edition}/download?suffix=tar.gz'
_SHA256_URL = 'https://download.maxmind.com/geoip/databases/{edition}/download?suffix=tar.gz.sha256'


class Command(BaseCommand):
    help = 'Download and verify the GeoLite2-City and GeoLite2-ASN .mmdb databases from MaxMind.'

    def handle(self, *args, **options):
        account_id = settings.MAXMIND_ACCOUNT_ID
        license_key = settings.MAXMIND_LICENSE_KEY
        if not account_id or not license_key:
            raise CommandError('MAXMIND_ACCOUNT_ID / MAXMIND_LICENSE_KEY are not set in .env')

        auth_header = 'Basic ' + base64.b64encode(f'{account_id}:{license_key}'.encode()).decode()

        for edition, setting_name in _EDITIONS:
            dest_path = getattr(settings, setting_name)
            self.stdout.write(f'Fetching {edition} -> {dest_path}')
            self._fetch_one(edition, dest_path, auth_header)
            self.stdout.write(self.style.SUCCESS(f'  OK ({edition})'))

    def _fetch_one(self, edition, dest_path, auth_header):
        tar_bytes = self._get(_DOWNLOAD_URL.format(edition=edition), auth_header)
        sha256_body = self._get(_SHA256_URL.format(edition=edition), auth_header).decode('utf-8')
        expected_sha256 = sha256_body.strip().split()[0]

        actual_sha256 = hashlib.sha256(tar_bytes).hexdigest()
        if actual_sha256 != expected_sha256:
            raise CommandError(
                f'{edition}: sha256 mismatch (expected {expected_sha256}, got {actual_sha256}) — aborting'
            )

        with tarfile.open(fileobj=io.BytesIO(tar_bytes), mode='r:gz') as tar:
            mmdb_member = next((m for m in tar.getmembers() if m.name.endswith('.mmdb')), None)
            if mmdb_member is None:
                raise CommandError(f'{edition}: no .mmdb file found inside downloaded archive')
            mmdb_bytes = tar.extractfile(mmdb_member).read()

        os.makedirs(os.path.dirname(dest_path), exist_ok=True)
        tmp_path = dest_path + '.tmp'
        with open(tmp_path, 'wb') as f:
            f.write(mmdb_bytes)
        os.replace(tmp_path, dest_path)

    def _get(self, url, auth_header):
        # MaxMind's download endpoint 302s to a presigned, short-lived S3/R2
        # URL. That redirect target rejects the request if our Authorization
        # header for download.maxmind.com is still attached (a plain
        # urlopen() forwards it by default) — strip it on redirect, same as
        # curl's default cross-host behavior.
        opener = urllib.request.build_opener(_DropAuthOnRedirect)
        req = urllib.request.Request(url, headers={
            'Authorization': auth_header,
            'User-Agent': 'acpwb.com-geoip-download/1.0',
        })
        try:
            with opener.open(req, timeout=60) as resp:
                return resp.read()
        except urllib.error.HTTPError as exc:
            raise CommandError(f'HTTP {exc.code} fetching {url}: {exc.reason}') from exc


class _DropAuthOnRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        new_req = super().redirect_request(req, fp, code, msg, headers, newurl)
        if new_req is not None:
            new_req.remove_header('Authorization')
        return new_req
