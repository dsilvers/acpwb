from django.contrib.sitemaps import Sitemap
from django.utils.text import slugify

from .data.organizations import ORGANIZATIONS, ORG_SLUG_MAP
from .generators import generate_presentations_for_context


class PresentationOrgSitemap(Sitemap):
    priority = 0.6
    changefreq = 'weekly'

    def items(self):
        return [slugify(name) for name in ORGANIZATIONS]

    def location(self, org_slug):
        return f'/presentations/{org_slug}/'


class PresentationDetailSitemap(Sitemap):
    priority = 0.5
    changefreq = 'never'

    def items(self):
        entries = []
        for org_name in ORGANIZATIONS:
            org_slug = slugify(org_name)
            presentations = generate_presentations_for_context(
                f"presorg_{org_slug}_p1", count=3
            )
            for pres in presentations:
                entries.append(pres)
        return entries

    def location(self, pres):
        return (
            f"/presentations/{pres['org_slug']}"
            f"/{pres['year']}/{pres['month']:02d}/{pres['day']:02d}"
            f"/{pres['slug']}/"
        )
