from __future__ import annotations
import os
from django.contrib import messages
from django.contrib.admin.views.decorators import staff_member_required
from django.db import transaction
from django.http import FileResponse, Http404, HttpResponseForbidden
from django.shortcuts import get_object_or_404, redirect, render
from django.urls import reverse

from ratelimit.decorators import ratelimit

from .forms import IINSearchForm, EventForm, TemplateUploadForm, ImportUploadForm, MappingForm
from .models import Event, Template, ImportBatch, ParticipantRow, Certificate, AuditLog
from .utils import read_table, default_mapping, mask_iin, validate_required
from .tasks import generate_batch

TOKENS = ["fname", "fschool", "fclass", "fplace", "fteacher", "fnomination", "fid"]

@ratelimit(key="ip", rate="10/m", block=True)
def home(request):
    form = IINSearchForm(request.POST or None)
    if request.method == "POST":
        if form.is_valid():
            request.session["iin"] = form.cleaned_data["iin"]
            return redirect("my_certificates")
    return render(request, "public/home.html", {"form": form})

def my_certificates(request):
    iin = request.session.get("iin", "")
    if not iin:
        return redirect("home")
    certs = Certificate.objects.filter(iin=iin).select_related("event").order_by("-created_at")
    return render(request, "public/my.html", {"iin": iin, "certificates": certs})

def download_certificate(request, code: str):
    iin = request.session.get("iin", "")
    if not iin:
        return HttpResponseForbidden("Сначала введите ИИН.")
    cert = get_object_or_404(Certificate, code=code)
    if cert.iin != iin:
        return HttpResponseForbidden("Нет доступа к этому сертификату.")
    if not cert.pdf:
        raise Http404()
    return FileResponse(cert.pdf.open("rb"), as_attachment=True, filename=os.path.basename(cert.pdf.name))

def verify(request, code: str):
    cert = Certificate.objects.filter(code=code).select_related("event").first()
    if not cert:
        return render(request, "public/verify.html", {"status": "NOT_FOUND", "cert": None})

    payload = cert.payload_json or {}
    ctx = {
        "status": "REVOKED" if cert.status == "revoked" else "VALID",
        "cert": cert,
        "payload": payload,
        "masked_iin": mask_iin(cert.iin),
    }
    return render(request, "public/verify.html", ctx)

# ----------------------- Staff MVP Dashboard -----------------------

@staff_member_required
def staff_events(request):
    events = Event.objects.order_by("-created_at")
    if request.method == "POST":
        form = EventForm(request.POST)
        if form.is_valid():
            ev = form.save()
            AuditLog.objects.create(actor=request.user, action="create_event", object_type="Event", object_id=str(ev.pk))
            return redirect("staff_events")
    else:
        form = EventForm()
    return render(request, "staff/events.html", {"events": events, "form": form})

@staff_member_required
def staff_event_detail(request, event_id: int):
    event = get_object_or_404(Event, pk=event_id)
    latest_template = Template.objects.filter(event=event).order_by("-created_at").first()
    batches = ImportBatch.objects.filter(event=event).order_by("-created_at")[:20]

    if request.method == "POST":
        if "upload_template" in request.POST:
            tform = TemplateUploadForm(request.POST, request.FILES)
            if tform.is_valid():
                tpl = tform.save(commit=False)
                tpl.event = event
                tpl.save()
                AuditLog.objects.create(actor=request.user, action="upload_template", object_type="Template", object_id=str(tpl.pk), meta={"event_id": event.pk})
                messages.success(request, "Шаблон загружен.")
                return redirect("staff_event_detail", event_id=event.pk)
        elif "upload_data" in request.POST:
            iform = ImportUploadForm(request.POST, request.FILES)
            if iform.is_valid():
                batch = iform.save(commit=False)
                batch.event = event
                batch.status = "uploaded"
                batch.save()
                AuditLog.objects.create(actor=request.user, action="upload_data", object_type="ImportBatch", object_id=str(batch.pk), meta={"event_id": event.pk})

                # Parse file and create ParticipantRows (MVP)
                file_path = batch.file.path
                cols, rows = read_table(file_path)
                batch.rows_total = len(rows)
                missing = validate_required(cols)
                report = {"columns": cols, "missing_required": missing, "preview": rows[:10]}
                batch.report_json = report
                batch.mapping_json = default_mapping(cols)
                batch.status = "validated" if not missing else "uploaded"
                batch.save(update_fields=["rows_total", "report_json", "mapping_json", "status"])

                # Create rows
                with transaction.atomic():
                    ParticipantRow.objects.filter(batch=batch).delete()
                    for r in rows:
                        iin = (r.get("iin") or "").strip()
                        name = (r.get("name") or "").strip()
                        ParticipantRow.objects.create(
                            batch=batch,
                            iin=iin,
                            name=name,
                            payload_json=r,
                            status="pending" if (iin and name) else "failed",
                            error="" if (iin and name) else "Missing iin/name",
                        )

                messages.success(request, "Данные загружены. Проверь mapping и запускай Generate.")
                return redirect("staff_batch_mapping", batch_id=batch.pk)

    else:
        tform = TemplateUploadForm()
        iform = ImportUploadForm()

    return render(request, "staff/event_detail.html", {
        "event": event,
        "latest_template": latest_template,
        "batches": batches,
        "tform": tform,
        "iform": iform,
    })

@staff_member_required
def staff_batch_mapping(request, batch_id: int):
    batch = get_object_or_404(ImportBatch, pk=batch_id)
    report = batch.report_json or {}
    columns = report.get("columns", [])
    mapping = batch.mapping_json or {}

    class _DynMappingForm(MappingForm):
        pass

    for tok in TOKENS:
        _DynMappingForm.base_fields[tok] = __import__("django").forms.ChoiceField(
            label=f"{tok} → колонка",
            choices=[("", "— не вставлять —")] + [(c, c) for c in columns],
            required=False,
            initial=mapping.get(tok, ""),
        )

    if request.method == "POST":
        form = _DynMappingForm(request.POST)
        if form.is_valid():
            batch.mapping_json = {tok: form.cleaned_data.get(tok, "") for tok in TOKENS}
            batch.save(update_fields=["mapping_json"])
            AuditLog.objects.create(actor=request.user, action="mapping_update", object_type="ImportBatch", object_id=str(batch.pk))
            messages.success(request, "Mapping сохранён.")
            if "save_and_generate" in request.POST:
                return redirect("staff_batch_generate", batch_id=batch.pk)
    else:
        form = _DynMappingForm()

    return render(request, "staff/mapping.html", {"batch": batch, "form": form, "report": report})

@staff_member_required
def staff_batch_generate(request, batch_id: int):
    batch = get_object_or_404(ImportBatch, pk=batch_id)
    if request.method == "POST":
        # start celery task
        batch.status = "generating"
        batch.save(update_fields=["status"])
        generate_batch.delay(batch.pk)
        messages.success(request, "Генерация запущена. Обновляй страницу для прогресса.")
        return redirect("staff_batch_generate", batch_id=batch.pk)

    # progress
    total = batch.rows_total or 0
    ok = batch.rows_ok or 0
    failed = batch.rows_failed or 0
    return render(request, "staff/generate.html", {"batch": batch, "total": total, "ok": ok, "failed": failed})
