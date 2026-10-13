from django.db import migrations

# Hourly chunks for every hypertable, compressed once a chunk is 6 hours
# old. Daily (and, for EdgeConnection, weekly) chunks left up to a week of
# edge data uncompressed — ~95 GB/day at production volume — and data is
# kept forever, so compressing promptly is what keeps the disk bounded.
#
# set_chunk_time_interval only affects chunks created from now on; existing
# chunks keep their ranges and compress under the new policy once they end.
# 6 hours leaves the drains room to catch up after a stall; rows that arrive
# later still insert (TimescaleDB inserts into compressed chunks).

TABLES = [
    'honeypot_crawlervisit',
    'honeypot_archivevisit',
    'honeypot_edgeconnection',
    'honeypot_edgerequest',
]


def _sql(interval, compress_after):
    return '\n'.join(
        f"""
        SELECT set_chunk_time_interval('{t}', INTERVAL '{interval}');
        SELECT remove_compression_policy('{t}', if_exists => true);
        SELECT add_compression_policy(
            '{t}', compress_after => INTERVAL '{compress_after}',
            schedule_interval => INTERVAL '1 hour'
        );
        """
        for t in TABLES
    )


class Migration(migrations.Migration):

    dependencies = [
        ('honeypot', '0023_edgerequest'),
    ]

    operations = [
        migrations.RunSQL(
            sql=_sql('1 hour', '6 hours'),
            reverse_sql=migrations.RunSQL.noop,
        ),
    ]
