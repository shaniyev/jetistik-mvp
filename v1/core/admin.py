from django.contrib import admin
from django.utils.html import format_html
from .models import Event, Template, ImportBatch, ParticipantRow, Certificate, AuditLog, UserProfile, TeacherStudent, Organization, OrganizationUser

@admin.register(Event)
class EventAdmin(admin.ModelAdmin):
    list_display = ("id", "title", "organization", "date", "city", "created_at")
    search_fields = ("title", "city", "organization__name")

@admin.register(Template)
class TemplateAdmin(admin.ModelAdmin):
    list_display = ("id", "event", "pptx_file", "created_at")
    search_fields = ("event__title",)

@admin.register(ImportBatch)
class ImportBatchAdmin(admin.ModelAdmin):
    list_display = ("id", "event", "status", "rows_total", "rows_ok", "rows_failed", "created_at")
    list_filter = ("status", "event")
    search_fields = ("event__title",)

@admin.register(ParticipantRow)
class ParticipantRowAdmin(admin.ModelAdmin):
    list_display = ("id", "batch", "iin", "name", "status")
    list_filter = ("status", "batch__event")
    search_fields = ("iin", "name")

@admin.action(description="Revoke selected certificates")
def revoke(modeladmin, request, queryset):
    queryset.update(status="revoked", revoked_reason="Revoked by admin")
    AuditLog.objects.create(actor=request.user, action="revoke", object_type="Certificate", meta={"count": queryset.count()})

@admin.action(description="Unrevoke selected certificates")
def unrevoke(modeladmin, request, queryset):
    queryset.update(status="valid", revoked_reason="")
    AuditLog.objects.create(actor=request.user, action="unrevoke", object_type="Certificate", meta={"count": queryset.count()})

@admin.register(Certificate)
class CertificateAdmin(admin.ModelAdmin):
    list_display = ("id", "event", "organization", "iin", "name", "status", "created_at", "pdf_link")
    list_filter = ("status", "event", "organization")
    search_fields = ("iin", "name", "code")
    actions = [revoke, unrevoke]

    def pdf_link(self, obj):
        return format_html('<a href="{}" target="_blank">PDF</a>', obj.pdf.url)
    pdf_link.short_description = "PDF"

@admin.register(AuditLog)
class AuditLogAdmin(admin.ModelAdmin):
    list_display = ("id", "created_at", "actor", "action", "object_type")
    list_filter = ("action", "object_type", "actor")
    search_fields = ("action", "object_type", "object_id")

@admin.register(Organization)
class OrganizationAdmin(admin.ModelAdmin):
    list_display = ("id", "name", "created_at")
    search_fields = ("name",)

@admin.register(OrganizationUser)
class OrganizationUserAdmin(admin.ModelAdmin):
    list_display = ("id", "organization", "user", "created_at")
    search_fields = ("organization__name", "user__username")

@admin.register(UserProfile)
class UserProfileAdmin(admin.ModelAdmin):
    list_display = ("id", "user", "iin", "created_at")
    search_fields = ("user__username", "iin")

@admin.register(TeacherStudent)
class TeacherStudentAdmin(admin.ModelAdmin):
    list_display = ("id", "teacher", "student_iin", "created_at")
    search_fields = ("teacher__username", "student_iin")
