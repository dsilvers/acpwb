server {
    listen 80 backlog=65535;
    listen [::]:80 backlog=65535;
    server_name acpwb.com www.acpwb.com *.acpwb.com;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl backlog=65535;
    listen [::]:443 ssl backlog=65535;
    server_name www.acpwb.com;

    ssl_certificate     /etc/letsencrypt/live/acpwb.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/acpwb.com/privkey.pem;
    include             /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam         /etc/letsencrypt/ssl-dhparams.pem;

    return 301 https://acpwb.com$request_uri;
}

server {
    # backlog= intentionally omitted here — it's set once above on the same
    # 0.0.0.0:443/[::]:443 socket; nginx errors on "duplicate listen options"
    # if repeated, even with an identical value.
    listen 443 ssl;
    listen [::]:443 ssl;
    server_name acpwb.com *.acpwb.com;

    ssl_certificate     /etc/letsencrypt/live/acpwb.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/acpwb.com/privkey.pem;
    include             /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam         /etc/letsencrypt/ssl-dhparams.pem;

    # Buffered logging — flushes every 5s instead of a syscall per request
    access_log /var/log/nginx/acpwb.com.access.log acpwb buffer=16k flush=5s;
    error_log  /var/log/nginx/acpwb.com.error.log;

    # Static files served directly — no Docker round-trip
    location /static/ {
        alias /home/acpwb/acpwb/acpwb/staticfiles/;
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # Archive day-level trap pages (/archive/<year>/<month>/<day>[/<slug>][/])
    # are served by acpwb_go — the highest-volume archive content. This is
    # narrower than /archive/ itself: the index (/archive/) and year index
    # (/archive/<year>/) still fall through to django_backend below.
    location ~ "^/archive/\d{4}/\d{1,2}/\d{1,2}(/|$)" {
        proxy_pass         http://acpwb_go;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location /archive/ {
        proxy_pass         http://django_backend;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    # Public policy pages (index/year/month/detail) are served by acpwb_go —
    # the other bulk-content honeypot surface alongside archives. Subdomain
    # policy rendering (policy-<agency>.acpwb.com) is not yet cut over and
    # still falls through to django_backend below.
    location /public-policy/ {
        proxy_pass         http://acpwb_go;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location /wiki/ {
        proxy_pass         http://django_backend;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location /reports/ {
        proxy_pass         http://django_backend;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    # /.well-known/ must reach Django (honeypot endpoints — not cached)
    location /.well-known/ {
        proxy_pass         http://django_backend;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location /ws/requests/ {
        proxy_pass         http://127.0.0.1:8765;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade           $http_upgrade;
        proxy_set_header   Connection        "upgrade";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    location / {
        proxy_pass         http://django_backend;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
    }
}
