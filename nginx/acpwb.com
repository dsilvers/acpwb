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

    # /.well-known/ must reach Django (honeypot endpoints)
    location /.well-known/ {
        proxy_pass         http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location / {
        proxy_pass         http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
    }
}
