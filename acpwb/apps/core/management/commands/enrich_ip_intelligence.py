"""
Enrich IPIntelligence rows with MaxMind GeoLite2 City/ASN data plus the
hosting and Tor-exit heuristics. Purely local mmdb lookups — no network
calls, no rate limiting needed, so this can drain an arbitrarily large
backlog; it's still time-boxed for cron-friendliness.

Usage:
    python manage.py enrich_ip_intelligence
    python manage.py enrich_ip_intelligence --limit 100
    python manage.py enrich_ip_intelligence --reprocess-stale
    python manage.py enrich_ip_intelligence --reprocess-all
"""
import datetime
import fcntl
import time

import geoip2.database
import geoip2.errors
from django.conf import settings
from django.core.management.base import BaseCommand
from django.db.models import Q
from django.utils import timezone

from apps.core.ip_intel_classify import classify_hosting
from apps.core.tor_exit_list import is_tor_exit
from apps.honeypot.models import IPIntelligence

_LOCK_FILE = '/tmp/acpwb-ip-intel-enrich.lock'


class Command(BaseCommand):
    help = 'Enrich pending IPIntelligence rows with MaxMind GeoLite2 City/ASN + heuristics.'

    def add_arguments(self, parser):
        parser.add_argument('--batch-size', type=int, default=5000)
        parser.add_argument('--max-seconds', type=int, default=0, help='0 = unlimited (default)')
        parser.add_argument('--limit', type=int, default=0, help='Cap total rows processed, 0 = unlimited')
        parser.add_argument('--reprocess-stale', action='store_true', help='Also re-enrich rows enriched against an older mmdb build')
        parser.add_argument('--reprocess-all', action='store_true', help='Re-enrich every row, including already-enriched ones')
        parser.add_argument('--city-db', default=None)
        parser.add_argument('--asn-db', default=None)

    def handle(self, *args, **options):
        with open(_LOCK_FILE, 'w') as lock_fh:
            try:
                fcntl.flock(lock_fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError:
                self.stdout.write('Another enrichment run is already in progress — exiting.')
                return
            self._run(options)

    def _run(self, options):
        city_db_path = options['city_db'] or settings.GEOIP2_CITY_DB_PATH
        asn_db_path = options['asn_db'] or settings.GEOIP2_ASN_DB_PATH

        city_reader = geoip2.database.Reader(city_db_path)
        asn_reader = geoip2.database.Reader(asn_db_path)
        try:
            db_build_date = datetime.date.fromtimestamp(city_reader.metadata().build_epoch)

            qs = IPIntelligence.objects.all()
            if options['reprocess_all']:
                pass
            elif options['reprocess_stale']:
                qs = qs.filter(Q(enriched_at__isnull=True) | Q(geoip_db_date__lt=db_build_date))
            else:
                qs = qs.filter(enriched_at__isnull=True)

            if options['limit']:
                qs = qs.order_by('id')[:options['limit']]
            else:
                qs = qs.order_by('id')

            batch_size = options['batch_size']
            max_seconds = options['max_seconds']
            started = time.monotonic()
            processed = 0
            batch = []

            for obj in qs.iterator(chunk_size=batch_size):
                self._enrich_one(obj, city_reader, asn_reader, db_build_date)
                batch.append(obj)
                processed += 1

                if len(batch) >= batch_size:
                    self._flush(batch)
                    batch = []
                    self.stdout.write(f'  ...{processed:,} processed')

                if max_seconds and time.monotonic() - started >= max_seconds:
                    break

            if batch:
                self._flush(batch)

            self.stdout.write(self.style.SUCCESS(f'Enriched {processed:,} IP(s) (mmdb build {db_build_date}).'))
        finally:
            city_reader.close()
            asn_reader.close()

    def _enrich_one(self, obj, city_reader, asn_reader, db_build_date):
        obj.enriched_at = timezone.now()
        obj.geoip_db_date = db_build_date

        try:
            city_resp = city_reader.city(obj.ip_address)
            obj.country_code = city_resp.country.iso_code or ''
            obj.country_name = city_resp.country.name or ''
            obj.region_name = city_resp.subdivisions.most_specific.name or ''
            obj.city_name = city_resp.city.name or ''
            obj.latitude = city_resp.location.latitude
            obj.longitude = city_resp.location.longitude
            obj.accuracy_radius_km = city_resp.location.accuracy_radius
            city_ok = True
        except geoip2.errors.AddressNotFoundError:
            city_ok = False

        try:
            asn_resp = asn_reader.asn(obj.ip_address)
            obj.asn = asn_resp.autonomous_system_number
            obj.asn_org = asn_resp.autonomous_system_organization or ''
            asn_ok = True
        except geoip2.errors.AddressNotFoundError:
            asn_ok = False

        obj.lookup_ok = city_ok or asn_ok
        obj.enrichment_note = '' if obj.lookup_ok else 'address not found (private/reserved/unmapped)'
        obj.is_hosting = classify_hosting(obj.asn_org)
        obj.is_tor_exit = is_tor_exit(obj.ip_address)

    def _flush(self, batch):
        IPIntelligence.objects.bulk_update(batch, [
            'country_code', 'country_name', 'region_name', 'city_name',
            'latitude', 'longitude', 'accuracy_radius_km',
            'asn', 'asn_org', 'is_hosting', 'is_tor_exit',
            'lookup_ok', 'enrichment_note', 'enriched_at', 'geoip_db_date',
        ])
