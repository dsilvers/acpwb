from django.contrib.sitemaps import Sitemap
from django.urls import reverse

from apps.projects.models import ProjectStory


class StaticPagesSitemap(Sitemap):
    priority = 0.8
    changefreq = 'weekly'

    def items(self):
        return [
            'home', 'our-people', 'mission', 'project-list',
            'careers', 'partners', 'privacy', 'reports-list',
        ]

    def location(self, item):
        return reverse(item)


class ProjectStorySitemap(Sitemap):
    priority = 0.5
    changefreq = 'never'

    def items(self):
        return ProjectStory.objects.all().order_by('page_number', 'slug')

    def location(self, obj):
        return f'/projects/{obj.slug}/'

    def lastmod(self, obj):
        return obj.generated_at
