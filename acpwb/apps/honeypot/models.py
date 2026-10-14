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
        ('source_code', 'Legacy Source Code'),
        ('source_code_download', 'Legacy Source Code Download'),
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
    # acpwb_edge's ID for the HTTP request that produced this row
    # (X-Request-Id) — joins to EdgeRequest.request_id, and through it to
    # the connection and client. Nullable with no default so adding it to
    # the compressed hypertable is a catalog-only change; NULL on rows from
    # before the edge, or written outside it.
    request_id = models.UUIDField(null=True, blank=True, default=None)
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
    # acpwb_edge's ID for the HTTP request that produced this row
    # (X-Request-Id) — joins to EdgeRequest.request_id, and through it to
    # the connection and client. Nullable with no default so adding it to
    # the compressed hypertable is a catalog-only change; NULL on rows from
    # before the edge, or written outside it.
    request_id = models.UUIDField(null=True, blank=True, default=None)
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

    Header *order* is only captured via ja4h, which acpwb_edge computes from
    the raw request bytes before either backend sees them: order survives in
    Django (gunicorn/WSGI preserves wire order in request.META) but not in
    acpwb_go (Go's net/http.Header is an unordered map), and a signal that
    means different things on the two backends that write this table isn't
    worth having.
    """
    timestamp = models.DateTimeField(default=timezone.now, editable=False, db_index=True)
    ip_address = models.GenericIPAddressField(db_index=True)
    host = models.CharField(max_length=253, blank=True)
    user_agent = models.TextField(blank=True)
    referrer_present = models.BooleanField(default=False)
    first_seen_ip = models.BooleanField(default=False, db_index=True)

    # What the client actually negotiated with acpwb_edge — not the same as
    # the plain HTTP/1.1 connection the edge proxies over, so it's forwarded
    # explicitly (X-Client-Protocol / X-TLS-*, nginx-compatible naming).
    # Blank if the header wasn't forwarded (e.g. a request that reached
    # Django/Go directly, bypassing the edge).
    client_protocol = models.CharField(max_length=16, blank=True)
    tls_protocol = models.CharField(max_length=16, blank=True)
    tls_cipher = models.CharField(max_length=64, blank=True)

    # ClientHello / HTTP fingerprints computed by acpwb_edge (X-JA4, X-JA3,
    # X-JA4H, X-H2-Fingerprint). ja3 is the MD5 form; h2_fingerprint is the
    # Akamai-style SETTINGS|WINDOW_UPDATE|PRIORITY|pseudo-header-order
    # string, blank on HTTP/1.x. The connection-level detail (JA4_r, the JA3
    # string, failed handshakes) lives in EdgeConnection.
    ja4 = models.CharField(max_length=40, blank=True, db_index=True)
    ja3 = models.CharField(max_length=32, blank=True)
    ja4h = models.CharField(max_length=64, blank=True)
    h2_fingerprint = models.CharField(max_length=512, blank=True)

    # Sorted, comma-joined subset of a fixed candidate list of headers real
    # browsers send (Accept-Language, Sec-Fetch-*, sec-ch-ua, ...) that were
    # actually present on this request — see apps.core.signal_capture for the
    # candidate list. A UA claiming a modern browser with few/none of these
    # present is the core "scripted client wearing a browser UA" signal.
    browser_headers_present = models.CharField(max_length=512, blank=True)

    # acpwb_edge's request ID (X-Request-Id) — joins to EdgeRequest.
    request_id = models.UUIDField(null=True, blank=True, default=None)
    idempotency_key = models.UUIDField(unique=True, null=True, blank=True, default=None)

    class Meta:
        ordering = ['-timestamp']
        indexes = [
            models.Index(fields=['ip_address', 'timestamp']),
        ]
        verbose_name = 'Request Fingerprint'

    def __str__(self):
        return f"{self.ip_address} @ {self.timestamp:%Y-%m-%d %H:%M} ({self.client_protocol or '?'})"


class EdgeConnection(models.Model):
    """One row per client TCP connection accepted by acpwb_edge (the Go TLS
    front-end that replaced nginx), including connections that never
    produced a request — failed or aborted TLS handshakes, scanners that
    connect and leave. Pushed by the edge to acpwb:edge_conn_queue and
    batch-drained by drain_edge_conn_queue, same pattern as
    RequestFingerprint.

    Per-request fingerprints already reach RequestFingerprint via forwarded
    headers; this table is the connection-level view (raw JA4_r / JA3
    string, handshake outcome, timing, byte counts) plus a summary of the
    connection's first request, since most bot connections carry only one.

    A TimescaleDB hypertable (see migration 0022): no primary key, and the
    idempotency constraint includes the partitioning column, as TimescaleDB
    requires. Joined to everything else by ip_address value, never by FK.
    """
    OUTCOME_CHOICES = [
        ('ok', 'Handshake completed'),
        ('handshake_failed', 'Handshake failed'),
        ('no_hello', 'No ClientHello'),
        ('timeout', 'Timed out'),
        ('plain', 'Plain HTTP (port 80)'),
    ]

    timestamp = models.DateTimeField(default=timezone.now, editable=False, db_index=True)
    ip_address = models.GenericIPAddressField()
    # BigInteger: ASNs are 32-bit unsigned. No Positive*: its CHECK constraint
    # would scan the whole hypertable when the column is added.
    asn = models.BigIntegerField(null=True, blank=True)  # from the edge's GeoLite2-ASN lookup
    network = models.CharField(max_length=12, blank=True)  # hosting, residential, unknown
    sni = models.CharField(max_length=253, blank=True)
    alpn = models.CharField(max_length=16, blank=True)
    tls_version = models.CharField(max_length=16, blank=True)
    tls_cipher = models.CharField(max_length=64, blank=True)
    outcome = models.CharField(max_length=20, choices=OUTCOME_CHOICES)

    ja4 = models.CharField(max_length=40, blank=True)
    ja4_r = models.TextField(blank=True)
    ja3 = models.CharField(max_length=32, blank=True)
    ja3_str = models.TextField(blank=True)
    h2_fingerprint = models.CharField(max_length=512, blank=True)

    rtt_us = models.BigIntegerField(default=0)
    handshake_us = models.BigIntegerField(default=0)
    duration_ms = models.BigIntegerField(default=0)
    bytes_in = models.BigIntegerField(default=0)
    bytes_out = models.BigIntegerField(default=0)
    request_count = models.IntegerField(default=0)
    hdr_capture_lost = models.BooleanField(default=False)

    # How the connection ended: client_closed / client_reset / read_timeout
    # (handshake, header, idle or body stall) / write_timeout (client
    # stopped reading) / read_error / write_error / server_closed.
    close_reason = models.CharField(max_length=32, blank=True, db_index=True)
    # Requests that reached the edge's handler (request_count), of those
    # the ones answered 503 because the process-wide overload valve was
    # shut (rejected_requests — the edge never limits individual clients),
    # and requests read that no handler served (unserved_requests): HTTP/1
    # requests the server rejected as malformed/oversized, HTTP/2 streams
    # refused or reset before a handler. Each served request, and the first
    # 20 unserved ones per connection, has an EdgeRequest row.
    rejected_requests = models.IntegerField(default=0)
    unserved_requests = models.IntegerField(default=0)

    # Joins to EdgeRequest.conn_id. NULL on aggregated rows.
    conn_id = models.UUIDField(null=True, blank=True, default=None)
    # The first acpwb_cid client ID seen on the connection, if any.
    client_id = models.CharField(max_length=22, blank=True)

    # HTTP/2 client behaviour — attack signatures: Resets ≈ Streams is rapid
    # reset (CVE-2023-44487); many continuations is a CONTINUATION flood;
    # h2_goaway is the error code the server ended the connection with
    # (ENHANCE_YOUR_CALM, PROTOCOL_ERROR, FRAME_SIZE_ERROR, NO_ERROR…).
    h2_streams = models.IntegerField(default=0)
    h2_resets = models.IntegerField(default=0)
    h2_continuations = models.IntegerField(default=0)
    h2_goaway = models.CharField(max_length=32, blank=True)

    # Connections this row stands for. Connections that carried no request
    # from an IP past the edge's per-minute record budget
    # (EDGE_RECORDS_PER_IP_MINUTE) get one aggregated row per (IP, outcome,
    # JA4) per minute: exact conn_count and summed bytes/h2 counters, the
    # other fields from the first of them. Connections with requests are
    # never aggregated. Count connections with Sum('conn_count'), never
    # Count('id').
    conn_count = models.IntegerField(default=1)
    aggregated = models.BooleanField(default=False)

    first_host = models.CharField(max_length=253, blank=True)
    first_method = models.CharField(max_length=16, blank=True)
    first_path = models.CharField(max_length=512, blank=True)
    first_user_agent = models.TextField(blank=True)
    first_ja4h = models.CharField(max_length=64, blank=True)
    first_header_names = models.TextField(blank=True)
    first_status = models.IntegerField(default=0)
    first_route = models.CharField(max_length=32, blank=True)

    idempotency_key = models.UUIDField(null=True, blank=True, default=None)

    class Meta:
        ordering = ['-timestamp']
        indexes = [
            models.Index(fields=['ip_address', 'timestamp']),
            models.Index(fields=['ja4', 'timestamp']),
            models.Index(fields=['conn_id']),
        ]
        constraints = [
            models.UniqueConstraint(
                fields=['idempotency_key', 'timestamp'], name='honeypot_edgeconn_idem_ts_uniq',
            ),
        ]
        verbose_name = 'Edge Connection'

    def __str__(self):
        return f"{self.ip_address} @ {self.timestamp:%Y-%m-%d %H:%M} {self.outcome} {self.ja4 or '-'}"


class EdgeRequest(models.Model):
    """One row per HTTP request seen by acpwb_edge — every request, static
    files and redirects included, and the failed ones: requests the HTTP
    server rejected before any handler ran, HTTP/2 streams never served,
    clients that went away mid-response, upstream errors (see `error`).
    Pushed by the edge to acpwb:edge_request_queue and batch-drained by
    drain_edge_request_queue.

    How it links up (by value — these are hypertables, no FKs):

        IPIntelligence ─ip_address─ EdgeConnection ─conn_id─ EdgeRequest
        EdgeRequest ─request_id─ CrawlerVisit / ArchiveVisit / RequestFingerprint
        EdgeRequest ─client_id─ every request from the same browser (the
            acpwb_cid cookie), across connections and IPs

    A TimescaleDB hypertable (see migration 0023): no primary key;
    request_id + timestamp is the idempotency constraint.
    """
    ERROR_CHOICES = [
        ('', 'Served'),
        ('client_closed', 'Client went away'),
        ('write_failed', 'Response write failed'),
        ('overload', 'Overload valve (503)'),
        ('body_too_large', 'Body too large (413)'),
        ('upstream_error', 'Upstream error (502)'),
        ('upstream_timeout', 'Upstream timeout (504)'),
        ('rejected', 'Rejected by HTTP server'),
        ('incomplete', 'Incomplete request head'),
        ('h2_not_served', 'HTTP/2 stream not served'),
        ('plain_on_tls', 'Plain HTTP on TLS port'),
    ]

    timestamp = models.DateTimeField(default=timezone.now, editable=False, db_index=True)
    request_id = models.UUIDField()
    conn_id = models.UUIDField(null=True, blank=True)
    client_id = models.CharField(max_length=22, blank=True)
    # This response issued the client cookie (first request from a browser,
    # or one that doesn't keep cookies — those get a new ID every time).
    new_client = models.BooleanField(default=False)
    seq = models.IntegerField(default=0)  # position on its connection, from 1

    ip_address = models.GenericIPAddressField()
    tls = models.BooleanField(default=True)
    protocol = models.CharField(max_length=16, blank=True)

    host = models.CharField(max_length=253, blank=True)
    method = models.CharField(max_length=16, blank=True)
    path = models.CharField(max_length=1024, blank=True)
    query = models.CharField(max_length=1024, blank=True)
    user_agent = models.CharField(max_length=512, blank=True)
    referer = models.CharField(max_length=512, blank=True)
    accept_language = models.CharField(max_length=128, blank=True)
    ja4h = models.CharField(max_length=64, blank=True)
    header_names = models.CharField(max_length=2048, blank=True)

    status = models.IntegerField(default=0)  # 0: no response (e.g. h2 stream never served); 499: client closed
    route = models.CharField(max_length=32, blank=True)  # django, acpwb_go, static, redirect, …
    error = models.CharField(max_length=20, choices=ERROR_CHOICES, blank=True)
    content_type = models.CharField(max_length=128, blank=True)
    request_bytes = models.BigIntegerField(default=0)
    response_bytes = models.BigIntegerField(default=0)  # body as sent (compressed)
    ttfb_us = models.BigIntegerField(default=0)
    duration_us = models.BigIntegerField(default=0)

    # What the client really is, classified by the edge per request
    # (acpwb_edge/classify): the User-Agent's claim checked against the TLS
    # stack, ALPN, HTTP/2 framing and headers, plus the network it came from.
    CLIENT_CLASS_CHOICES = [
        ('', 'Not classified'),
        ('known_bot', 'Known bot'),             # names itself, or a known bot IP range
        ('tool', 'HTTP tool'),                  # curl, python-requests…
        ('browser', 'Browser'),                 # claims a browser and is one
        ('impersonator', 'Fake browser'),       # claims a browser its fingerprint contradicts
        ('unknown', 'Unknown / unverified'),
    ]
    ja4 = models.CharField(max_length=40, blank=True)  # the connection's, copied for queries without a join
    asn = models.BigIntegerField(null=True, blank=True)  # see EdgeConnection.asn
    network = models.CharField(max_length=12, blank=True)  # hosting, residential, unknown
    client_class = models.CharField(max_length=16, choices=CLIENT_CLASS_CHOICES, blank=True)
    stack = models.CharField(max_length=48, blank=True)  # the real TLS client: "Go TLS", or the JA4 "a" section
    signals = models.IntegerField(default=0)  # classify.Sig* bitmask; see apps.core.traffic_signals

    class Meta:
        ordering = ['-timestamp']
        indexes = [
            models.Index(fields=['ip_address', 'timestamp']),
            models.Index(fields=['client_id', 'timestamp']),
            models.Index(fields=['conn_id']),
            models.Index(fields=['request_id']),
        ]
        constraints = [
            models.UniqueConstraint(
                fields=['request_id', 'timestamp'], name='honeypot_edgereq_id_ts_uniq',
            ),
        ]
        verbose_name = 'Edge Request'

    def __str__(self):
        return f"{self.ip_address} @ {self.timestamp:%Y-%m-%d %H:%M:%S} {self.method} {self.host}{self.path} {self.status}"


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


class SourceCodeEmail(models.Model):
    """One (address, page) pair: a programmer address shown on an archive
    article's legacy source-code page (served by acpwb_go at
    <article>/source-code/).

    Addresses use the site-wide first.last@acpwb.com format, so the same
    address can appear on many pages. acpwb_go RPUSHes every newly-served
    pair onto acpwb:code_email_queue (deduped per pair for 30 days) and
    drain_code_email_queue folds them in here, so an inbound email to the
    address can be matched back to the page(s) it was harvested from.
    """
    address = models.EmailField(db_index=True)
    host = models.CharField(max_length=253, blank=True)
    path = models.CharField(max_length=512)
    first_seen = models.DateTimeField(default=timezone.now, db_index=True)
    last_seen = models.DateTimeField(default=timezone.now)
    first_ip = models.GenericIPAddressField(null=True, blank=True)
    times_served = models.PositiveIntegerField(default=1)

    class Meta:
        unique_together = [('address', 'path')]
        ordering = ['-first_seen']
        verbose_name = 'Source Code Email'
        verbose_name_plural = 'Source Code Emails'

    def __str__(self):
        return f"{self.address} @ {self.host}{self.path}"
