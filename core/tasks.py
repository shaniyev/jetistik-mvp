from __future__ import annotations
import os
import tempfile
import uuid
from django.conf import settings
from django.core.files.base import ContentFile
from django.db import transaction
from pptx import Presentation

try:
    from celery import shared_task
except ModuleNotFoundError:
    def shared_task(*args, **kwargs):
        def decorator(func):
            func.delay = func
            return func
        return decorator

from .models import ImportBatch, ParticipantRow, Certificate, Template, AuditLog
from .utils import read_table, default_mapping, generate_certificate_pdf

@shared_task(bind=True)
def generate_batch(self, batch_id: int):
    batch = ImportBatch.objects.select_related("event").get(pk=batch_id)
    event = batch.event

    template = Template.objects.filter(event=event).order_by("-created_at").first()
    if not template:
        batch.status = "failed"
        batch.report_json = {"error": "Template not found"}
        batch.save(update_fields=["status", "report_json"])
        return

    batch.status = "generating"
    batch.rows_ok = 0
    batch.rows_failed = 0
    batch.save(update_fields=["status", "rows_ok", "rows_failed"])

    mapping = batch.mapping_json or {}
    # sanity mapping: token->column
    # required for iin/name/id as well
    # (iin used for access even if not in tokens)
    rows = batch.rows.all().order_by("id")

    for row in rows:
        try:
            payload = row.payload_json or {}
            # Prepare token values
            token_to_value = {}
            for token, col in mapping.items():
                if not col:
                    continue
                token_to_value[token] = payload.get(col, "")

            # Always include fid fallback if not mapped
            token_to_value.setdefault("fname", row.name)

            code = uuid.uuid4().hex
            verify_url = settings.PUBLIC_BASE_URL.rstrip("/") + f"/verify/{code}/"
            pdf_bytes = generate_certificate_pdf(template.pptx_file.path, token_to_value, verify_url)

            cert_payload = dict(payload)
            cert_payload.update({
                "event_title": event.title,
                "event_date": str(event.date) if event.date else "",
                "event_city": event.city,
                "organization_name": event.organization.name if event.organization else "",
            })

            cert = Certificate(
                event=event,
                organization=event.organization,
                iin=row.iin,
                name=row.name,
                code=code,
                status="valid",
                payload_json=cert_payload,
            )
            cert.pdf.save(f"{code}.pdf", ContentFile(pdf_bytes), save=True)

            row.status = "ok"
            row.error = ""
            row.save(update_fields=["status", "error"])

            batch.rows_ok += 1
            batch.save(update_fields=["rows_ok"])

        except Exception as e:
            row.status = "failed"
            row.error = str(e)[:2000]
            row.save(update_fields=["status", "error"])
            batch.rows_failed += 1
            batch.save(update_fields=["rows_failed"])

    batch.status = "done_with_errors" if batch.rows_failed else "done"
    batch.save(update_fields=["status"])
    AuditLog.objects.create(action="generate", object_type="ImportBatch", object_id=str(batch.pk), meta={"ok": batch.rows_ok, "failed": batch.rows_failed})
    return {"ok": batch.rows_ok, "failed": batch.rows_failed}
