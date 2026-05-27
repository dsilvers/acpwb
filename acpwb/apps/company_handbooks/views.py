import re
from django.http import Http404
from django.shortcuts import render

from apps.honeypot.views import _log_crawler
from .generators import (
    generate_agency_index_page, generate_handbook_index,
    generate_year_index, generate_revision_toc, generate_group_page,
    HANDBOOK_YEARS,
)
from .data.sections import GROUP_SLUG_LIST
from apps.honeypot.policy_data import AGENCIES

_SEED4_RE = re.compile(r'^\d{4}$')
_INSTANCE_RE = re.compile(r'^([a-z0-9\-]+)-(\d{4})$')


def _parse_instance(instance_str):
    m = _INSTANCE_RE.match(instance_str)
    if not m:
        return None, None
    agency_slug, seed4 = m.group(1), m.group(2)
    if agency_slug not in AGENCIES:
        return None, None
    return agency_slug, seed4


def handbook_index(request):
    _log_crawler(request, 'handbook')
    page = max(1, int(request.GET.get('page', 1)))
    data = generate_agency_index_page(page=page)
    return render(request, 'company_handbooks/index.html', data)


def handbook_agency(request, instance):
    _log_crawler(request, 'handbook')
    agency_slug, seed4 = _parse_instance(instance)
    if not agency_slug:
        raise Http404
    data = generate_handbook_index(agency_slug, seed4)
    if not data:
        raise Http404
    return render(request, 'company_handbooks/agency.html', data)


def handbook_year(request, instance, year):
    _log_crawler(request, 'handbook')
    agency_slug, seed4 = _parse_instance(instance)
    if not agency_slug or year not in HANDBOOK_YEARS:
        raise Http404
    data = generate_year_index(agency_slug, seed4, year)
    return render(request, 'company_handbooks/year.html', data)


def handbook_revision(request, instance, year, revision):
    _log_crawler(request, 'handbook')
    agency_slug, seed4 = _parse_instance(instance)
    if not agency_slug or year not in HANDBOOK_YEARS or revision < 1:
        raise Http404
    data = generate_revision_toc(agency_slug, seed4, year, revision)
    return render(request, 'company_handbooks/revision.html', data)


def handbook_group(request, instance, year, revision, group_slug):
    _log_crawler(request, 'handbook')
    agency_slug, seed4 = _parse_instance(instance)
    if not agency_slug or year not in HANDBOOK_YEARS or revision < 1:
        raise Http404
    if group_slug not in GROUP_SLUG_LIST:
        raise Http404
    data = generate_group_page(agency_slug, seed4, year, revision, group_slug)
    if not data:
        raise Http404
    return render(request, 'company_handbooks/section.html', data)
