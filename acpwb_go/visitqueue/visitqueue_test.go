package visitqueue

import "testing"

// Reference buckets from Python: int(hashlib.md5(ip.encode()).hexdigest()[:8], 16) / 2**32.
// Must stay in lockstep with apps.core.signal_capture.fingerprint_sampled —
// acpwb/tests/test_signal_capture.py asserts the same cases.
func TestFingerprintSampledMatchesPython(t *testing.T) {
	cases := []struct {
		ip   string
		rate float64
		want bool
	}{
		{"203.0.113.9", 0.51, true},  // bucket fraction 0.5084
		{"203.0.113.9", 0.50, false},
		{"2001:db8::1", 0.12, true}, // 0.1148
		{"2001:db8::1", 0.11, false},
		{"8.8.8.8", 0.26, true}, // 0.2539
		{"8.8.8.8", 0.25, false},
		{"45.148.10.19", 1.0, true},
		{"45.148.10.19", 0.0, false},
	}
	for _, c := range cases {
		if got := FingerprintSampled(c.ip, c.rate); got != c.want {
			t.Errorf("FingerprintSampled(%q, %v) = %v, want %v", c.ip, c.rate, got, c.want)
		}
	}
}

func TestFingerprintCaptureDefaultsOff(t *testing.T) {
	t.Setenv("FINGERPRINT_CAPTURE_ENABLED", "")
	q, err := New("redis://127.0.0.1:6379/0")
	if err != nil {
		t.Fatal(err)
	}
	if q.fingerprintEnabled {
		t.Error("fingerprint capture must default to off")
	}
	if q.fingerprintQueueMax != 1_000_000 || q.fingerprintSampleRate != 1.0 {
		t.Errorf("unexpected defaults: max=%d rate=%v", q.fingerprintQueueMax, q.fingerprintSampleRate)
	}

	t.Setenv("FINGERPRINT_CAPTURE_ENABLED", "true")
	t.Setenv("FINGERPRINT_SAMPLE_RATE", "0.01")
	q, _ = New("redis://127.0.0.1:6379/0")
	if !q.fingerprintEnabled || q.fingerprintSampleRate != 0.01 {
		t.Errorf("env not applied: enabled=%v rate=%v", q.fingerprintEnabled, q.fingerprintSampleRate)
	}
}
