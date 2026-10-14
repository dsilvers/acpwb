from django.contrib.postgres.operations import AddIndexConcurrently
from django.db import migrations, models

# honeypot_sourcecodeemail is ~22M rows / 10 GB in production, so build the
# index without blocking drain_code_email_queue's inserts. CONCURRENTLY can't
# run in a transaction, hence atomic = False.


class Migration(migrations.Migration):
    atomic = False

    dependencies = [
        ('honeypot', '0029_edgeiphour_bare_page_reqs_and_more'),
    ]

    operations = [
        AddIndexConcurrently(
            model_name='sourcecodeemail',
            index=models.Index(fields=['address', 'first_seen'], name='honeypot_sce_addr_seen_idx'),
        ),
    ]
