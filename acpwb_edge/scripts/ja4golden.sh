#!/bin/sh
# Regenerates tlsfp/testdata/hellos: byte-exact browser ClientHellos (uTLS
# profiles, via cmd/ja4probe) and golden.tsv — Wireshark's own JA4, JA4_r,
# JA3 and JA3 string for each, which TestBrowserHellosMatchWireshark checks
# tlsfp against. tshark runs from the nicolaka/netshoot image, so nothing
# needs installing locally beyond Docker.
set -eu
cd "$(dirname "$0")/.."
dir=tlsfp/testdata/hellos
go run ./cmd/ja4probe -write-fixtures "$dir"
docker run --rm -v "$PWD/$dir:/cap:ro" nicolaka/netshoot \
	tshark -r /cap/hellos.pcap -Y 'tls.handshake.type == 1' -T fields -E separator=/t \
	-e tcp.srcport -e tls.handshake.ja4 -e tls.handshake.ja4_r -e tls.handshake.ja3 -e tls.handshake.ja3_full \
	> "$dir/golden.tsv"
echo "wrote $(wc -l < "$dir/golden.tsv") golden rows to $dir/golden.tsv"
