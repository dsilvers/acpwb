# Pinned to 3.0+ for the native req.ssl_ja3/req.ssl_ja4 fetches haproxy.cfg
# uses — verify this version still has them if bumping the tag.
FROM haproxy:3.0-alpine

USER root

# Local-dev-only self-signed cert, baked into the image at build time —
# there is no real domain/DNS for this frontend to get a Let's Encrypt cert
# for, and it only needs to exist so the JA3/JA4-capture path can be
# exercised end-to-end locally. Never used in production — see
# nginx/acpwb.com for the real cert path (certbot + Let's Encrypt) and the
# plan this implements for why HAProxy in production would need its own,
# separate cert-renewal handoff.
RUN apk add --no-cache openssl \
    && mkdir -p /usr/local/etc/haproxy/certs \
    && openssl req -x509 -nodes -newkey rsa:2048 -days 3650 \
        -keyout /usr/local/etc/haproxy/certs/key.pem \
        -out /usr/local/etc/haproxy/certs/cert.pem \
        -subj "/CN=acpwb.local" \
    && cat /usr/local/etc/haproxy/certs/cert.pem /usr/local/etc/haproxy/certs/key.pem \
        > /usr/local/etc/haproxy/certs/combined.pem \
    && chown -R haproxy:haproxy /usr/local/etc/haproxy/certs

COPY haproxy.cfg /usr/local/etc/haproxy/haproxy.cfg

USER haproxy
