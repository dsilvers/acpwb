from django.urls import path
from . import views

urlpatterns = [
    path('mailgun/inbound/',      views.mailgun_inbound,     name='mailgun-inbound'),
    path('pipe/inbound/',         views.pipe_inbound,        name='pipe-inbound'),
    path('twilio/recording/',     views.twilio_recording,    name='twilio-recording'),
    path('twilio/transcription/', views.twilio_transcription, name='twilio-transcription'),
]
