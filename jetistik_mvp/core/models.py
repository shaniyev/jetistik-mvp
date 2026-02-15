from __future__ import annotations
from django.db import models
from django.utils import timezone
from django.contrib.auth import get_user_model

User = get_user_model()

class Event(models.Model):
    title = models.CharField(max_length=255)
    date = models.DateField(null=True, blank=True)
    city = models.CharField(max_length=128, blank=True, default="")
    description = models.TextField(blank=True, default="")
    created_at = models.DateTimeField(auto_now_add=True)

    def __str__(self) -> str:
        return f"{self.title}"

class Template(models.Model):
    event = models.ForeignKey(Event, on_delete=models.CASCADE, related_name="templates")
    pptx_file = models.FileField(upload_to="templates/%Y/%m/")
    created_at = models.DateTimeField(auto_now_add=True)

    def __str__(self) -> str:
        return f"Template #{self.pk} for {self.event_id}"

class ImportBatch(models.Model):
    STATUS_CHOICES = [
        ("uploaded", "Uploaded"),
        ("validated", "Validated"),
        ("generating", "Generating"),
        ("done", "Done"),
        ("done_with_errors", "Done with errors"),
        ("failed", "Failed"),
    ]
    event = models.ForeignKey(Event, on_delete=models.CASCADE, related_name="batches")
    file = models.FileField(upload_to="imports/%Y/%m/")
    status = models.CharField(max_length=32, choices=STATUS_CHOICES, default="uploaded")
    rows_total = models.PositiveIntegerField(default=0)
    rows_ok = models.PositiveIntegerField(default=0)
    rows_failed = models.PositiveIntegerField(default=0)
    mapping_json = models.JSONField(default=dict, blank=True)  # token -> column
    report_json = models.JSONField(default=dict, blank=True)   # errors, etc.
    created_at = models.DateTimeField(auto_now_add=True)

    def __str__(self) -> str:
        return f"Batch #{self.pk} ({self.status})"

class ParticipantRow(models.Model):
    STATUS_CHOICES = [
        ("pending", "Pending"),
        ("ok", "OK"),
        ("failed", "Failed"),
        ("skipped", "Skipped"),
    ]
    batch = models.ForeignKey(ImportBatch, on_delete=models.CASCADE, related_name="rows")
    iin = models.CharField(max_length=12, db_index=True)
    name = models.CharField(max_length=255)
    payload_json = models.JSONField(default=dict, blank=True)
    status = models.CharField(max_length=16, choices=STATUS_CHOICES, default="pending")
    error = models.TextField(blank=True, default="")

    def __str__(self) -> str:
        return f"{self.iin} {self.name}"

class Certificate(models.Model):
    STATUS_CHOICES = [
        ("valid", "VALID"),
        ("revoked", "REVOKED"),
    ]
    event = models.ForeignKey(Event, on_delete=models.CASCADE, related_name="certificates")
    iin = models.CharField(max_length=12, db_index=True)
    name = models.CharField(max_length=255)
    code = models.CharField(max_length=64, unique=True, db_index=True)  # public code (uuid/token)
    pdf = models.FileField(upload_to="certificates/%Y/%m/")
    status = models.CharField(max_length=16, choices=STATUS_CHOICES, default="valid")
    revoked_reason = models.CharField(max_length=255, blank=True, default="")
    payload_json = models.JSONField(default=dict, blank=True)  # snapshot for verify page
    created_at = models.DateTimeField(auto_now_add=True)

    def __str__(self) -> str:
        return f"{self.name} ({self.event_id})"

class AuditLog(models.Model):
    actor = models.ForeignKey(User, null=True, blank=True, on_delete=models.SET_NULL)
    action = models.CharField(max_length=64)  # upload_template/upload_data/generate/revoke/unrevoke
    object_type = models.CharField(max_length=64, blank=True, default="")
    object_id = models.CharField(max_length=64, blank=True, default="")
    meta = models.JSONField(default=dict, blank=True)
    created_at = models.DateTimeField(auto_now_add=True)

    def __str__(self) -> str:
        return f"{self.created_at} {self.action}"
