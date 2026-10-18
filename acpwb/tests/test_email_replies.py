from email.message import EmailMessage as MIMEMessage

import pytest
from django.contrib.auth.models import User
from django.core import mail

from apps.webhooks.models import InboundEmail, OutboundEmail


@pytest.fixture
def staff_client(client, db):
    user = User.objects.create_user('staff', password='pass', is_staff=True)
    client.force_login(user)
    return client


def _raw(reply_to=None):
    msg = MIMEMessage()
    msg['From'] = '"Studenec, Patrick J" <patrick@example.gov>'
    msg['To'] = '"General Inquiries" <general-inquiry@acpwb.com>, inquiry@acpwb.com'
    msg['Subject'] = 'Complaint 185872'
    msg['Date'] = 'Wed, 14 Oct 2026 15:17:01 +0000'
    msg['Message-ID'] = '<orig-1@example.gov>'
    msg['References'] = '<older@example.gov>'
    if reply_to:
        msg['Reply-To'] = reply_to
    msg.set_content('Please respond.\n\nThanks')
    msg.add_attachment(b'%PDF-1.4 fake', maintype='application', subtype='pdf', filename='complaint.pdf')
    msg.add_attachment(b'<script>x</script>', maintype='text', subtype='html', filename='evil.html')
    return msg.as_string()


@pytest.fixture
def inbound(db):
    return InboundEmail.objects.create(
        sender='prvs=1234=patrick@example.gov',
        recipient='general-inquiry@acpwb.com',
        subject='Complaint 185872',
        body_plain='Please respond.\n\nThanks',
        raw_payload={'raw': _raw()},
        message_id='<orig-1@example.gov>',
    )


@pytest.mark.django_db
def test_detail_shows_attachments_and_reply_defaults(staff_client, inbound):
    r = staff_client.get(f'/acpwb-dashboard/emails/{inbound.pk}/')
    assert r.status_code == 200
    html = r.content.decode()
    assert 'complaint.pdf' in html and 'evil.html' in html
    reply = r.context['reply']
    assert reply['from_name'] == 'General Inquiries'
    assert reply['from_local'] == 'general-inquiry'
    assert reply['to'] == '"Studenec, Patrick J" <patrick@example.gov>'  # header, not the prvs= envelope
    assert reply['subject'] == 'Re: Complaint 185872'
    assert '> Please respond.' in reply['body']


@pytest.mark.django_db
def test_reply_to_header_wins(staff_client, db):
    e = InboundEmail.objects.create(
        sender='a@example.com', recipient='jane.doe@acpwb.com', subject='Re: hi',
        raw_payload={'raw': _raw(reply_to='desk@example.gov')},
    )
    reply = staff_client.get(f'/acpwb-dashboard/emails/{e.pk}/').context['reply']
    assert reply['to'] == 'desk@example.gov'
    assert reply['subject'] == 'Re: hi'
    assert reply['from_local'] == 'jane.doe'


@pytest.mark.django_db
def test_attachment_download(staff_client, inbound):
    r = staff_client.get(f'/acpwb-dashboard/emails/{inbound.pk}/attachments/0/?inline=1')
    assert r.content == b'%PDF-1.4 fake'
    assert r['Content-Type'] == 'application/pdf'
    assert r['Content-Disposition'].startswith('inline')

    # HTML never renders on the dashboard origin, even when asked
    r = staff_client.get(f'/acpwb-dashboard/emails/{inbound.pk}/attachments/1/?inline=1')
    assert r['Content-Type'] == 'application/octet-stream'
    assert r['Content-Disposition'].startswith('attachment')
    assert r['Content-Security-Policy'] == 'sandbox'

    assert staff_client.get(f'/acpwb-dashboard/emails/{inbound.pk}/attachments/9/').status_code == 404


@pytest.mark.django_db
def test_attachment_requires_staff(client, inbound):
    r = client.get(f'/acpwb-dashboard/emails/{inbound.pk}/attachments/0/')
    assert r.status_code == 302


@pytest.mark.django_db
def test_send_reply(staff_client, inbound):
    r = staff_client.post(f'/acpwb-dashboard/emails/{inbound.pk}/', {
        'from_name': 'ACPWB Inquiries',
        'from_local': 'press',
        'to': 'Patrick <patrick@example.gov>',
        'cc': 'other@example.com',
        'subject': 'Re: Complaint 185872',
        'body': 'Hello',
    })
    assert r.status_code == 302
    assert len(mail.outbox) == 1
    sent = mail.outbox[0]
    assert sent.from_email == 'ACPWB Inquiries <press@acpwb.com>'
    assert sent.to == ['Patrick <patrick@example.gov>']
    assert sent.cc == ['other@example.com']
    assert sent.extra_headers['In-Reply-To'] == '<orig-1@example.gov>'
    assert sent.extra_headers['References'] == '<older@example.gov> <orig-1@example.gov>'
    assert sent.extra_headers['Message-ID'].endswith('@acpwb.com>')
    out = OutboundEmail.objects.get()
    assert out.in_reply_to == inbound and out.sent_by == 'staff'


@pytest.mark.django_db
@pytest.mark.parametrize('from_local', ['someone@gmail.com', 'bad local', '', 'x\r\nBcc: a@b.c'])
def test_from_must_stay_on_acpwb(staff_client, inbound, from_local):
    r = staff_client.post(f'/acpwb-dashboard/emails/{inbound.pk}/', {
        'from_local': from_local, 'to': 'patrick@example.gov', 'subject': 's', 'body': 'b',
    })
    assert r.status_code == 200
    assert r.context['reply_error'].startswith('From:')
    assert mail.outbox == []
    assert not OutboundEmail.objects.exists()


@pytest.mark.django_db
def test_invalid_to_rejected(staff_client, inbound):
    r = staff_client.post(f'/acpwb-dashboard/emails/{inbound.pk}/', {
        'from_local': 'press', 'to': 'not-an-address', 'subject': 's', 'body': 'b',
    })
    assert r.context['reply_error'].startswith('To:')
    assert mail.outbox == []
