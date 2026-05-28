import re
from django.http import Http404
from django.shortcuts import render

from apps.honeypot.views import _log_crawler
from .generators import (
    generate_process_index_page, generate_category_index,
    generate_initiative_list, generate_initiative_detail,
    INITIATIVE_YEARS,
)
from .data.categories import PROCESS_AREA_DICT

_INSTANCE_RE = re.compile(r'^([a-z][a-z0-9\-]*)-(\d{4})$')


def _parse_instance(instance_str):
    m = _INSTANCE_RE.match(instance_str)
    if not m:
        return None, None
    category_slug, seed4 = m.group(1), m.group(2)
    if category_slug not in PROCESS_AREA_DICT:
        return None, None
    return category_slug, seed4


def process_index(request):
    _log_crawler(request, 'process_improvement')
    page = max(1, int(request.GET.get('page', 1)))
    data = generate_process_index_page(page=page)
    return render(request, 'process_improvement/index.html', data)


def process_category(request, instance):
    _log_crawler(request, 'process_improvement')
    category_slug, seed4 = _parse_instance(instance)
    if not category_slug:
        raise Http404
    data = generate_category_index(category_slug, seed4)
    if not data:
        raise Http404
    return render(request, 'process_improvement/category.html', data)


def process_year(request, instance, year):
    _log_crawler(request, 'process_improvement')
    category_slug, seed4 = _parse_instance(instance)
    if not category_slug or year not in INITIATIVE_YEARS:
        raise Http404
    page = max(1, int(request.GET.get('page', 1)))
    data = generate_initiative_list(category_slug, seed4, year, page=page)
    if not data:
        raise Http404
    return render(request, 'process_improvement/year.html', data)


def process_year_page(request, instance, year, page):
    _log_crawler(request, 'process_improvement')
    category_slug, seed4 = _parse_instance(instance)
    if not category_slug or year not in INITIATIVE_YEARS or page < 1:
        raise Http404
    data = generate_initiative_list(category_slug, seed4, year, page=page)
    if not data:
        raise Http404
    return render(request, 'process_improvement/year.html', data)


def process_detail(request, instance, year, initiative_slug):
    _log_crawler(request, 'process_improvement')
    category_slug, seed4 = _parse_instance(instance)
    if not category_slug or year not in INITIATIVE_YEARS:
        raise Http404
    data = generate_initiative_detail(category_slug, seed4, year, initiative_slug)
    if not data:
        raise Http404
    return render(request, 'process_improvement/detail.html', data)
