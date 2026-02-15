from django.conf import settings
from django.db import migrations, models
import django.db.models.deletion


class Migration(migrations.Migration):

    dependencies = [
        ("core", "0002_add_tokens_json"),
    ]

    operations = [
        migrations.CreateModel(
            name="UserProfile",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("iin", models.CharField(blank=True, db_index=True, default="", max_length=12)),
                ("created_at", models.DateTimeField(auto_now_add=True)),
                ("user", models.OneToOneField(on_delete=django.db.models.deletion.CASCADE, related_name="profile", to=settings.AUTH_USER_MODEL)),
            ],
        ),
        migrations.CreateModel(
            name="TeacherStudent",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("student_iin", models.CharField(db_index=True, max_length=12)),
                ("created_at", models.DateTimeField(auto_now_add=True)),
                ("teacher", models.ForeignKey(on_delete=django.db.models.deletion.CASCADE, related_name="student_links", to=settings.AUTH_USER_MODEL)),
            ],
            options={
                "unique_together": {("teacher", "student_iin")},
            },
        ),
    ]

