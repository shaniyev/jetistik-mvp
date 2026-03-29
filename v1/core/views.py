from __future__ import annotations
import os
import io
import json
import zipfile
from functools import wraps
from django.contrib import messages
from django.contrib.auth.decorators import login_required, user_passes_test
from django.contrib.auth import login
from django.contrib.auth.models import Group
from django.conf import settings
from django.db import transaction
from django.http import FileResponse, Http404, HttpResponseForbidden, HttpResponseNotAllowed, HttpResponse, JsonResponse
from django.shortcuts import get_object_or_404, redirect, render
from django.urls import reverse
from django.utils.text import slugify
from django.core.files.base import ContentFile
from django.core.mail import send_mail
from django.views.decorators.csrf import csrf_exempt

try:
    from django_ratelimit.decorators import ratelimit
except ModuleNotFoundError:
    def ratelimit(*args, **kwargs):
        def decorator(func):
            return func
        return decorator

from .forms import IINSearchForm, EventForm, TemplateUploadForm, ImportUploadForm, MappingForm, BatchEditForm, CertificateEditForm, TeacherAddStudentForm, RegistrationForm, ProfileIINForm, OrganizationRegistrationForm
from .models import Event, Template, ImportBatch, ParticipantRow, Certificate, AuditLog, UserProfile, TeacherStudent, Organization, OrganizationUser
from .utils import read_table, default_mapping, default_mapping_for_tokens, extract_tokens_from_pptx, mask_iin, validate_required, TOKEN_LIST, generate_certificate_pdf
from .tasks import generate_batch

DEFAULT_TOKENS = TOKEN_LIST

@ratelimit(key="ip", rate="10/m", block=True)
def home(request):
    form = IINSearchForm(request.POST or None)
    if request.method == "POST":
        if form.is_valid():
            request.session["iin"] = form.cleaned_data["iin"]
            return redirect("my_certificates")
    return render(request, "public/home.html", {"form": form})

def landing(request):
    return render(request, "public/landing.html")

@csrf_exempt
def organizer_request_submit(request):
    if request.method != "POST":
        return JsonResponse({"ok": False, "error": "method_not_allowed"}, status=405)
    try:
        payload = json.loads(request.body.decode("utf-8")) if request.body else {}
    except Exception:
        payload = request.POST.dict()
    full_name = (payload.get("fullName") or payload.get("full_name") or "").strip()
    organization = (payload.get("organization") or "").strip()
    phone = (payload.get("phone") or "").strip()
    if not full_name or not organization or not phone:
        return JsonResponse({"ok": False, "error": "invalid_payload"}, status=400)
    subject = "Jetistik: новая заявка от организатора"
    body = (
        "Получена новая заявка с формы /organizers/.\n\n"
        f"ФИО: {full_name}\n"
        f"Организация: {organization}\n"
        f"Телефон: {phone}\n"
    )
    try:
        send_mail(
            subject=subject,
            message=body,
            from_email=settings.DEFAULT_FROM_EMAIL,
            recipient_list=[settings.ORGANIZER_REQUEST_TO],
            fail_silently=False,
        )
        return JsonResponse({"ok": True})
    except Exception as exc:
        return JsonResponse({"ok": False, "error": str(exc)}, status=500)

def my_certificates(request):
    iin_from_get = (request.GET.get("iin") or "").strip()
    if iin_from_get:
        form = IINSearchForm({"iin": iin_from_get})
        if form.is_valid():
            request.session["iin"] = form.cleaned_data["iin"]
        else:
            return render(request, "public/home.html", {"form": form})
    iin = request.session.get("iin", "")
    if not iin:
        return redirect("home")
    certs = Certificate.objects.filter(iin=iin).select_related("event").order_by("-created_at")
    return render(request, "public/my.html", {"iin": iin, "certificates": certs})

def my_certificates_download(request):
    iin = request.session.get("iin", "")
    if not iin:
        return HttpResponseForbidden("Сначала введите ИИН.")
    certs = Certificate.objects.filter(iin=iin).select_related("event").order_by("-created_at")
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w", zipfile.ZIP_DEFLATED) as zf:
        for cert in certs:
            if not cert.pdf:
                continue
            event_part = slugify(cert.event.title) or "event"
            iin_part = (cert.iin or "")[:4] or "iin"
            name = f"{event_part}_{iin_part}_{cert.code}.pdf"
            with cert.pdf.open("rb") as f:
                zf.writestr(name, f.read())
    buf.seek(0)
    filename = f"certificates_{iin[:4]}.zip"
    response = HttpResponse(buf.getvalue(), content_type="application/zip")
    response["Content-Disposition"] = f'attachment; filename=\"{filename}\"'
    return response

def register(request):
    if request.method == "POST":
        form = RegistrationForm(request.POST)
        if form.is_valid():
            user = form.save()
            role = form.cleaned_data["role"]
            iin = form.cleaned_data["iin"]
            group, _ = Group.objects.get_or_create(name=role)
            user.groups.add(group)
            UserProfile.objects.get_or_create(user=user, defaults={"iin": iin})
            login(request, user)
            if role == "user_teacher":
                return redirect("teacher_dashboard")
            return redirect("student_dashboard")
    else:
        form = RegistrationForm()
    return render(request, "public/register.html", {"form": form})

def register_org(request):
    key = request.GET.get("key", "")
    if not key or key != settings.ORG_REGISTER_KEY:
        return HttpResponseForbidden("Недействительная ссылка регистрации организации.")
    if request.method == "POST":
        form = OrganizationRegistrationForm(request.POST, request.FILES)
        if form.is_valid():
            user = form.save()
            org = Organization.objects.create(
                name=form.cleaned_data["org_name"],
                logo=form.cleaned_data.get("org_logo"),
            )
            OrganizationUser.objects.create(organization=org, user=user)
            user.is_staff = True
            user.save(update_fields=["is_staff"])
            group, _ = Group.objects.get_or_create(name="staff_org")
            user.groups.add(group)
            login(request, user)
            return redirect("staff_events")
    else:
        form = OrganizationRegistrationForm()
    return render(request, "public/register_org.html", {"form": form})

def _get_org_for_user(user):
    if not user.is_authenticated:
        return None
    membership = OrganizationUser.objects.filter(user=user).select_related("organization").first()
    return membership.organization if membership else None

def _get_staff_event_queryset(user):
    org = _get_org_for_user(user)
    if not org:
        return Event.objects.none()
    return Event.objects.filter(organization=org, created_by=user)

def staff_portal_required(view_func):
    @wraps(view_func)
    @login_required(login_url="/accounts/login/")
    def _wrapped(request, *args, **kwargs):
        if not request.user.is_staff:
            return HttpResponseForbidden("Нет доступа к staff-разделу.")
        return view_func(request, *args, **kwargs)
    return _wrapped

@login_required
def profile(request):
    profile = UserProfile.objects.filter(user=request.user).first()
    if request.method == "POST":
        form = ProfileIINForm(request.POST)
        if form.is_valid():
            iin = form.cleaned_data["iin"]
            if profile:
                profile.iin = iin
                profile.save(update_fields=["iin"])
            else:
                UserProfile.objects.create(user=request.user, iin=iin)
            messages.success(request, "ИИН обновлён.")
            return redirect("profile")
    else:
        form = ProfileIINForm(initial={"iin": profile.iin if profile else ""})
    return render(request, "public/profile.html", {"form": form})

def _in_group(user, group_name: str) -> bool:
    return user.is_authenticated and user.groups.filter(name=group_name).exists()

def _require_group(group_name: str):
    return user_passes_test(lambda u: _in_group(u, group_name))

@login_required
def account_profile_redirect(request):
    if request.user.is_staff:
        return redirect("staff_events")
    if _in_group(request.user, "user_teacher"):
        return redirect("teacher_dashboard")
    if _in_group(request.user, "user_student"):
        return redirect("student_dashboard")
    return redirect("profile")

@login_required
@_require_group("user_student")
def student_dashboard(request):
    profile = UserProfile.objects.filter(user=request.user).first()
    if not profile or not profile.iin:
        return render(request, "student/dashboard.html", {"certificates": [], "profile": profile, "error": "Не задан ИИН в профиле."})
    certs = Certificate.objects.filter(iin=profile.iin).select_related("event").order_by("-created_at")
    return render(request, "student/dashboard.html", {"certificates": certs, "profile": profile})

@login_required
@_require_group("user_student")
def student_certificate_download(request, cert_id: int):
    cert = get_object_or_404(Certificate, pk=cert_id)
    profile = UserProfile.objects.filter(user=request.user).first()
    if not profile or cert.iin != profile.iin:
        return HttpResponseForbidden("Нет доступа к этому сертификату.")
    if not cert.pdf:
        raise Http404()
    event_part = slugify(cert.event.title) or "event"
    iin_part = (cert.iin or "")[:4] or "iin"
    filename = f"{event_part}_{iin_part}.pdf"
    return FileResponse(cert.pdf.open("rb"), as_attachment=True, filename=filename)

@login_required
@_require_group("user_teacher")
def teacher_dashboard(request):
    profile = UserProfile.objects.filter(user=request.user).first()
    own_iin = profile.iin if profile else ""
    links = TeacherStudent.objects.filter(teacher=request.user).order_by("-created_at")
    linked_iins = list(links.values_list("student_iin", flat=True))
    iins = [own_iin] if own_iin else []
    iins.extend([i for i in linked_iins if i and i not in iins])
    certs = Certificate.objects.filter(iin__in=iins).select_related("event").order_by("-created_at")
    form = TeacherAddStudentForm()
    return render(request, "teacher/dashboard.html", {
        "certificates": certs,
        "profile": profile,
        "linked_iins": linked_iins,
        "links": links,
        "form": form,
    })

@login_required
@_require_group("user_teacher")
def teacher_add_student(request):
    if request.method != "POST":
        return HttpResponseNotAllowed(["POST"])
    form = TeacherAddStudentForm(request.POST)
    if form.is_valid():
        iin = form.cleaned_data["student_iin"]
        TeacherStudent.objects.get_or_create(teacher=request.user, student_iin=iin)
        messages.success(request, "Ученик добавлен.")
    else:
        messages.error(request, "Ошибка в ИИН.")
    return redirect("teacher_dashboard")

@login_required
@_require_group("user_teacher")
def teacher_remove_student(request):
    if request.method != "POST":
        return HttpResponseNotAllowed(["POST"])
    iin = (request.POST.get("student_iin") or "").strip()
    if iin:
        TeacherStudent.objects.filter(teacher=request.user, student_iin=iin).delete()
        messages.success(request, "Ученик удалён.")
    return redirect("teacher_dashboard")

@login_required
@_require_group("user_teacher")
def teacher_certificate_download(request, cert_id: int):
    cert = get_object_or_404(Certificate, pk=cert_id)
    profile = UserProfile.objects.filter(user=request.user).first()
    own_iin = profile.iin if profile else ""
    linked_iins = list(TeacherStudent.objects.filter(teacher=request.user).values_list("student_iin", flat=True))
    allowed = cert.iin == own_iin or cert.iin in linked_iins
    if not allowed:
        return HttpResponseForbidden("Нет доступа к этому сертификату.")
    if not cert.pdf:
        raise Http404()
    event_part = slugify(cert.event.title) or "event"
    iin_part = (cert.iin or "")[:4] or "iin"
    filename = f"{event_part}_{iin_part}.pdf"
    return FileResponse(cert.pdf.open("rb"), as_attachment=True, filename=filename)

def download_certificate(request, code: str):
    iin = request.session.get("iin", "")
    if not iin:
        return HttpResponseForbidden("Сначала введите ИИН.")
    cert = get_object_or_404(Certificate, code=code)
    if cert.iin != iin:
        return HttpResponseForbidden("Нет доступа к этому сертификату.")
    if not cert.pdf:
        raise Http404()
    # Build friendly filename: <event>_<iin4>.pdf
    event_part = slugify(cert.event.title) or "event"
    iin_part = (cert.iin or "")[:4] or "iin"
    filename = f"{event_part}_{iin_part}.pdf"
    return FileResponse(cert.pdf.open("rb"), as_attachment=True, filename=filename)

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

@staff_portal_required
def staff_events(request):
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    events = _get_staff_event_queryset(request.user).order_by("-created_at")
    if request.method == "POST":
        form = EventForm(request.POST)
        if form.is_valid():
            ev = form.save(commit=False)
            ev.organization = org
            ev.created_by = request.user
            ev.save()
            AuditLog.objects.create(actor=request.user, action="create_event", object_type="Event", object_id=str(ev.pk))
            return redirect("staff_events")
    else:
        form = EventForm()
    return render(request, "staff/events.html", {"events": events, "form": form})

@staff_portal_required
def staff_event_detail(request, event_id: int):
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    event = get_object_or_404(_get_staff_event_queryset(request.user), pk=event_id)
    latest_template = Template.objects.filter(event=event).order_by("-created_at").first()
    batches = ImportBatch.objects.filter(event=event).order_by("-created_at")[:20]

    tform = TemplateUploadForm()
    iform = ImportUploadForm()

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
                _parse_batch_file(batch, event)

                messages.success(request, "Данные загружены. Проверь mapping и запускай Generate.")
                return redirect("staff_batch_mapping", batch_id=batch.pk)

    return render(request, "staff/event_detail.html", {
        "event": event,
        "latest_template": latest_template,
        "batches": batches,
        "tform": tform,
        "iform": iform,
    })

@staff_portal_required
def staff_event_delete(request, event_id: int):
    if request.method != "POST":
        return HttpResponseNotAllowed(["POST"])
    event = get_object_or_404(_get_staff_event_queryset(request.user), pk=event_id)
    AuditLog.objects.create(actor=request.user, action="delete_event", object_type="Event", object_id=str(event.pk))
    event.delete()
    messages.success(request, "Событие удалено.")
    return redirect("staff_events")

def _parse_batch_file(batch: ImportBatch, event: Event) -> None:
    # Parse file and create ParticipantRows (MVP)
    file_path = batch.file.path
    cols, rows = read_table(file_path)
    batch.rows_total = len(rows)
    missing = validate_required(cols)
    report = {"columns": cols, "missing_required": missing, "preview": rows[:10]}
    batch.report_json = report
    # Extract tokens from latest template (first slide)
    tpl = Template.objects.filter(event=event).order_by("-created_at").first()
    tokens = extract_tokens_from_pptx(tpl.pptx_file.path) if tpl else DEFAULT_TOKENS
    batch.tokens_json = tokens
    batch.mapping_json = default_mapping_for_tokens(cols, tokens)
    batch.status = "validated" if not missing else "uploaded"
    batch.save(update_fields=["rows_total", "report_json", "mapping_json", "tokens_json", "status"])

    # Create rows
    with transaction.atomic():
        ParticipantRow.objects.filter(batch=batch).delete()
        for r in rows:
            iin = (r.get("iin") or "").strip()
            name = (r.get("name") or "").strip()
            iin_valid = iin.isdigit() and len(iin) == 12
            if iin and not iin_valid:
                error = f"Неверный ИИН: {iin}"
                status = "failed"
            elif not iin or not name:
                error = "Missing iin/name"
                status = "failed"
            else:
                error = ""
                status = "pending"
            ParticipantRow.objects.create(
                batch=batch,
                iin=iin[:12],
                name=name,
                payload_json=r,
                status=status,
                error=error,
            )

@staff_portal_required
def staff_batch_mapping(request, batch_id: int):
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    batch = get_object_or_404(ImportBatch, pk=batch_id, event__organization=org, event__created_by=request.user)
    report = batch.report_json or {}
    columns = report.get("columns", [])
    mapping = batch.mapping_json or {}
    tokens = batch.tokens_json or []
    # Fallback: if batch wasn't parsed (older error), parse now
    if not columns and batch.file:
        cols, rows = read_table(batch.file.path)
        batch.rows_total = len(rows)
        missing = validate_required(cols)
        batch.report_json = {"columns": cols, "missing_required": missing, "preview": rows[:10]}
        if not tokens:
            tpl = Template.objects.filter(event=batch.event).order_by("-created_at").first()
            tokens = extract_tokens_from_pptx(tpl.pptx_file.path) if tpl else DEFAULT_TOKENS
            batch.tokens_json = tokens
        batch.mapping_json = default_mapping_for_tokens(cols, tokens or DEFAULT_TOKENS)
        batch.status = "validated" if not missing else "uploaded"
        batch.save(update_fields=["rows_total", "report_json", "mapping_json", "tokens_json", "status"])
        if not batch.rows.exists():
            with transaction.atomic():
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
        report = batch.report_json
        columns = report.get("columns", [])
        mapping = batch.mapping_json or {}
        tokens = batch.tokens_json or tokens
    if not tokens:
        tpl = Template.objects.filter(event=batch.event).order_by("-created_at").first()
        tokens = extract_tokens_from_pptx(tpl.pptx_file.path) if tpl else DEFAULT_TOKENS
        batch.tokens_json = tokens
        batch.save(update_fields=["tokens_json"])
    # If mapping is empty but columns exist, auto-fill mapping for detected tokens
    if not mapping and columns and tokens:
        batch.mapping_json = default_mapping_for_tokens(columns, tokens)
        batch.save(update_fields=["mapping_json"])
        mapping = batch.mapping_json or {}
    # Do not show fqr in mapping UI (QR is generated automatically)
    mapping_tokens = [t for t in tokens if t != "fqr"]

    class _DynMappingForm(MappingForm):
        pass

    for tok in mapping_tokens:
        _DynMappingForm.base_fields[tok] = __import__("django").forms.ChoiceField(
            label=f"{tok} → колонка",
            choices=[("", "— не вставлять —")] + [(c, c) for c in columns],
            required=False,
            initial=mapping.get(tok, ""),
        )

    if request.method == "POST":
        form = _DynMappingForm(request.POST)
        if form.is_valid():
            batch.mapping_json = {tok: form.cleaned_data.get(tok, "") for tok in mapping_tokens}
            batch.save(update_fields=["mapping_json"])
            AuditLog.objects.create(actor=request.user, action="mapping_update", object_type="ImportBatch", object_id=str(batch.pk))
            messages.success(request, "Mapping сохранён.")
            if "save_and_generate" in request.POST:
                return redirect("staff_batch_generate", batch_id=batch.pk)
    else:
        form = _DynMappingForm()

    return render(request, "staff/mapping.html", {"batch": batch, "form": form, "report": report})

@staff_portal_required
def staff_batch_edit(request, batch_id: int):
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    batch = get_object_or_404(ImportBatch, pk=batch_id, event__organization=org, event__created_by=request.user)
    if request.method == "POST":
        form = BatchEditForm(request.POST, request.FILES, instance=batch)
        if form.is_valid():
            form.save()
            _parse_batch_file(batch, batch.event)
            AuditLog.objects.create(actor=request.user, action="batch_edit", object_type="ImportBatch", object_id=str(batch.pk))
            messages.success(request, "Batch обновлён.")
            return redirect("staff_batch_mapping", batch_id=batch.pk)
    else:
        form = BatchEditForm(instance=batch)
    return render(request, "staff/batch_edit.html", {"batch": batch, "form": form})

@staff_portal_required
def staff_batch_delete(request, batch_id: int):
    if request.method != "POST":
        return HttpResponseNotAllowed(["POST"])
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    batch = get_object_or_404(ImportBatch, pk=batch_id, event__organization=org, event__created_by=request.user)
    AuditLog.objects.create(actor=request.user, action="batch_delete", object_type="ImportBatch", object_id=str(batch.pk))
    batch.delete()
    messages.success(request, "Batch удалён.")
    return redirect("staff_event_detail", event_id=batch.event_id)

@staff_portal_required
def staff_batch_generate(request, batch_id: int):
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    batch = get_object_or_404(ImportBatch, pk=batch_id, event__organization=org, event__created_by=request.user)
    if request.method == "POST":
        # start celery task
        batch.status = "generating"
        batch.save(update_fields=["status"])
        if hasattr(generate_batch, "delay"):
            generate_batch.delay(batch.pk)
            messages.success(request, "Генерация запущена. Обновляй страницу для прогресса.")
        else:
            generate_batch(batch.pk)
            messages.success(request, "Генерация выполнена синхронно в веб-процессе.")
        return redirect("staff_batch_generate", batch_id=batch.pk)

    # progress
    total = batch.rows_total or 0
    ok = batch.rows_ok or 0
    failed = batch.rows_failed or 0
    done = ok + failed
    percent = 0
    if total:
        percent = int(round((done / total) * 100))
        if percent > 100:
            percent = 100
    return render(request, "staff/generate.html", {"batch": batch, "total": total, "ok": ok, "failed": failed, "percent": percent})

@staff_portal_required
def staff_event_certificates(request, event_id: int):
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    event = get_object_or_404(_get_staff_event_queryset(request.user), pk=event_id)
    certs = Certificate.objects.filter(event=event).order_by("-created_at")
    total = certs.count()
    return render(request, "staff/certificates.html", {"event": event, "certificates": certs, "total": total})

@staff_portal_required
def staff_event_certificates_download(request, event_id: int):
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    event = get_object_or_404(_get_staff_event_queryset(request.user), pk=event_id)
    certs = Certificate.objects.filter(event=event).order_by("-created_at")
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w", zipfile.ZIP_DEFLATED) as zf:
        for cert in certs:
            if not cert.pdf:
                continue
            event_part = slugify(event.title) or "event"
            iin_part = (cert.iin or "")[:4] or "iin"
            name = f"{event_part}_{iin_part}_{cert.code}.pdf"
            with cert.pdf.open("rb") as f:
                zf.writestr(name, f.read())
    buf.seek(0)
    filename = f"{slugify(event.title) or 'event'}_certificates.zip"
    response = HttpResponse(buf.getvalue(), content_type="application/zip")
    response["Content-Disposition"] = f'attachment; filename="{filename}"'
    return response

@staff_portal_required
def staff_certificate_download(request, cert_id: int):
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    cert = get_object_or_404(Certificate, pk=cert_id, organization=org, event__created_by=request.user)
    if not cert.pdf:
        raise Http404()
    event_part = slugify(cert.event.title) or "event"
    iin_part = (cert.iin or "")[:4] or "iin"
    filename = f"{event_part}_{iin_part}.pdf"
    return FileResponse(cert.pdf.open("rb"), as_attachment=True, filename=filename)

@staff_portal_required
def staff_certificate_edit(request, cert_id: int):
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    cert = get_object_or_404(Certificate, pk=cert_id, organization=org, event__created_by=request.user)
    # limit editable tokens
    desired_tokens = {"name", "school", "class", "teacher", "text", "id"}
    tpl = Template.objects.filter(event=cert.event).order_by("-created_at").first()
    tpl_tokens = extract_tokens_from_pptx(tpl.pptx_file.path) if tpl else []
    tokens = [t for t in tpl_tokens if t in desired_tokens] or [t for t in desired_tokens if t != "qr"]

    if request.method == "POST":
        form = CertificateEditForm(request.POST, instance=cert, tokens=tokens, payload=cert.payload_json or {})
        if form.is_valid():
            cert = form.save()
            payload = cert.payload_json or {}
            for tok in tokens:
                payload[tok] = form.cleaned_data.get(f"tok_{tok}", "")
            # keep name in payload for template usage
            payload["name"] = cert.name
            cert.payload_json = payload
            cert.save(update_fields=["payload_json"])

            # regenerate PDF
            if tpl and cert.pdf:
                verify_url = settings.PUBLIC_BASE_URL.rstrip("/") + f"/verify/{cert.code}/"
                token_to_value = dict(payload)
                # legacy f* aliases
                token_to_value.setdefault("fname", cert.name)
                token_to_value.setdefault("fschool", payload.get("school", ""))
                token_to_value.setdefault("fclass", payload.get("class", ""))
                token_to_value.setdefault("fteacher", payload.get("teacher", ""))
                token_to_value.setdefault("ftext", payload.get("text", ""))
                token_to_value.setdefault("fid", payload.get("id", ""))
                pdf_bytes = generate_certificate_pdf(tpl.pptx_file.path, token_to_value, verify_url)
                cert.pdf.save(f"{cert.code}.pdf", ContentFile(pdf_bytes), save=True)

            AuditLog.objects.create(actor=request.user, action="certificate_edit", object_type="Certificate", object_id=str(cert.pk))
            messages.success(request, "Сертификат обновлён.")
            return redirect("staff_event_certificates", event_id=cert.event_id)
    else:
        form = CertificateEditForm(instance=cert, tokens=tokens, payload=cert.payload_json or {})
    return render(request, "staff/certificate_edit.html", {"cert": cert, "form": form})

@staff_portal_required
def staff_certificate_delete(request, cert_id: int):
    if request.method != "POST":
        return HttpResponseNotAllowed(["POST"])
    org = _get_org_for_user(request.user)
    if not org:
        return HttpResponseForbidden("Нет доступа: пользователь не привязан к организации.")
    cert = get_object_or_404(Certificate, pk=cert_id, organization=org, event__created_by=request.user)
    event_id = cert.event_id
    AuditLog.objects.create(actor=request.user, action="certificate_delete", object_type="Certificate", object_id=str(cert.pk))
    cert.delete()
    messages.success(request, "Сертификат удалён.")
    return redirect("staff_event_certificates", event_id=event_id)
