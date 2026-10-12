#!/bin/sh
# certbot deploy hook: copy a renewed certificate where acpwb_edge reads it
# and tell the edge to reload. Install as
#   /etc/letsencrypt/renewal-hooks/deploy/acpwb-edge.sh  (root, 0755)
# so certbot runs it after every successful renewal of any lineage
# (acpwb.com via dns-cloudflare, botseed.net via webroot).
#
# /etc/letsencrypt/{live,archive} are root-only, and the edge runs as acpwb,
# hence the copy rather than pointing the edge at live/ directly.
#
# Seed it once by hand at install time, per lineage:
#   sudo RENEWED_LINEAGE=/etc/letsencrypt/live/acpwb.com   /etc/letsencrypt/renewal-hooks/deploy/acpwb-edge.sh
#   sudo RENEWED_LINEAGE=/etc/letsencrypt/live/botseed.net /etc/letsencrypt/renewal-hooks/deploy/acpwb-edge.sh
set -eu

: "${RENEWED_LINEAGE:?certbot sets RENEWED_LINEAGE}"
name=$(basename "$RENEWED_LINEAGE")
dest=/etc/acpwb-edge/certs/$name

install -d -o root -g acpwb -m 0750 /etc/acpwb-edge /etc/acpwb-edge/certs "$dest"
# Write both files under temporary names and rename, so the edge never
# loads a new chain with the old key (or vice versa).
install -o root -g acpwb -m 0640 "$RENEWED_LINEAGE/fullchain.pem" "$dest/.fullchain.pem.new"
install -o root -g acpwb -m 0640 "$RENEWED_LINEAGE/privkey.pem" "$dest/.privkey.pem.new"
mv -f "$dest/.privkey.pem.new" "$dest/privkey.pem"
mv -f "$dest/.fullchain.pem.new" "$dest/fullchain.pem"

# Not an error before the edge is installed/running (e.g. while seeding).
systemctl reload acpwb-edge 2>/dev/null || true
echo "acpwb-edge: installed $name certificate"
