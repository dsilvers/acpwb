import sys
import environ
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent.parent

# True when running under pytest. Honeypot logging normally defers Redis/DB
# writes to a background gevent greenlet (see apps.core.crawler_queue), which
# relies on gunicorn's gevent worker monkey-patching the process — that never
# happens under the test runner, so nothing yields control back to the
# greenlet before a test's assertions run. Tests need those writes to stay
# synchronous.
TESTING = 'pytest' in sys.modules

env = environ.Env(
    DJANGO_DEBUG=(bool, False),
    DJANGO_ALLOWED_HOSTS=(list, ['.acpwb.com', 'localhost', '127.0.0.1']),
)

environ.Env.read_env(BASE_DIR / '.env')

SECRET_KEY = env('DJANGO_SECRET_KEY')
DEBUG = env('DJANGO_DEBUG')
ALLOWED_HOSTS = env('DJANGO_ALLOWED_HOSTS')

INSTALLED_APPS = [
    'django.contrib.admin',
    'django.contrib.auth',
    'django.contrib.contenttypes',
    'django.contrib.sessions',
    'django.contrib.messages',
    'django.contrib.staticfiles',
    'django.contrib.sitemaps',
    # Project apps
    'apps.core',
    'apps.public',
    'apps.people',
    'apps.projects',
    'apps.honeypot',
    'apps.webhooks',
    'apps.company_handbooks',
    'apps.process_improvement',
    'apps.presentations',
]

MIDDLEWARE = [
    'apps.core.stream_middleware.RequestStreamMiddleware',
    'django.middleware.security.SecurityMiddleware',
    'whitenoise.middleware.WhiteNoiseMiddleware',
    'django.middleware.common.CommonMiddleware',
    'apps.core.subdomain_middleware.SubdomainMiddleware',
    'django.middleware.csrf.CsrfViewMiddleware',
    'apps.core.middleware.ConditionalAuthMiddleware',
    'django.middleware.clickjacking.XFrameOptionsMiddleware',
    'apps.core.middleware.BotTrackingMiddleware',
]

ROOT_URLCONF = 'config.urls'

TEMPLATES = [
    {
        'BACKEND': 'django.template.backends.jinja2.Jinja2',
        'DIRS': [BASE_DIR / 'templates/jinja2'],
        'APP_DIRS': False,
        'OPTIONS': {
            'environment': 'config.jinja2_env.environment',
        },
    },
    {
        'BACKEND': 'django.template.backends.django.DjangoTemplates',
        'DIRS': [BASE_DIR / 'templates'],
        'APP_DIRS': True,
        'OPTIONS': {
            'context_processors': [
                'django.template.context_processors.debug',
                'django.template.context_processors.request',
                'django.contrib.auth.context_processors.auth',
                'django.contrib.messages.context_processors.messages',
                'apps.core.context_processors.honeypot_context',
            ],
        },
    },
]

WSGI_APPLICATION = 'config.wsgi.application'

DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql',
        'NAME': env('DB_NAME', default='acpwb'),
        'USER': env('DB_USER', default='acpwb'),
        'PASSWORD': env('DB_PASSWORD', default='acpwb_dev'),
        'HOST': env('DB_HOST', default='db'),
        'PORT': env('DB_PORT', default='5432'),
        'CONN_MAX_AGE': env.int('DB_CONN_MAX_AGE', default=60),
    }
}

AUTH_PASSWORD_VALIDATORS = [
    {'NAME': 'django.contrib.auth.password_validation.UserAttributeSimilarityValidator'},
    {'NAME': 'django.contrib.auth.password_validation.MinimumLengthValidator'},
    {'NAME': 'django.contrib.auth.password_validation.CommonPasswordValidator'},
    {'NAME': 'django.contrib.auth.password_validation.NumericPasswordValidator'},
]

LANGUAGE_CODE = 'en-us'
TIME_ZONE = 'America/Chicago'
USE_I18N = True
USE_TZ = True

STATIC_URL = '/static/'
STATIC_ROOT = BASE_DIR / 'staticfiles'
STATICFILES_DIRS = [BASE_DIR / 'static']
STATICFILES_STORAGE = 'whitenoise.storage.CompressedManifestStaticFilesStorage'

DEFAULT_AUTO_FIELD = 'django.db.models.BigAutoField'

# ConditionalAuthMiddleware wraps Session/Auth/Message internally for /django-admin paths.
# Django's system check only scans MIDDLEWARE strings, so silence those false positives.
SILENCED_SYSTEM_CHECKS = ['admin.E408', 'admin.E409', 'admin.E410']

# GeoIP / IP intelligence (see apps/core/management/commands/download_geoip_db.py,
# discover_ip_intelligence.py, enrich_ip_intelligence.py, ip_intelligence_report.py)
MAXMIND_ACCOUNT_ID = env('MAXMIND_ACCOUNT_ID', default='')
MAXMIND_LICENSE_KEY = env('MAXMIND_LICENSE_KEY', default='')
GEOIP2_CITY_DB_PATH = env('GEOIP2_CITY_DB_PATH', default=str(BASE_DIR / 'var/geoip/GeoLite2-City.mmdb'))
GEOIP2_ASN_DB_PATH = env('GEOIP2_ASN_DB_PATH', default=str(BASE_DIR / 'var/geoip/GeoLite2-ASN.mmdb'))
TOR_EXIT_LIST_PATH = env('TOR_EXIT_LIST_PATH', default=str(BASE_DIR / 'var/tor_exit_nodes.txt'))
TOR_EXIT_LIST_URL = env('TOR_EXIT_LIST_URL', default='https://check.torproject.org/torbulkexitlist')

CSRF_TRUSTED_ORIGINS = env.list('CSRF_TRUSTED_ORIGINS', default=[
    'https://acpwb.com',
    'https://*.acpwb.com',
    'http://acpwb.example',
    'http://*.acpwb.example',
    'http://localhost',
    'http://127.0.0.1',
])

# Mailgun
MAILGUN_WEBHOOK_SIGNING_KEY = env('MAILGUN_WEBHOOK_SIGNING_KEY', default='')
MAILGUN_DOMAIN = env('MAILGUN_DOMAIN', default='acpwb.com')

# Twilio
TWILIO_ACCOUNT_SID = env('TWILIO_ACCOUNT_SID', default='')
TWILIO_AUTH_TOKEN  = env('TWILIO_AUTH_TOKEN',  default='')

# Proxy headers — nginx terminates TLS; needed so build_absolute_uri() returns https://
# (required for Twilio webhook signature verification to match in production)
SECURE_PROXY_SSL_HEADER = ('HTTP_X_FORWARDED_PROTO', 'https')
USE_X_FORWARDED_HOST = True

# Canary Tokens
CANARYTOKENS_WEBHOOK_URL = env('CANARYTOKENS_WEBHOOK_URL', default=None)

# Real-time request stream (Redis pub/sub → WebSocket service)
REDIS_URL = env('REDIS_URL', default='redis://redis:6379/0')
STREAM_WS_TOKEN = env('STREAM_WS_TOKEN', default='')


# Sentry
_SENTRY_DSN = env('SENTRY_DSN', default='')
if _SENTRY_DSN:
    import sentry_sdk
    from sentry_sdk.integrations.django import DjangoIntegration
    sentry_sdk.init(
        dsn=_SENTRY_DSN,
        integrations=[DjangoIntegration()],
        # traces_sample_rate=0.1,
        send_default_pii=True,
    )

