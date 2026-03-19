from django.shortcuts import render
from .models import Fortune500Company
from apps.projects.models import ProjectStory


def home(request):
    recent_projects = list(ProjectStory.objects.order_by('?')[:3])
    return render(request, 'public/home.html', {'recent_projects': recent_projects})


def careers(request):
    return render(request, 'public/careers.html')


def mission(request):
    return render(request, 'public/mission.html')


def partners(request):
    companies = list(Fortune500Company.objects.order_by('?')[:40])
    return render(request, 'public/partners.html', {'companies': companies})


def privacy(request):
    return render(request, 'public/privacy.html')
