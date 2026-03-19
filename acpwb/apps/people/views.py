from django.shortcuts import render
from .models import PeoplePageVisit, GeneratedEmployee
from .generators import generate_employee_batch


def _get_ip(request):
    x_forwarded = request.META.get('HTTP_X_FORWARDED_FOR')
    if x_forwarded:
        return x_forwarded.split(',')[0].strip()
    return request.META.get('REMOTE_ADDR', '0.0.0.0')


def people_page(request):
    visit = PeoplePageVisit.objects.create(
        ip_address=_get_ip(request),
        user_agent=request.META.get('HTTP_USER_AGENT', '')[:512],
        referrer=request.META.get('HTTP_REFERER', '')[:256],
        session_key=request.session.session_key or '',
    )

    employees_data = generate_employee_batch(n=12)
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

    # Re-query to get model instances with IDs
    employees = visit.employees.all()

    return render(request, 'people/people.html', {
        'employees': employees,
        'visit': visit,
    })
