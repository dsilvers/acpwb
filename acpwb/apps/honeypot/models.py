from django.db import models
from django.utils import timezone


class CrawlerVisit(models.Model):
    TRAP_CHOICES = [
        ('archive', 'Archive Loop'),
        ('ghost_link', 'Ghost Link'),
        ('well_known', 'Well-Known File'),
        ('api', 'Fake API'),
        ('wiki', 'Wiki Page'),
        ('pow', 'PoW Challenge'),
        ('report_list', 'Report Listing'),
        ('report_download', 'Report Download'),
        ('dataset', 'Training Dataset'),
        ('policy', 'Public Policy Filing'),
        ('scanner_probe', 'Scanner Probe (404)'),
        ('env_probe', 'Config File Probe'),
        ('wp_probe', 'WordPress Probe'),
        ('webshell_probe', 'Webshell Probe'),
        ('canary_trigger', 'Canary Token Triggered'),
        ('handbook', 'Company Handbook'),
        ('process_improvement', 'Process Improvement'),
        ('presentation', 'Presentation'),
        ('other', 'Other'),
    ]

    timestamp = models.DateTimeField(default=timezone.now, editable=False, db_index=True)
    ip_address = models.GenericIPAddressField(db_index=True)
    user_agent = models.TextField(blank=True)
    host = models.CharField(max_length=253, blank=True, db_index=True)
    path = models.TextField()
    referrer = models.TextField(blank=True)
    trap_type = models.CharField(max_length=32, choices=TRAP_CHOICES, default='other', db_index=True)
    query_string = models.TextField(blank=True)
    # Denormalized at write time — enables fast GROUP BY without Python-side UA parsing
    bot_type = models.CharField(max_length=64, blank=True, db_index=True)
    bot_group = models.CharField(max_length=64, blank=True, db_index=True)
    # Minted once when an item is first queued in Redis (crawler_queue.py) so
    # the drain consumer's crash-recovery path can tell whether a recovered
    # batch already committed before re-inserting it. Deliberately NOT a DB
    # unique constraint/index: this table is a TimescaleDB hypertable, which
    # rejects CREATE INDEX CONCURRENTLY outright, and a non-concurrent build
    # would block live writes on this table for the build's duration. The
    # drain command instead checks for existing keys with a plain query
    # (see drain_crawler_queue.py) before re-inserting a recovered batch.
    # NULL on rows predating this queue design and on any row written
    # outside it — no backfill needed or wanted.
    idempotency_key = models.UUIDField(null=True, blank=True, default=None)

    class Meta:
        ordering = ['-timestamp']
        indexes = [
            models.Index(fields=['trap_type', 'timestamp']),
            models.Index(fields=['bot_type', 'timestamp']),
            models.Index(fields=['bot_group', 'timestamp']),
        ]
        verbose_name = 'Crawler Visit'

    def __str__(self):
        return f"{self.ip_address} [{self.trap_type}] {self.path[:60]} @ {self.timestamp:%Y-%m-%d %H:%M}"


class WikiPage(models.Model):
    topic = models.SlugField(max_length=128, unique=True, db_index=True)
    title = models.CharField(max_length=256)
    body_paragraphs = models.JSONField(default=list)
    watermark_token = models.CharField(max_length=16)
    related_topics = models.JSONField(default=list)
    generated_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['topic']
        verbose_name = 'Wiki Page'

    def __str__(self):
        return f"{self.title} [{self.watermark_token}]"


class PublicReport(models.Model):
    FILE_TYPES = [('csv', 'CSV Dataset'), ('pdf', 'PDF Document')]

    slug            = models.SlugField(max_length=128, unique=True, db_index=True)
    title           = models.CharField(max_length=256)
    category        = models.CharField(max_length=64)
    file_type       = models.CharField(max_length=8, choices=FILE_TYPES)
    pub_date        = models.DateField()
    summary         = models.TextField()
    watermark_token = models.CharField(max_length=16)
    page_number     = models.PositiveIntegerField(db_index=True, default=0)
    generated_at    = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['-pub_date', 'slug']
        verbose_name = 'Public Report'

    def __str__(self):
        return f"[{self.file_type.upper()}] {self.title} ({self.watermark_token})"


class ArchiveVisit(models.Model):
    timestamp = models.DateTimeField(default=timezone.now, editable=False, db_index=True)
    ip_address = models.GenericIPAddressField(db_index=True)
    user_agent = models.TextField(blank=True)
    year = models.IntegerField()
    month = models.IntegerField()
    day = models.IntegerField()
    slug = models.CharField(max_length=512)
    depth = models.PositiveIntegerField(default=0, db_index=True)
    # See CrawlerVisit.idempotency_key — same purpose, same deliberate lack
    # of a DB constraint/index (hypertable + no CONCURRENTLY support), same
    # NULL-for-old-rows / no-backfill approach.
    idempotency_key = models.UUIDField(null=True, blank=True, default=None)

    class Meta:
        ordering = ['-timestamp']
        indexes = [
            models.Index(fields=['ip_address', 'timestamp']),
            models.Index(fields=['depth', 'timestamp']),
        ]
        verbose_name = 'Archive Visit'

    def __str__(self):
        return f"{self.ip_address} /archive/{self.year}/{self.month}/{self.day}/{self.slug[:40]}"


class CanaryToken(models.Model):
    """A trackable token embedded in fake credential files.

    Self-hosted callback URL embedded in the fake config file; fires when the
    bot GETs /.well-known/tokens/<token>/ping.
    """
    TOKEN_TYPES = [
        ('env_url',   'Self-hosted .env canary URL'),
        ('wp_config', 'wp-config.php canary URL'),
        ('git_config', '.git/config canary URL'),
    ]
    token = models.CharField(max_length=128, unique=True, db_index=True)
    token_type = models.CharField(max_length=32, choices=TOKEN_TYPES)
    # Lifecycle
    served_to_ip = models.GenericIPAddressField(null=True, blank=True)
    served_at = models.DateTimeField(null=True, blank=True, db_index=True)
    triggered = models.BooleanField(default=False, db_index=True)
    triggered_at = models.DateTimeField(null=True, blank=True)
    triggered_ip = models.GenericIPAddressField(null=True, blank=True)
    triggered_ua = models.TextField(blank=True)
    notes = models.TextField(blank=True)
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['-created_at']
        indexes = [
            models.Index(fields=['token_type', 'served_at']),
            models.Index(fields=['triggered', 'triggered_at']),
        ]
        verbose_name = 'Canary Token'

    def __str__(self):
        status = 'TRIGGERED' if self.triggered else 'unserved' if not self.served_at else 'served'
        return f"[{self.token_type}] {self.token[:16]}... ({status})"


class InternalLoginAttempt(models.Model):
    ip_address = models.GenericIPAddressField(db_index=True)
    user_agent = models.TextField(blank=True)
    username = models.CharField(max_length=255, blank=True)
    password = models.CharField(max_length=255, blank=True)
    next_url = models.CharField(max_length=500, blank=True)
    created_at = models.DateTimeField(auto_now_add=True, db_index=True)

    class Meta:
        ordering = ['-created_at']
        verbose_name = 'Internal Login Attempt'

    def __str__(self):
        return f"{self.ip_address} tried '{self.username}' @ {self.created_at:%Y-%m-%d %H:%M}"


class PathStat(models.Model):
    host = models.CharField(max_length=253, blank=True, db_index=True)
    path = models.TextField()
    count = models.BigIntegerField(default=0)

    class Meta:
        constraints = [
            models.UniqueConstraint(fields=['host', 'path'], name='pathstat_host_path_uniq')
        ]

    def __str__(self):
        prefix = f"{self.host}" if self.host else 'acpwb.com'
        return f"{prefix}{self.path} ({self.count:,})"


class IPIntelligence(models.Model):
    """One row per distinct IP seen in CrawlerVisit, enriched with MaxMind
    GeoLite2 geo/ASN data plus best-effort hosting/Tor heuristics.

    Deliberately NOT a ForeignKey target from CrawlerVisit — that table is a
    90M-rows/day TimescaleDB hypertable, and adding a column there for every
    historical + future row isn't worth it. This is joined to CrawlerVisit by
    the shared ip_address value only (see discover_ip_intelligence, which
    populates first_seen/last_seen/visit_count from CrawlerVisit directly).
    """
    ip_address = models.GenericIPAddressField(unique=True, db_index=True)
    ip_version = models.PositiveSmallIntegerField(default=4)

    # MaxMind GeoLite2-City
    country_code = models.CharField(max_length=2, blank=True)  # indexed via (country_code, is_hosting)
    country_name = models.CharField(max_length=128, blank=True)
    region_name = models.CharField(max_length=128, blank=True)
    city_name = models.CharField(max_length=128, blank=True)
    latitude = models.FloatField(null=True, blank=True)
    longitude = models.FloatField(null=True, blank=True)
    accuracy_radius_km = models.PositiveIntegerField(null=True, blank=True)

    # MaxMind GeoLite2-ASN — asn_org doubles as the "ISP" field
    asn = models.PositiveIntegerField(null=True, blank=True)
    asn_org = models.CharField(max_length=256, blank=True)

    # Best-effort heuristics — not authoritative, see apps.core.ip_intel_classify
    # and apps.core.tor_exit_list for how these are derived.
    is_hosting = models.BooleanField(default=False)
    is_tor_exit = models.BooleanField(default=False)

    # Enrichment bookkeeping
    lookup_ok = models.BooleanField(default=False)
    enrichment_note = models.CharField(max_length=255, blank=True)
    enriched_at = models.DateTimeField(null=True, blank=True, db_index=True)
    geoip_db_date = models.DateField(null=True, blank=True)

    # Populated by discover_ip_intelligence as a side effect of the GROUP BY
    # it already has to do — avoids a second pass over CrawlerVisit.
    # Deliberately NOT indexed: discover_ip_intelligence rewrites these (and
    # visit_count) on every IP it sees, every run. With no index on any of
    # them, those updates are HOT (heap-only) and touch no index at all —
    # an index here made every one rewrite all of the table's indexes
    # instead (see migration 0020).
    first_seen = models.DateTimeField(null=True, blank=True)
    last_seen = models.DateTimeField(null=True, blank=True)
    visit_count = models.BigIntegerField(default=0)

    class Meta:
        # Keep this list short — every index is written on every
        # enrich_ip_intelligence update. Lookups that matter: ip_address
        # (unique, discover's upsert), enriched_at (enrich's backlog), and
        # country/hosting breakdowns. Reports otherwise seq-scan, which is
        # fine for an occasional command.
        indexes = [
            models.Index(fields=['country_code', 'is_hosting']),
        ]
        verbose_name = 'IP Intelligence'
        verbose_name_plural = 'IP Intelligence'

    def __str__(self):
        return f"{self.ip_address} [{self.country_code or '??'}] {self.asn_org or 'unknown org'}"


class RequestFingerprint(models.Model):
    """Raw per-request signal for residential-proxy detection, queued through
    Redis and batch-drained exactly like CrawlerVisit/ArchiveVisit (see
    drain_fingerprint_queue) — never a synchronous write on the request path.

    Unlike CrawlerVisit, this is a brand-new table (not an existing 373M-row
    hypertable), so idempotency_key gets a real unique constraint from day
    one instead of CrawlerVisit's app-level dedup workaround.

    Deliberately captures header *presence*, not header *order*: order
    survives reliably in Django (gunicorn/WSGI preserves wire order in
    request.META) but not in acpwb_go (Go's net/http.Header is an unordered
    map), and a signal that means different things on the two backends that
    write this table isn't worth having.
    """
    timestamp = models.DateTimeField(default=timezone.now, editable=False, db_index=True)
    ip_address = models.GenericIPAddressField(db_index=True)
    host = models.CharField(max_length=253, blank=True)
    user_agent = models.TextField(blank=True)
    referrer_present = models.BooleanField(default=False)
    first_seen_ip = models.BooleanField(default=False, db_index=True)

    # What the client actually negotiated with nginx — not the same as the
    # HTTP/1.1 connection nginx always re-encodes to for the upstream, so
    # this has to be forwarded explicitly (see nginx's $server_protocol ->
    # X-Client-Protocol). Blank if the header wasn't forwarded (e.g. a
    # request that reached Django/Go directly, bypassing nginx).
    client_protocol = models.CharField(max_length=16, blank=True)
    tls_protocol = models.CharField(max_length=16, blank=True)
    tls_cipher = models.CharField(max_length=64, blank=True)

    # Sorted, comma-joined subset of a fixed candidate list of headers real
    # browsers send (Accept-Language, Sec-Fetch-*, sec-ch-ua, ...) that were
    # actually present on this request — see apps.core.signal_capture for the
    # candidate list. A UA claiming a modern browser with few/none of these
    # present is the core "scripted client wearing a browser UA" signal.
    browser_headers_present = models.CharField(max_length=512, blank=True)

    idempotency_key = models.UUIDField(unique=True, null=True, blank=True, default=None)

    class Meta:
        ordering = ['-timestamp']
        indexes = [
            models.Index(fields=['ip_address', 'timestamp']),
        ]
        verbose_name = 'Request Fingerprint'

    def __str__(self):
        return f"{self.ip_address} @ {self.timestamp:%Y-%m-%d %H:%M} ({self.client_protocol or '?'})"


class IPReputationScore(models.Model):
    """Rule-based residential-proxy / automation score for an IP, computed
    off the request path by apps.core.management.commands.score_ip_reputation
    from RequestFingerprint + CrawlerVisit + IPIntelligence.

    Kept separate from IPIntelligence (rather than adding columns there) so
    the scoring model can be revised/rebuilt independently of the
    geo/ASN/Tor enrichment it already does — score_version exists precisely
    so a rescoring pass never silently changes the meaning of old rows in
    place.
    """
    CLASSIFICATIONS = [
        ('human', 'Likely Human'),
        ('datacenter_bot', 'Datacenter Bot'),
        ('residential_proxy_bot', 'Residential Proxy Bot'),
        ('unknown', 'Unknown'),
    ]

    ip_address = models.GenericIPAddressField(unique=True, db_index=True)
    residential_proxy_score = models.PositiveSmallIntegerField(default=0, db_index=True)
    classification = models.CharField(max_length=32, choices=CLASSIFICATIONS, default='unknown', db_index=True)
    # Which signals fired, e.g. ["trap_tainted", "header_mismatch"] — internal
    # evidence, never exposed via PublishedIPReputation.
    evidence = models.JSONField(default=list)
    score_version = models.CharField(max_length=16, default='v1')

    first_flagged_at = models.DateTimeField(null=True, blank=True)
    last_scored_at = models.DateTimeField(null=True, blank=True, db_index=True)

    class Meta:
        indexes = [
            models.Index(fields=['classification', 'residential_proxy_score']),
        ]
        verbose_name = 'IP Reputation Score'
        verbose_name_plural = 'IP Reputation Scores'

    def __str__(self):
        return f"{self.ip_address} [{self.classification}] score={self.residential_proxy_score}"


class PublishedIPReputation(models.Model):
    """The external-facing surface of the traffic-intelligence product —
    deliberately decoupled from IPReputationScore's internal evidence.

    This table (not IPReputationScore directly) is what any future
    blacklist/API export reads from: it exposes classification + confidence
    only, never which of our honeypot paths or internal signals triggered
    it. Keeping the two separate means the detection logic in
    score_ip_reputation can change freely without redefining an external
    contract someone else may depend on.
    """
    ip_address = models.GenericIPAddressField(unique=True, db_index=True)
    classification = models.CharField(max_length=32, choices=IPReputationScore.CLASSIFICATIONS, default='unknown')
    confidence = models.PositiveSmallIntegerField(default=0)
    published_at = models.DateTimeField(auto_now=True, db_index=True)

    class Meta:
        verbose_name = 'Published IP Reputation'
        verbose_name_plural = 'Published IP Reputation'

    def __str__(self):
        return f"{self.ip_address} [{self.classification}] confidence={self.confidence}"
