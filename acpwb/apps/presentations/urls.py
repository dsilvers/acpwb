from django.urls import path
from . import views

urlpatterns = [
    path('', views.presentation_landing, name='presentation-landing'),
    path('<slug:org_slug>/', views.org_page, name='presentation-org'),
    path(
        '<slug:org_slug>/<int:year>/<int:month>/<int:day>/<slug:slug>/',
        views.presentation_detail,
        name='presentation-detail',
    ),
    path(
        '<slug:org_slug>/<int:year>/<int:month>/<int:day>/<slug:slug>/<int:slide_num>/',
        views.presentation_slide,
        name='presentation-slide',
    ),
    path(
        '<slug:org_slug>/<int:year>/<int:month>/<int:day>/<slug:slug>/present/<int:slide_num>/',
        views.presentation_present,
        name='presentation-present',
    ),
    path(
        '<slug:org_slug>/<int:year>/<int:month>/<int:day>/<slug:slug>/download.pdf',
        views.presentation_download_pdf,
        name='presentation-download-pdf',
    ),
    path(
        '<slug:org_slug>/<int:year>/<int:month>/<int:day>/<slug:slug>/download.pptx',
        views.presentation_download_pptx,
        name='presentation-download-pptx',
    ),
]
