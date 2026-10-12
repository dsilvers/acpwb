"""
Drain the Redis edge-connection queue (acpwb:edge_conn_queue, pushed by
acpwb_edge once per client connection) into the EdgeConnection hypertable.

Same reliable-queue drain as drain_fingerprint_queue — only the queue,
model and lock file differ. idempotency_key + timestamp is unique (the
hypertable requires the partition column in every unique constraint), and a
replayed batch carries the same timestamps, so bulk_create(ignore_conflicts=
True) still dedups recovery replays.

Run via cron every minute (see deploy/acpwb-crontab).
"""
from apps.core.crawler_queue import edge_conn_queue_length, pop_edge_connections, recover_edge_connections
from apps.core.management.commands.drain_fingerprint_queue import Command as DrainFingerprintCommand
from apps.honeypot.models import EdgeConnection


class Command(DrainFingerprintCommand):
    help = 'Drain the Redis edge-connection queue into PostgreSQL via bulk_create'

    model = EdgeConnection
    lock_file = '/tmp/acpwb-edge-conn-drain.lock'
    pop = staticmethod(pop_edge_connections)
    recover = staticmethod(recover_edge_connections)
    queue_length = staticmethod(edge_conn_queue_length)
