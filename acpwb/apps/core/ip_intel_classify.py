"""
Best-effort hosting/datacenter classification from an ASN organization name.

This is a heuristic, not an authoritative VPN/proxy/residential-proxy
detector: it only catches ASN orgs that are themselves well-known cloud or
hosting providers. A residential ISP reselling cloud capacity, or a small
business legitimately hosted on AWS, will be misclassified. Treat is_hosting
as "probably not a home internet connection," nothing stronger.
"""

_HOSTING_ORG_KEYWORDS = [
    'AMAZON', 'AWS', 'GOOGLE', 'MICROSOFT', 'AZURE',
    'DIGITALOCEAN', 'DIGITAL OCEAN', 'OVH', 'HETZNER', 'LINODE', 'AKAMAI',
    'VULTR', 'CHOOPA', 'ALIBABA', 'ALIYUN', 'TENCENT', 'ORACLE',
    'CONTABO', 'LEASEWEB', 'SCALEWAY', 'ONLINE S.A.S', 'M247',
    'COLOCROSSING', 'ZENLAYER', 'HOSTINGER', 'IONOS', 'GODADDY',
    'RACKSPACE', 'PSYCHZ', 'DEDIPATH', 'FRANTECH', 'DATACAMP',
    'G-CORE', 'GCORE', 'CLOUDFLARE', 'FASTLY', 'CLOUDVPS',
    'HOSTWINDS', 'HOSTKEY', 'NFORCE', 'SERVERASTRA', 'HIVELOCITY',
    'QUADRANET', 'UNIFIEDLAYER', 'BLUEHOST', 'DREAMHOST', 'OVHCLOUD',
    'KAMATERA', 'UPCLOUD', 'CLOUDSIGMA', 'INTERSERVER', 'RELIABLESITE',
]


def classify_hosting(asn_org: str) -> bool:
    """Return True if asn_org looks like a cloud/hosting provider."""
    if not asn_org:
        return False
    upper = asn_org.upper()
    return any(keyword in upper for keyword in _HOSTING_ORG_KEYWORDS)
