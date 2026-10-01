import pytest
from django.urls import reverse

from apps.public.conference_data import CONFERENCES, CURRENT_YEAR
from apps.public.models import ConferenceRegistration


CURRENT_PAGES = [
    'perch-conference', 'perch-conference-speakers', 'perch-conference-schedule',
    'perch-conference-about', 'perch-conference-venue', 'perch-conference-sponsors',
    'perch-conference-dinner', 'perch-conference-register',
]


def test_current_year_is_open_and_only_one_is():
    assert CONFERENCES[CURRENT_YEAR]['registration_open'] is True
    assert [y for y, c in CONFERENCES.items() if c['registration_open']] == [CURRENT_YEAR]


@pytest.mark.django_db
@pytest.mark.parametrize('name', CURRENT_PAGES)
def test_current_pages_show_current_year(client, name):
    resp = client.get(reverse(name))
    assert resp.status_code == 200
    body = resp.content.decode()
    assert f'PERCH {CURRENT_YEAR}' in body
    assert f'PERCH {CURRENT_YEAR - 1} &mdash;' not in body


@pytest.mark.django_db
@pytest.mark.parametrize('suffix', ['', '-speakers', '-schedule'])
def test_previous_year_is_archived(client, suffix):
    prev = CURRENT_YEAR - 1
    resp = client.get(reverse(f'perch-conference-archive{suffix}', args=[prev]))
    assert resp.status_code == 200
    assert CONFERENCES[prev]['theme'].replace("'", '&#x27;') in resp.content.decode() or suffix


@pytest.mark.django_db
def test_registration_uses_current_year(client):
    resp = client.post(reverse('perch-conference-register'), {
        'first_name': 'Pat', 'last_name': 'Example', 'email': 'pat@example.com',
        'address_line1': '1 Main St', 'city': 'Milwaukee', 'state': 'WI',
        'postal_code': '53202', 'badge_name': 'Pat', 'registration_type': 'full',
    })
    reg = ConferenceRegistration.objects.get(email='pat@example.com')
    assert reg.year == CURRENT_YEAR
    confirmation = client.get(resp['Location']).content.decode()
    assert f'PERCH-{CURRENT_YEAR}-' in confirmation
    assert CONFERENCES[CURRENT_YEAR]['dinner_venue'] in confirmation


@pytest.mark.django_db
def test_old_registration_confirmation_keeps_its_year(client):
    prev = CURRENT_YEAR - 1
    reg = ConferenceRegistration.objects.create(
        year=prev, first_name='Old', last_name='Reg', email='old@example.com',
        address_line1='x', city='x', state='x', postal_code='x', badge_name='x',
    )
    body = client.get(reverse('perch-conference-register-confirmation', args=[reg.token])).content.decode()
    assert f'PERCH {prev}' in body
    assert CONFERENCES[prev]['dinner_venue'] in body
