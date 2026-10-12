#!/bin/sh
# Live check of acpwb_edge's fingerprints against Wireshark, on the local
# docker stack (docker-compose-local.yml must be up).
#
# Captures inside the edge container's network namespace with tshark (from
# the nicolaka/netshoot image — nothing to install), drives traffic at
# https://localhost:8001 (every uTLS browser profile via cmd/ja4probe, curl
# over TLS 1.0–1.3 / h1 / h2, and — with BROWSE=<seconds> — whatever real
# browsers you point at it meanwhile), drains acpwb:edge_conn_queue, then
# compares every ClientHello tshark saw against the EdgeConnection rows:
# same (JA4, JA4_r, JA3) set, same count each. Exits non-zero on mismatch.
#
#   acpwb_edge/scripts/ja4check.sh
#   BROWSE=60 acpwb_edge/scripts/ja4check.sh   # then open https://localhost:8001 in Chrome/Firefox/Safari
set -eu
cd "$(dirname "$0")/../.."
compose="docker compose -f docker-compose-local.yml"
edge=$($compose ps -q edge)
[ -n "$edge" ] || { echo "edge isn't running: $compose up -d edge" >&2; exit 1; }

work=$(mktemp -d)
trap 'docker rm -f ja4check-cap >/dev/null 2>&1 || true; rm -rf "$work"' EXIT

docker rm -f ja4check-cap >/dev/null 2>&1 || true
docker run -d --name ja4check-cap --net "container:$edge" -v "$work:/cap" nicolaka/netshoot \
	tshark -q -i any -f 'tcp port 8001' -w /cap/live.pcap >/dev/null
# Wait for tshark to be capturing before generating anything.
for _ in $(seq 50); do
	docker logs ja4check-cap 2>&1 | grep -q "Capturing on" && break
	sleep 0.2
done
start=$(date -u +%Y-%m-%dT%H:%M:%S)
sleep 1

echo "── uTLS browser profiles"
(cd acpwb_edge && go run ./cmd/ja4probe -addr localhost:8001) || true
echo "── curl"
# (TLS 1.3 is what --http1.1/--http2 negotiate; macOS's curl can't force it.)
for args in "--http1.1" "--http2" "--tlsv1.2 --tls-max 1.2" "--tlsv1.1 --tls-max 1.1 --http1.1" "--tlsv1.0 --tls-max 1.0 --http1.1"; do
	# shellcheck disable=SC2086
	printf '%-40s %s\n' "$args" "$(curl -sk -o /dev/null -w '%{http_code} %{http_version}' $args https://localhost:8001/ || echo FAIL)"
done
if [ "${BROWSE:-0}" -gt 0 ]; then
	echo "── browse https://localhost:8001 now (any browsers) — capturing for ${BROWSE}s"
	sleep "$BROWSE"
fi
sleep 2 # let connections close and the edge's queue flush
docker stop -t 5 ja4check-cap >/dev/null

$compose exec -T web python manage.py drain_edge_conn_queue >/dev/null

docker run --rm -v "$work:/cap:ro" nicolaka/netshoot \
	tshark -r /cap/live.pcap -Y 'tls.handshake.type == 1' -T fields -E separator=/t \
	-e tls.handshake.ja4 -e tls.handshake.ja4_r -e tls.handshake.ja3 >"$work/wireshark.tsv" 2>/dev/null

$compose exec -T db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tA -F "$(printf "\t")"' >"$work/edge.tsv" <<SQL
select ja4, ja4_r, ja3 from honeypot_edgeconnection where timestamp >= '$start' and ja4 <> '';
SQL

python3 - "$work/wireshark.tsv" "$work/edge.tsv" <<'PY'
import sys
from collections import Counter
def load(p):
    return Counter(tuple(l.rstrip("\n").split("\t")) for l in open(p) if l.strip())
ws, edge = load(sys.argv[1]), load(sys.argv[2])
print(f"── {sum(ws.values())} ClientHellos captured, {sum(edge.values())} EdgeConnection rows with a JA4")
bad = 0
for key in sorted(set(ws) | set(edge)):
    w, e = ws[key], edge[key]
    mark = "ok  " if w == e else "DIFF"
    bad += w != e
    print(f"{mark} wireshark={w:<3} edge={e:<3} {key[0]}  ja3={key[2]}")
if bad or not ws:
    print(f"FAIL: {bad} mismatched fingerprint(s)" if ws else "FAIL: nothing captured")
    sys.exit(1)
print("PASS: every ClientHello fingerprinted identically by the edge and Wireshark (JA4, JA4_r, JA3)")
PY
