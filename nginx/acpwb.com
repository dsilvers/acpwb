server {
    listen 80;
    listen [::]:80;
    server_name acpwb.com www.acpwb.com;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://acpwb.com$request_uri;
    }
}

server {
    listen 443 ssl;
    listen [::]:443 ssl;
    server_name www.acpwb.com;

    ssl_certificate     /etc/letsencrypt/live/acpwb.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/acpwb.com/privkey.pem;
    include             /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam         /etc/letsencrypt/ssl-dhparams.pem;

    return 301 https://acpwb.com$request_uri;
}

proxy_cache_path /tmp/nginx_trap_cache levels=1:2 keys_zone=trap_cache:10m max_size=200m inactive=15m use_temp_path=off;

server {
    listen 443 ssl;
    listen [::]:443 ssl;
    server_name acpwb.com;

    ssl_certificate     /etc/letsencrypt/live/acpwb.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/acpwb.com/privkey.pem;
    include             /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam         /etc/letsencrypt/ssl-dhparams.pem;

    access_log /var/log/nginx/acpwb.com.access.log;
    error_log  /var/log/nginx/acpwb.com.error.log;

    # Static files served directly — no Docker round-trip
    location /static/ {
        alias /home/dan/acpwb.com/acpwb/staticfiles/;
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # Archive — fully deterministic; cache aggressively
    location /archive/ {
        proxy_pass         http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_cache        trap_cache;
        proxy_cache_key    "$request_method$host$request_uri";
        proxy_cache_valid  200 15m;
        proxy_cache_valid  any 0;
        proxy_cache_use_stale error timeout updating;
        proxy_cache_lock   on;
        add_header         X-Cache-Status $upstream_cache_status;
    }

    # Wiki — deterministic per slug
    location /wiki/ {
        proxy_pass         http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_cache        trap_cache;
        proxy_cache_key    "$request_method$host$request_uri";
        proxy_cache_valid  200 15m;
        proxy_cache_valid  any 0;
        proxy_cache_use_stale error timeout updating;
        proxy_cache_lock   on;
        add_header         X-Cache-Status $upstream_cache_status;
    }

    # Reports — deterministic per slug/page
    location /reports/ {
        proxy_pass         http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_cache        trap_cache;
        proxy_cache_key    "$request_method$host$request_uri";
        proxy_cache_valid  200 15m;
        proxy_cache_valid  any 0;
        proxy_cache_use_stale error timeout updating;
        proxy_cache_lock   on;
        add_header         X-Cache-Status $upstream_cache_status;
    }

    # /.well-known/ must reach Django (honeypot endpoints — not cached)
    location /.well-known/ {
        proxy_pass         http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location / {
        proxy_pass         http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
    }
}
