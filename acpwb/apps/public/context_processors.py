from .conference_data import CONFERENCES, CURRENT_YEAR


def perch_context(request):
    """Expose the upcoming PERCH conference to every template (nav, home, archive links)."""
    return {
        'perch_year': CURRENT_YEAR,
        'perch_conf': CONFERENCES[CURRENT_YEAR],
    }
