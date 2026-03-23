from .base import *  # noqa

DEBUG = True

INTERNAL_IPS = ['127.0.0.1']

# Django test client uses 'testserver' as HTTP_HOST
# .acpwb.example covers local dnsmasq wildcard dev domain (archives-YYYY.acpwb.example)
ALLOWED_HOSTS = ['.acpwb.com', '.acpwb.example', 'localhost', '127.0.0.1', 'testserver']
