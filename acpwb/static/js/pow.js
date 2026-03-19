/**
 * ACPWB Proof-of-Work client
 * Finds a solution to SHA-256(nonce + solution) with DIFFICULTY leading zero bits.
 */
(function () {
  'use strict';

  async function sha256hex(message) {
    const msgBuffer = new TextEncoder().encode(message);
    const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
  }

  function countLeadingZeroBits(hexString) {
    let bits = 0;
    for (let i = 0; i < hexString.length; i++) {
      const nibble = parseInt(hexString[i], 16);
      if (nibble === 0) {
        bits += 4;
      } else {
        // Count leading zeros in this nibble
        if (nibble < 8) bits++;
        if (nibble < 4) bits++;
        if (nibble < 2) bits++;
        break;
      }
    }
    return bits;
  }

  async function solve(nonce, difficulty) {
    let counter = 0;
    while (true) {
      const candidate = counter.toString();
      const hash = await sha256hex(nonce + candidate);
      if (countLeadingZeroBits(hash) >= difficulty) {
        return candidate;
      }
      counter++;
      // Yield to browser every 1000 iterations
      if (counter % 1000 === 0) {
        await new Promise(r => setTimeout(r, 0));
      }
    }
  }

  async function runPoW() {
    const statusEl = document.getElementById('pow-status');
    const containerEl = document.getElementById('pow-container');

    if (!statusEl) return;

    try {
      statusEl.textContent = 'Verifying your browser...';

      // Get challenge
      const challengeRes = await fetch('/api/pow/challenge/');
      const { nonce, difficulty } = await challengeRes.json();

      statusEl.textContent = 'Computing verification...';

      const solution = await solve(nonce, difficulty);

      // Submit solution
      const verifyRes = await fetch('/api/pow/verify/', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ nonce, solution }),
      });

      const result = await verifyRes.json();

      if (result.valid) {
        statusEl.textContent = 'Verification complete. Loading content...';
        // Reload the page — session now has the valid pow_token
        window.location.reload();
      } else {
        statusEl.textContent = 'Verification failed. Retrying...';
        setTimeout(runPoW, 2000);
      }
    } catch (err) {
      console.error('PoW error:', err);
      if (statusEl) statusEl.textContent = 'Loading content...';
      setTimeout(() => window.location.reload(), 2000);
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    if (document.getElementById('pow-container')) {
      runPoW();
    }
  });
})();
