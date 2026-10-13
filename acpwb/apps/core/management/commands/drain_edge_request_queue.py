"""
Drain the Redis edge-request queue (acpwb:edge_request_queue, pushed by
acpwb_edge once per HTTP request) into the EdgeRequest hypertable.

Same reliable-queue drain as drain_edge_conn_queue. request_id + timestamp
is unique, and a replayed batch carries the same values, so
bulk_create(ignore_conflicts=True) dedups recovery replays.

This is the highest-volume queue (every request, static files included), so
the batch is larger than the other drains'.

Run via cron every minute (see deploy/acpwb-crontab).
"""
from apps.core.crawler_queue import edge_request_queue_length, pop_edge_requests, recover_edge_requests
from apps.core.management.commands.drain_fingerprint_queue import Command as DrainFingerprintCommand
from apps.honeypot.models import EdgeRequest


class Command(DrainFingerprintCommand):
    help = 'Drain the Redis edge-request queue into PostgreSQL via bulk_create'

    model = EdgeRequest
    lock_file = '/tmp/acpwb-edge-request-drain.lock'
    pop = staticmethod(pop_edge_requests)
    recover = staticmethod(recover_edge_requests)
    queue_length = staticmethod(edge_request_queue_length)

    def add_arguments(self, parser):
        super().add_arguments(parser)
        parser.set_defaults(batch=2000)
