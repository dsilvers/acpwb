from django.urls import path
from . import dashboard_views as views

urlpatterns = [
    path('',          views.overview,       name='dashboard-overview'),
    path('crawlers/', views.crawlers,       name='dashboard-crawlers'),
    path('honeypot/', views.honeypot_traps, name='dashboard-honeypot'),
    path('archive/',  views.archive,        name='dashboard-archive'),
    path('emails/',           views.emails,        name='dashboard-emails'),
    path('emails/<int:pk>/',  views.email_detail,  name='dashboard-email-detail'),
    path('people/',   views.people,         name='dashboard-people'),
    path('logins/',   views.logins,         name='dashboard-logins'),
    path('careers/',  views.careers_applications, name='dashboard-careers'),
    path('conference/', views.conference_registrations, name='dashboard-conference'),
    path('conference/<int:pk>/', views.conference_registration_detail, name='dashboard-conference-detail'),
    path('careers/<int:pk>/', views.career_application_detail, name='dashboard-career-detail'),
    path('careers/<int:pk>/resume/', views.career_download_resume, name='dashboard-career-resume'),
    path('careers/doc/<int:doc_pk>/', views.career_download_document, name='dashboard-career-doc'),
    path('live/',                              views.live_stream,    name='dashboard-live'),
    path('voicemails/',                        views.voicemails,     name='dashboard-voicemails'),
    path('voicemails/audio/<str:recording_sid>/', views.voicemail_audio, name='dashboard-voicemail-audio'),
]
