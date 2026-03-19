import pytest
from django.test import RequestFactory
from django.template import Context, Template
from apps.core.templatetags.acpwb_tags import avatar_card, initials, AVATAR_PALETTES


# ── avatar_card tag ────────────────────────────────────────────────────────────

def test_avatar_card_returns_div():
    result = avatar_card('some-seed', 'AB')
    assert result.startswith('<div')
    assert result.endswith('</div>')


def test_avatar_card_contains_initials():
    result = avatar_card('seed123', 'JD')
    assert 'JD' in result


def test_avatar_card_contains_gradient():
    result = avatar_card('seed123', 'JD')
    assert 'linear-gradient' in result


def test_avatar_card_default_size():
    result = avatar_card('seed', 'AB')
    assert 'width:80px' in result
    assert 'height:80px' in result


def test_avatar_card_custom_size():
    result = avatar_card('seed', 'AB', size=120)
    assert 'width:120px' in result
    assert 'height:120px' in result


def test_avatar_card_deterministic():
    result1 = avatar_card('fixed-seed', 'XY')
    result2 = avatar_card('fixed-seed', 'XY')
    assert result1 == result2


def test_avatar_card_different_seeds_may_differ():
    result1 = avatar_card('seed-aaa', 'AB')
    result2 = avatar_card('seed-zzz', 'AB')
    # Different seeds should (very likely) produce different palette colors
    # We can't guarantee they differ, but we can verify both are valid HTML
    assert '<div' in result1
    assert '<div' in result2


def test_avatar_card_is_marked_safe():
    from django.utils.safestring import SafeData
    result = avatar_card('seed', 'AB')
    assert isinstance(result, SafeData)


def test_avatar_card_uses_valid_palette():
    result = avatar_card('seed', 'AB')
    # One of the palette color pairs should appear in the output
    found = any(color1.lower() in result.lower() for color1, _ in AVATAR_PALETTES)
    assert found


# ── initials filter ────────────────────────────────────────────────────────────

def test_initials_two_names():
    assert initials('John Smith') == 'JS'


def test_initials_three_names_uses_first_and_last():
    assert initials('John Michael Smith') == 'JS'


def test_initials_single_word():
    result = initials('Alice')
    assert len(result) == 2


def test_initials_uppercased():
    assert initials('john smith') == 'JS'


def test_initials_empty_string():
    result = initials('')
    assert result == '??'


def test_initials_strips_whitespace():
    assert initials('  Jane  Doe  ') == 'JD'


# ── Template rendering ─────────────────────────────────────────────────────────

def test_avatar_card_in_template():
    template = Template(
        '{% load acpwb_tags %}'
        '{% avatar_card "seed123" "AB" %}'
    )
    result = template.render(Context({}))
    assert '<div' in result
    assert 'AB' in result


def test_initials_filter_in_template():
    template = Template(
        '{% load acpwb_tags %}'
        '{{ name|initials }}'
    )
    result = template.render(Context({'name': 'Jane Doe'}))
    assert result == 'JD'
