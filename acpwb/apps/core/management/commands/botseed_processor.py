"""
Long-running management command that taps into ACPWB's request_stream Redis
pub/sub channel and produces entropy-seeded random numbers for botseed.net.

Each incoming request event is mixed with hardware entropy (secrets.token_bytes)
via SHA-256 to produce a seeded random integer. Results are published to
'botseed_stream' at ≤20 events/sec and cached in Redis at 'botseed:latest'.

Usage:
    python manage.py botseed_processor
"""

import hashlib
import json
import logging
import os
import random
import secrets
import signal
import sys
import threading
import time
from collections import deque
from datetime import datetime, timezone

import redis
from django.core.management.base import BaseCommand

logger = logging.getLogger(__name__)

REDIS_CHANNEL_IN = 'request_stream'
REDIS_CHANNEL_OUT = 'botseed_stream'
REDIS_KEY_LATEST = 'botseed:latest'
MAX_EVENTS_PER_SEC = 20
MIN_INTERVAL = 1.0 / MAX_EVENTS_PER_SEC  # 0.05 seconds
MAX_PENDING = 60  # 3 seconds of backlog at 20 eps


def generate_chaos_number(log_json_str):
    """
    Mix hardware entropy with request-log entropy via SHA-256 → seeded PRNG.

    Uses random.Random(seed) — an isolated instance — so this never touches
    the global random state used by Django and other application code.

    Returns (system_secret_hex, combined_hash_hex, seed_int_str, random_int).
    random_int is the PRNG float with the leading "0." stripped and cast to int:
    e.g. 0.739182... → 739182...
    """
    system_secret = secrets.token_bytes(32)
    log_entropy = log_json_str.encode('utf-8')
    combined_hash = hashlib.sha256(system_secret + log_entropy).digest()
    seed_int = int.from_bytes(combined_hash, 'big')
    rng = random.Random(seed_int)
    chaos_float = rng.random()
    # f"{:.17f}" always yields fixed-point notation ("0.000060805..."), unlike
    # str(), which switches to scientific notation ("6.08...e-05") for values
    # below 1e-4 and breaks the int() parse below.
    random_int = int(f'{chaos_float:.17f}'[2:])
    return system_secret.hex(), combined_hash.hex(), str(seed_int), random_int


class Command(BaseCommand):
    help = (
        'Long-running botseed processor: subscribes to request_stream, '
        'generates random numbers from traffic entropy, publishes to botseed_stream.'
    )

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._shutdown = threading.Event()
        self._pending = deque()
        self._pending_lock = threading.Lock()
        self._rps_window = deque()  # monotonic timestamps of received events
        self._rps_lock = threading.Lock()

    def handle(self, *args, **options):
        signal.signal(signal.SIGINT, self._handle_signal)
        signal.signal(signal.SIGTERM, self._handle_signal)

        redis_url = os.environ.get('REDIS_URL', 'redis://redis:6379/0')
        self.stdout.write(f'botseed_processor starting — Redis: {redis_url}')

        sender = threading.Thread(target=self._sender_loop, args=(redis_url,), daemon=True)
        sender.start()

        self._receiver_loop(redis_url)
        self.stdout.write('botseed_processor stopped.')

    def _handle_signal(self, signum, frame):
        self.stderr.write('Shutdown signal received — stopping.')
        self._shutdown.set()
        sys.exit(0)

    def _receiver_loop(self, redis_url):
        """Subscribe to request_stream, compute chaos values, push to pending deque."""
        delay = 5
        while not self._shutdown.is_set():
            try:
                r = redis.from_url(redis_url, decode_responses=True)
                pubsub = r.pubsub()
                pubsub.subscribe(REDIS_CHANNEL_IN)
                self.stdout.write(f"Subscribed to '{REDIS_CHANNEL_IN}'")
                delay = 5

                for message in pubsub.listen():
                    if self._shutdown.is_set():
                        return
                    if message['type'] != 'message':
                        continue

                    raw = message['data']

                    try:
                        log_data = json.loads(raw)
                    except Exception:
                        continue

                    # RPS: count events received in the last 1 second
                    now = time.monotonic()
                    with self._rps_lock:
                        self._rps_window.append(now)
                        while self._rps_window and self._rps_window[0] < now - 1.0:
                            self._rps_window.popleft()
                        rps = len(self._rps_window)

                    system_secret, combined_hash, seed_int, random_int = generate_chaos_number(raw)

                    result = json.dumps({
                        'random_int': random_int,
                        'seed_int': seed_int,
                        'combined_hash': combined_hash,
                        'system_secret': system_secret,
                        'log_json': log_data,
                        'requests_per_second': rps,
                        'generated_at': datetime.now(timezone.utc).isoformat(),
                    })

                    with self._pending_lock:
                        if len(self._pending) >= MAX_PENDING:
                            self._pending.popleft()  # drop oldest, keep freshest
                        self._pending.append(result)

            except Exception as exc:
                if not self._shutdown.is_set():
                    self.stderr.write(f'Receiver error ({exc}), retrying in {delay}s')
                    time.sleep(delay)
                    delay = min(delay * 2, 60)

    def _sender_loop(self, redis_url):
        """Drain pending deque at ≤20 events/sec; publish to botseed_stream + set botseed:latest."""
        last_sent = time.monotonic()
        r = None
        delay = 5

        while not self._shutdown.is_set():
            time.sleep(0.02)  # tight loop, check every 20ms

            now = time.monotonic()
            if now - last_sent < MIN_INTERVAL:
                continue

            with self._pending_lock:
                if not self._pending:
                    continue
                payload = self._pending.popleft()

            try:
                if r is None:
                    r = redis.from_url(
                        redis_url,
                        decode_responses=True,
                        socket_connect_timeout=2,
                        socket_timeout=1,
                    )
                r.set(REDIS_KEY_LATEST, payload)
                r.publish(REDIS_CHANNEL_OUT, payload)
                last_sent = now
                delay = 5

            except Exception as exc:
                self.stderr.write(f'Sender Redis error ({exc}), retrying in {delay}s')
                r = None
                time.sleep(delay)
                delay = min(delay * 2, 60)
