from django.conf import settings
from django.db import migrations, models
import django.db.models.deletion


class Migration(migrations.Migration):

    initial = True

    dependencies = [
        migrations.swappable_dependency(settings.AUTH_USER_MODEL),
    ]

    operations = [
        migrations.CreateModel(
            name="Event",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("title", models.CharField(max_length=255)),
                ("date", models.DateField(blank=True, null=True)),
                ("city", models.CharField(blank=True, default="", max_length=128)),
                ("description", models.TextField(blank=True, default="")),
                ("created_at", models.DateTimeField(auto_now_add=True)),
            ],
        ),
        migrations.CreateModel(
            name="ImportBatch",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("file", models.FileField(upload_to="imports/%Y/%m/")),
                ("status", models.CharField(choices=[("uploaded", "Uploaded"), ("validated", "Validated"), ("generating", "Generating"), ("done", "Done"), ("done_with_errors", "Done with errors"), ("failed", "Failed")], default="uploaded", max_length=32)),
                ("rows_total", models.PositiveIntegerField(default=0)),
                ("rows_ok", models.PositiveIntegerField(default=0)),
                ("rows_failed", models.PositiveIntegerField(default=0)),
                ("mapping_json", models.JSONField(blank=True, default=dict)),
                ("tokens_json", models.JSONField(blank=True, default=list)),
                ("report_json", models.JSONField(blank=True, default=dict)),
                ("created_at", models.DateTimeField(auto_now_add=True)),
                ("event", models.ForeignKey(on_delete=django.db.models.deletion.CASCADE, related_name="batches", to="core.event")),
            ],
        ),
        migrations.CreateModel(
            name="ParticipantRow",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("iin", models.CharField(db_index=True, max_length=12)),
                ("name", models.CharField(max_length=255)),
                ("payload_json", models.JSONField(blank=True, default=dict)),
                ("status", models.CharField(choices=[("pending", "Pending"), ("ok", "OK"), ("failed", "Failed"), ("skipped", "Skipped")], default="pending", max_length=16)),
                ("error", models.TextField(blank=True, default="")),
                ("batch", models.ForeignKey(on_delete=django.db.models.deletion.CASCADE, related_name="rows", to="core.importbatch")),
            ],
        ),
        migrations.CreateModel(
            name="Template",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("pptx_file", models.FileField(upload_to="templates/%Y/%m/")),
                ("created_at", models.DateTimeField(auto_now_add=True)),
                ("event", models.ForeignKey(on_delete=django.db.models.deletion.CASCADE, related_name="templates", to="core.event")),
            ],
        ),
        migrations.CreateModel(
            name="Certificate",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("iin", models.CharField(db_index=True, max_length=12)),
                ("name", models.CharField(max_length=255)),
                ("code", models.CharField(db_index=True, max_length=64, unique=True)),
                ("pdf", models.FileField(upload_to="certificates/%Y/%m/")),
                ("status", models.CharField(choices=[("valid", "VALID"), ("revoked", "REVOKED")], default="valid", max_length=16)),
                ("revoked_reason", models.CharField(blank=True, default="", max_length=255)),
                ("payload_json", models.JSONField(blank=True, default=dict)),
                ("created_at", models.DateTimeField(auto_now_add=True)),
                ("event", models.ForeignKey(on_delete=django.db.models.deletion.CASCADE, related_name="certificates", to="core.event")),
            ],
        ),
        migrations.CreateModel(
            name="AuditLog",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("action", models.CharField(max_length=64)),
                ("object_type", models.CharField(blank=True, default="", max_length=64)),
                ("object_id", models.CharField(blank=True, default="", max_length=64)),
                ("meta", models.JSONField(blank=True, default=dict)),
                ("created_at", models.DateTimeField(auto_now_add=True)),
                ("actor", models.ForeignKey(blank=True, null=True, on_delete=django.db.models.deletion.SET_NULL, to=settings.AUTH_USER_MODEL)),
            ],
        ),
    ]

