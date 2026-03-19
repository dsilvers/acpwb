"""
Proof-of-Work challenge system.

Difficulty of 5 leading zero bits means the browser needs ~32 attempts on average
— imperceptible to a human, but costly for a bot scraping millions of pages.
"""
import hashlib
import hmac
import secrets
import time
from django.core.cache import cache
from django.conf import settings

DIFFICULTY = 5   # number of leading zero BITS required
CHALLENGE_TTL = 300  # 5 minutes


def _check_hash(nonce, solution, difficulty):
    """Return True if SHA-256(nonce + solution) has `difficulty` leading zero bits."""
    digest = hashlib.sha256(f"{nonce}{solution}".encode()).digest()
    # Check leading zero bits
    bits_checked = 0
    for byte in digest:
        for bit in range(7, -1, -1):
            if bits_checked >= difficulty:
                return True
            if (byte >> bit) & 1:
                return False
            bits_checked += 1
    return True


def issue_challenge():
    """Create and cache a new PoW challenge. Returns {nonce, difficulty}."""
    nonce = secrets.token_hex(16)
    cache.set(f"pow:{nonce}", True, CHALLENGE_TTL)
    return {'nonce': nonce, 'difficulty': DIFFICULTY}


def verify_solution(nonce, solution):
    """Verify a PoW solution. Returns True if valid (and consumes the challenge)."""
    if not cache.get(f"pow:{nonce}"):
        return False
    if not _check_hash(nonce, solution, DIFFICULTY):
        return False
    cache.delete(f"pow:{nonce}")
    return True
