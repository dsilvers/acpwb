import pytest
from apps.people.generators import generate_employee_batch
from apps.people.models import PeoplePageVisit, GeneratedEmployee


def test_generate_employee_batch_default_count():
    employees = generate_employee_batch()
    assert len(employees) == 12


def test_generate_employee_batch_custom_count():
    employees = generate_employee_batch(5)
    assert len(employees) == 5


def test_employee_email_format():
    employees = generate_employee_batch(20)
    for emp in employees:
        assert emp['email'].endswith('@acpwb.com')
        # email should be firstname.lastname@acpwb.com (or with digit suffix)
        local = emp['email'].split('@')[0]
        first = emp['first_name'].lower()
        last = emp['last_name'].lower()
        assert local.startswith(f'{first}.{last}')


def test_employee_emails_unique_within_batch():
    employees = generate_employee_batch(30)
    emails = [e['email'] for e in employees]
    assert len(emails) == len(set(emails))


def test_employee_has_required_fields():
    employees = generate_employee_batch(3)
    for emp in employees:
        assert 'first_name' in emp
        assert 'last_name' in emp
        assert 'email' in emp
        assert 'title' in emp
        assert 'department' in emp
        assert 'avatar_seed' in emp


def test_employees_sorted_by_last_name():
    employees = generate_employee_batch(20)
    last_names = [e['last_name'] for e in employees]
    assert last_names == sorted(last_names)


@pytest.mark.django_db
def test_people_page_creates_visit(client):
    assert PeoplePageVisit.objects.count() == 0
    client.get('/our-people/')
    assert PeoplePageVisit.objects.count() == 1


@pytest.mark.django_db
def test_people_page_creates_employees(client):
    assert GeneratedEmployee.objects.count() == 0
    client.get('/our-people/')
    assert GeneratedEmployee.objects.count() == 12


@pytest.mark.django_db
def test_people_page_links_employees_to_visit(client):
    client.get('/our-people/')
    visit = PeoplePageVisit.objects.first()
    assert GeneratedEmployee.objects.filter(visit=visit).count() == 12


@pytest.mark.django_db
def test_people_page_different_employees_each_load(client):
    client.get('/our-people/')
    client.get('/our-people/')
    assert GeneratedEmployee.objects.count() == 24
    # emails across both loads should not be guaranteed unique, but visits are distinct
    assert PeoplePageVisit.objects.count() == 2


@pytest.mark.django_db
def test_generated_employee_full_name():
    from apps.people.models import PeoplePageVisit, GeneratedEmployee
    visit = PeoplePageVisit.objects.create(
        ip_address='127.0.0.1',
        user_agent='test',
        referrer='',
        session_key='abc',
    )
    emp = GeneratedEmployee.objects.create(
        visit=visit,
        first_name='Jane',
        last_name='Smith',
        email='jane.smith@acpwb.com',
        title='Director',
        department='Strategy',
        avatar_seed='abc123',
    )
    assert emp.full_name == 'Jane Smith'
