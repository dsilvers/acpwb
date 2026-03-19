import hashlib
import pytest
from apps.projects.pow import _check_hash, issue_challenge, verify_solution, DIFFICULTY


# ── PoW logic ─────────────────────────────────────────────────────────────────

def _find_valid_solution(nonce, difficulty=DIFFICULTY):
    """Brute-force a valid PoW solution (for testing)."""
    for candidate in range(100000):
        if _check_hash(nonce, str(candidate), difficulty):
            return str(candidate)
    raise RuntimeError("Could not find solution in 100k attempts")


def test_check_hash_valid():
    nonce = 'testnonce'
    solution = _find_valid_solution(nonce)
    assert _check_hash(nonce, solution, DIFFICULTY) is True


def test_check_hash_invalid():
    # A hash of all-ones bytes can't have 5 leading zero bits — use a known-bad value
    # Instead, verify that a wrong solution fails
    nonce = 'testnonce'
    valid = _find_valid_solution(nonce)
    # Modify the valid solution slightly to make it invalid
    bad = str(int(valid) + 1)
    # This MIGHT accidentally pass — so we use a purposely bad value
    digest = hashlib.sha256(f"{nonce}INTENTIONALLY_WRONG".encode()).digest()
    first_bits = (digest[0] >> 3) & 0x1F  # top 5 bits
    if first_bits != 0:
        assert not _check_hash(nonce, 'INTENTIONALLY_WRONG', DIFFICULTY)


def test_issue_challenge_returns_nonce_and_difficulty():
    challenge = issue_challenge()
    assert 'nonce' in challenge
    assert 'difficulty' in challenge
    assert challenge['difficulty'] == DIFFICULTY
    assert len(challenge['nonce']) == 32  # secrets.token_hex(16)


def test_verify_solution_valid():
    challenge = issue_challenge()
    nonce = challenge['nonce']
    solution = _find_valid_solution(nonce)
    assert verify_solution(nonce, solution) is True


def test_verify_solution_wrong_answer():
    challenge = issue_challenge()
    nonce = challenge['nonce']
    assert verify_solution(nonce, 'definitely_wrong_solution') is False


def test_verify_solution_single_use():
    challenge = issue_challenge()
    nonce = challenge['nonce']
    solution = _find_valid_solution(nonce)
    assert verify_solution(nonce, solution) is True
    # Second attempt with same nonce must fail (challenge consumed)
    assert verify_solution(nonce, solution) is False


def test_verify_solution_unknown_nonce():
    assert verify_solution('nonexistent-nonce-xyz', '0') is False


# ── Projects views ────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_projects_list_requires_pow(client):
    """Project list should require a PoW token (redirect or challenge page)."""
    response = client.get('/projects/')
    # Without PoW session token, should return 200 with challenge JS, not project stories
    assert response.status_code == 200
    content = response.content.decode()
    # Should render the PoW challenge, not actual project content
    assert 'pow' in content.lower() or 'challenge' in content.lower() or 'sha' in content.lower()


@pytest.mark.django_db
def test_pow_challenge_endpoint(client):
    response = client.get('/api/pow/challenge/')
    assert response.status_code == 200
    data = response.json()
    assert 'nonce' in data
    assert 'difficulty' in data


@pytest.mark.django_db
def test_pow_verify_endpoint_invalid(client):
    challenge = client.get('/api/pow/challenge/').json()
    response = client.post(
        '/api/pow/verify/',
        data='{"nonce": "' + challenge['nonce'] + '", "solution": "bad"}',
        content_type='application/json',
    )
    assert response.status_code == 400
    assert response.json()['valid'] is False


@pytest.mark.django_db
def test_pow_verify_endpoint_valid(client):
    challenge = client.get('/api/pow/challenge/').json()
    nonce = challenge['nonce']
    solution = _find_valid_solution(nonce)
    response = client.post(
        '/api/pow/verify/',
        data=f'{{"nonce": "{nonce}", "solution": "{solution}"}}',
        content_type='application/json',
    )
    assert response.status_code == 200
    assert response.json()['valid'] is True


@pytest.mark.django_db
def test_projects_accessible_with_pow_session(client):
    """After solving PoW, project list should return stories."""
    challenge = client.get('/api/pow/challenge/').json()
    nonce = challenge['nonce']
    solution = _find_valid_solution(nonce)
    client.post(
        '/api/pow/verify/',
        data=f'{{"nonce": "{nonce}", "solution": "{solution}"}}',
        content_type='application/json',
    )
    response = client.get('/projects/')
    assert response.status_code == 200


@pytest.mark.django_db
def test_projects_page_never_404(client):
    """Arbitrary page numbers should never 404."""
    # Bypass PoW via session manipulation
    session = client.session
    session['pow_token'] = 'test:token'
    session.save()

    for page in [1, 5, 99, 999, 9999]:
        response = client.get(f'/projects/?page={page}')
        assert response.status_code == 200
