from django.contrib import admin
from .models import PeoplePageVisit, GeneratedEmployee


class GeneratedEmployeeInline(admin.TabularInline):
    model = GeneratedEmployee
    extra = 0
    readonly_fields = ('first_name', 'last_name', 'email', 'title', 'department', 'avatar_seed', 'created_at')
    can_delete = False


@admin.register(PeoplePageVisit)
class PeoplePageVisitAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'ip_address', 'user_agent_short', 'employee_count')
    list_filter = ('timestamp',)
    search_fields = ('ip_address', 'user_agent')
    readonly_fields = ('timestamp',)
    inlines = [GeneratedEmployeeInline]

    @admin.display(description='UA')
    def user_agent_short(self, obj):
        return obj.user_agent[:60] if obj.user_agent else '—'

    @admin.display(description='# Employees')
    def employee_count(self, obj):
        return obj.employees.count()


@admin.register(GeneratedEmployee)
class GeneratedEmployeeAdmin(admin.ModelAdmin):
    list_display = ('full_name', 'email', 'title', 'department', 'created_at', 'visit_ip')
    list_filter = ('department', 'created_at')
    search_fields = ('first_name', 'last_name', 'email', 'title')
    readonly_fields = ('created_at',)

    @admin.display(description='Visit IP')
    def visit_ip(self, obj):
        return obj.visit.ip_address
