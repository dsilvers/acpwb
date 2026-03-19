from django.shortcuts import render, get_object_or_404
from django.http import JsonResponse
from django.views.decorators.http import require_GET, require_POST
from django.views.decorators.csrf import csrf_exempt
import json

from .models import ProjectStory, ProjectPageVisit
from .generators import generate_project_stories
from .pow import issue_challenge, verify_solution

STORIES_PER_PAGE = 10


def _get_ip(request):
    x_forwarded = request.META.get('HTTP_X_FORWARDED_FOR')
    if x_forwarded:
        return x_forwarded.split(',')[0].strip()
    return request.META.get('REMOTE_ADDR', '0.0.0.0')


def _ensure_stories_for_page(page):
    """Generate and save stories for a page if they don't exist yet."""
    existing = ProjectStory.objects.filter(page_number=page).count()
    if existing >= STORIES_PER_PAGE:
        return ProjectStory.objects.filter(page_number=page).order_by('id')

    story_data = generate_project_stories(page=page, count=STORIES_PER_PAGE)
    to_create = []
    existing_slugs = set(ProjectStory.objects.filter(
        slug__in=[s['slug'] for s in story_data]
    ).values_list('slug', flat=True))

    for s in story_data:
        if s['slug'] not in existing_slugs:
            to_create.append(ProjectStory(**s))

    if to_create:
        ProjectStory.objects.bulk_create(to_create, ignore_conflicts=True)

    return ProjectStory.objects.filter(page_number=page).order_by('id')


def project_list(request):
    page = max(1, int(request.GET.get('page', 1)))
    pow_token = request.session.get('pow_token', '')

    stories = _ensure_stories_for_page(page)

    ProjectPageVisit.objects.create(
        ip_address=_get_ip(request),
        user_agent=request.META.get('HTTP_USER_AGENT', '')[:512],
        page_number=page,
        pow_token=pow_token[:128],
    )

    return render(request, 'projects/list.html', {
        'stories': stories,
        'page': page,
        'next_page': page + 1,
        'prev_page': page - 1 if page > 1 else None,
    })


def project_detail(request, slug):
    story = ProjectStory.objects.filter(slug=slug).first()
    if not story:
        # Generate a story deterministically from the slug
        from .generators import _rng_from_seed, ADJECTIVES, NOUNS, INDUSTRIES, CITIES, ORGANIZATIONS, METRICS, PARAGRAPH_TEMPLATES
        import random
        rng = _rng_from_seed(slug)
        org = rng.choice(ORGANIZATIONS)
        adj = rng.choice(ADJECTIVES)
        noun = rng.choice(NOUNS)
        industry = rng.choice(INDUSTRIES)
        city = rng.choice(CITIES)
        metric_template = rng.choice(METRICS)
        impact_metric = metric_template.format(
            n=rng.randint(15, 340), m=rng.randint(2, 850), x=rng.randint(2, 12)
        )
        paragraphs = []
        for _ in range(rng.randint(4, 7)):
            tmpl = rng.choice(PARAGRAPH_TEMPLATES)
            paragraphs.append(tmpl.format(
                org=org, adj=adj.lower(), noun=noun.lower(), industry=industry,
                city=city, months=rng.randint(6, 24), years=rng.randint(5, 18),
                weeks=rng.randint(4, 12), n=rng.randint(12, 97),
                m=rng.randint(1, 500), x=rng.randint(2, 10),
                regions=rng.randint(3, 12),
            ))
        story = ProjectStory(
            slug=slug,
            title=f"{adj} {noun}: {industry} Excellence in {city}",
            summary=f"ACPWB partnered with {org} to deliver a {adj.lower()} {noun.lower()} serving the {industry} sector in {city}. {impact_metric}.",
            body_paragraphs=paragraphs,
            impact_metric=impact_metric,
            industry_tag=industry,
            page_number=0,
        )
        story.save()

    # Generate related project links
    from .generators import _rng_from_seed
    rng = _rng_from_seed(f"related_{slug}")
    related = list(ProjectStory.objects.exclude(slug=slug).order_by('?')[:3])

    return render(request, 'projects/detail.html', {
        'story': story,
        'related': related,
    })


@require_GET
def pow_challenge(request):
    challenge = issue_challenge()
    return JsonResponse(challenge)


@csrf_exempt
@require_POST
def pow_verify(request):
    try:
        data = json.loads(request.body)
        nonce = data.get('nonce', '')
        solution = data.get('solution', '')
    except (json.JSONDecodeError, AttributeError):
        return JsonResponse({'valid': False}, status=400)

    if verify_solution(nonce, solution):
        request.session['pow_token'] = f"{nonce}:{solution}"
        return JsonResponse({'valid': True})
    return JsonResponse({'valid': False}, status=400)
