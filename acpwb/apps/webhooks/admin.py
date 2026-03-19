from django.contrib import admin
from .models import InboundEmail, HoneypotMatch


class HoneypotMatchInline(admin.TabularInline):
    model = HoneypotMatch
    extra = 0
    readonly_fields = ('matched_at', 'match_confidence', 'generated_employee', 'original_visit', 'notes')
    can_delete = False


@admin.register(InboundEmail)
class InboundEmailAdmin(admin.ModelAdmin):
    list_display = ('received_at', 'sender', 'recipient', 'subject_short', 'match_count')
    list_filter = ('received_at',)
    search_fields = ('sender', 'recipient', 'subject')
    readonly_fields = ('received_at', 'mailgun_message_id')
    inlines = [HoneypotMatchInline]

    @admin.display(description='Subject')
    def subject_short(self, obj):
        return obj.subject[:60] if obj.subject else '(no subject)'

    @admin.display(description='Matches')
    def match_count(self, obj):
        return obj.matches.count()


@admin.register(HoneypotMatch)
class HoneypotMatchAdmin(admin.ModelAdmin):
    list_display = ('matched_at', 'match_confidence', 'inbound_sender', 'recipient_email', 'matched_employee')
    list_filter = ('match_confidence', 'matched_at')
    search_fields = ('inbound_email__sender', 'inbound_email__recipient', 'generated_employee__email')
    readonly_fields = ('matched_at',)

    @admin.display(description='From')
    def inbound_sender(self, obj):
        return obj.inbound_email.sender

    @admin.display(description='To')
    def recipient_email(self, obj):
        return obj.inbound_email.recipient

    @admin.display(description='Matched Employee')
    def matched_employee(self, obj):
        if obj.generated_employee:
            return str(obj.generated_employee)
        return '—'
