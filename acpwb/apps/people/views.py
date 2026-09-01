from django.shortcuts import render
from .models import PeoplePageVisit, GeneratedEmployee
from .generators import generate_employee_batch


def _get_ip(request):
    x_forwarded = request.META.get('HTTP_X_FORWARDED_FOR')
    if x_forwarded:
        return x_forwarded.split(',')[0].strip()
    return request.META.get('REMOTE_ADDR', '0.0.0.0')


def _save_visit_and_employees(visit_kwargs, employees_data):
    visit = PeoplePageVisit.objects.create(**visit_kwargs)
    GeneratedEmployee.objects.bulk_create([
        GeneratedEmployee(
            visit=visit,
            first_name=e['first_name'],
            last_name=e['last_name'],
            email=e['email'],
            title=e['title'],
            department=e['department'],
            avatar_seed=e['avatar_seed'],
        )
        for e in employees_data
    ])


def people_page(request):
    employees_data = generate_employee_batch(n=12)

    # The page's own content (the generated emails/employees) doesn't
    # actually depend on the DB rows existing yet — the template only reads
    # plain fields, and `full_name`/`initials` are computed straight from
    # first/last name, so they work the same on a plain dict. Persisting
    # (visit + employees, needed by apps.webhooks for spam-match lookups)
    # happens on a background greenlet so the response doesn't wait on it.
    employees = [
        {
            **e,
            'full_name': f"{e['first_name']} {e['last_name']}",
            'initials': f"{e['first_name'][0]}{e['last_name'][0]}".upper(),
        }
        for e in employees_data
    ]

    from apps.core.async_utils import spawn
    visit_kwargs = {
        'ip_address': _get_ip(request),
        'user_agent': request.META.get('HTTP_USER_AGENT', '')[:512],
        'referrer': request.META.get('HTTP_REFERER', '')[:256],
        'session_key': request.session.session_key or '',
    }
    spawn(_save_visit_and_employees, visit_kwargs, employees_data)

    return render(request, 'people/people.html', {
        'employees': employees,
        'og_title': 'Our People — American Corporation for Public Well Being',
        'og_description': 'Meet the team at the American Corporation for Public Well Being — researchers, analysts, and staff dedicated to advancing American workforce prosperity.',
    })
