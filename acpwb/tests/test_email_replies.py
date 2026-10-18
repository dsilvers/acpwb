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
        raw_payload={'raw': _raw(reply_to='desk@example.gov').replace('Subject: Complaint 185872', 'Subject: Re: hi')},
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


@pytest.mark.django_db
def test_encoded_subject_is_decoded(staff_client, db):
    raw = _raw().replace('Subject: Complaint 185872', 'Subject: =?UTF-8?Q?Application_=E2=80=94_Analyst?=')
    e = InboundEmail.objects.create(
        sender='a@example.com', recipient='info@acpwb.com',
        subject='=?UTF-8?Q?Application_=E2=80=94_Analyst?=', raw_payload={'raw': raw},
    )
    reply = staff_client.get(f'/acpwb-dashboard/emails/{e.pk}/').context['reply']
    assert reply['subject'] == 'Re: Application \u2014 Analyst'


@pytest.mark.django_db
def test_page_renders_reply_form_once(staff_client, inbound):
    html = staff_client.get(f'/acpwb-dashboard/emails/{inbound.pk}/').content.decode()
    assert html.count('id="reply"') == 1
    assert html.count('name="from_local"') == 1
    title = html[html.index('<title>'):html.index('</title>')]
    assert '<' not in title[len('<title>'):]


@pytest.mark.django_db
def test_html_only_email_quotes_text_and_drops_echo_name(staff_client, db):
    e = InboundEmail.objects.create(
        sender='lily@example.com', recipient='fermina.salus@acpwb.com', subject='Watches',
        body_html='<html><head><title>T</title><style>p{color:red}</style></head>'
                  '<body><p>Luxury&nbsp;watches</p><p>From $250</p></body></html>',
        raw_payload={'raw': 'From: lily <lily@example.com>\r\nTo: fermina.salus <fermina.salus@acpwb.com>\r\n'
                            'Subject: Watches\r\nContent-Type: text/html\r\n\r\n<p>x</p>\r\n'},
    )
    r = staff_client.get(f'/acpwb-dashboard/emails/{e.pk}/')
    reply = r.context['reply']
    assert reply['from_name'] == ''
    assert '> Luxury watches\n> From $250' in reply['body']
    assert '<' not in reply['body'].split('wrote:')[1]
    assert r.context['body'] == 'Luxury watches\nFrom $250'


# ── Inbox ─────────────────────────────────────────────────────────────────────

@pytest.fixture
def many_emails(db):
    for i in range(60):
        InboundEmail.objects.create(
            sender=f'user{i}@spam{i % 3}.example', recipient=f'person{i}@acpwb.com',
            subject=f'Offer number {i}', body_plain='buy now' if i == 7 else 'hello',
        )


@pytest.mark.django_db
def test_inbox_paginates(staff_client, many_emails):
    r = staff_client.get('/acpwb-dashboard/emails/inbox/')
    assert r.status_code == 200
    page = r.context['page']
    assert page.paginator.count == 60 and len(page.object_list) == 50
    assert page.object_list[0].subject == 'Offer number 59'  # newest first
    r = staff_client.get('/acpwb-dashboard/emails/inbox/?page=2')
    assert len(r.context['page'].object_list) == 10


@pytest.mark.django_db
@pytest.mark.parametrize('q,count', [
    ('user5@spam2.example', 1),   # address
    ('spam1.example', 20),        # domain
    ('number 42', 1),             # subject words
    ('buy now', 1),               # body
    ('person3@acpwb.com', 1),     # recipient
    ('nothing-matches', 0),
])
def test_inbox_search(staff_client, many_emails, q, count):
    r = staff_client.get('/acpwb-dashboard/emails/inbox/', {'q': q})
    assert r.context['page'].paginator.count == count


@pytest.mark.django_db
def test_sidebar_has_email_section(staff_client):
    html = staff_client.get('/acpwb-dashboard/emails/inbox/').content.decode()
    html = html[html.index('<nav'):html.index('</nav>')]
    assert html.index('Crawler Visits') < html.index('TLS Fingerprints') < html.index('Honeypot Traps')
    assert html.index('>Email<') < html.index('Email Inbox') < html.index('Inbound Email') < html.index('Engagement')


def test_decode_mime_header():
    from apps.webhooks.views import decode_mime_header
    assert decode_mime_header('=?UTF-8?Q?Application_=E2=80=94_Analyst?=') == 'Application — Analyst'
    assert decode_mime_header('=?UTF-8?B?Q2Fmw6k=?= menu') == 'Café menu'
    assert decode_mime_header('Plain subject') == 'Plain subject'
    assert decode_mime_header('') == ''


# ── Matches and From display ──────────────────────────────────────────────────

@pytest.mark.django_db
def test_no_match_row_is_not_counted(staff_client, inbound):
    from apps.webhooks.models import HoneypotMatch
    HoneypotMatch.objects.create(inbound_email=inbound, match_confidence='none', notes='nothing')
    r = staff_client.get('/acpwb-dashboard/emails/inbox/')
    assert r.context['page'].object_list[0].match_count == 0
    r = staff_client.get(f'/acpwb-dashboard/emails/{inbound.pk}/')
    assert list(r.context['honeypot_matches']) == []
    assert 'No match' in r.content.decode()
    assert 'id="matches"' not in r.content.decode()


@pytest.mark.django_db
def test_match_details_shown_below_reply(staff_client, inbound):
    from apps.people.models import GeneratedEmployee, PeoplePageVisit
    from apps.webhooks.models import HoneypotMatch
    visit = PeoplePageVisit.objects.create(ip_address='203.0.113.9', user_agent='EvilBot/1.0')
    emp = GeneratedEmployee.objects.create(
        visit=visit, first_name='Gen', last_name='Eral', email='general-inquiry@acpwb.com',
        title='Analyst', department='Research', avatar_seed='x',
    )
    HoneypotMatch.objects.create(inbound_email=inbound, generated_employee=emp, original_visit=visit, match_confidence='exact')
    HoneypotMatch.objects.create(inbound_email=inbound, match_confidence='none')
    r = staff_client.get('/acpwb-dashboard/emails/inbox/')
    assert r.context['page'].object_list[0].match_count == 1
    html = staff_client.get(f'/acpwb-dashboard/emails/{inbound.pk}/').content.decode()
    assert html.index('id="reply"') < html.index('id="matches"')
    assert '203.0.113.9' in html and 'EvilBot/1.0' in html and 'Gen Eral' in html


@pytest.mark.django_db
def test_inbox_shows_from_header_not_envelope(staff_client, db):
    InboundEmail.objects.create(
        sender='bounces+123-abc=acpwb.com@em1.example.jp', from_header='Nostalgia Online <news@example.jp>',
        recipient='x@acpwb.com', subject='Hi',
    )
    html = staff_client.get('/acpwb-dashboard/emails/inbox/').content.decode()
    assert 'Nostalgia Online' in html and 'news@example.jp' in html
    assert staff_client.get('/acpwb-dashboard/emails/inbox/', {'q': 'Nostalgia'}).context['page'].paginator.count == 1


@pytest.mark.django_db
def test_pipe_stores_decoded_from_header(client, settings):
    import json
    settings.PIPE_WEBHOOK_SECRET = 's'
    raw = ('From: =?UTF-8?Q?Jos=C3=A9?= <jose@example.com>\r\nTo: a@acpwb.com\r\nSubject: =?UTF-8?Q?Caf=C3=A9?=\r\n'
           'Message-ID: <f1@example.com>\r\n\r\nhi\r\n')
    client.post('/webhooks/pipe/inbound/', data=json.dumps({
        'sender': 'prvs=1=jose@example.com', 'recipient': 'a@acpwb.com', 'subject': '=?UTF-8?Q?Caf=C3=A9?=', 'raw': raw,
    }), content_type='application/json', HTTP_X_WEBHOOK_SECRET='s')
    e = InboundEmail.objects.get()
    assert e.from_header == 'José <jose@example.com>'
    assert e.from_name == 'José' and e.from_address == 'jose@example.com'
    assert e.subject == 'Café'
