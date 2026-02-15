from django.contrib import admin
from django.urls import path, include
from django.conf import settings
from django.conf.urls.static import static
from django.views.static import serve as static_serve

from core import views

urlpatterns = [
    path("admin/", admin.site.urls),
    path("accounts/", include("django.contrib.auth.urls")),

    # Public
    path("", views.landing, name="landing"),
    path("verify/", views.home, name="home"),
    path("organizers/", views.landing, name="organizers_landing"),
    path("api/organizer-request/", views.organizer_request_submit, name="organizer_request_submit"),
    path("my/", views.my_certificates, name="my_certificates"),
    path("my/download/", views.my_certificates_download, name="my_certificates_download"),
    path("register/", views.register, name="register"),
    path("register/org/", views.register_org, name="register_org"),
    path("profile/", views.profile, name="profile"),
    path("student/", views.student_dashboard, name="student_dashboard"),
    path("student/certificates/<int:cert_id>/download/", views.student_certificate_download, name="student_certificate_download"),
    path("teacher/", views.teacher_dashboard, name="teacher_dashboard"),
    path("teacher/add-student/", views.teacher_add_student, name="teacher_add_student"),
    path("teacher/remove-student/", views.teacher_remove_student, name="teacher_remove_student"),
    path("teacher/certificates/<int:cert_id>/download/", views.teacher_certificate_download, name="teacher_certificate_download"),
    path("verify/<str:code>/", views.verify, name="verify"),
    path("download/<str:code>/", views.download_certificate, name="download_certificate"),

    # Staff dashboard (минимальный)
    path("staff/events/", views.staff_events, name="staff_events"),
    path("staff/events/<int:event_id>/", views.staff_event_detail, name="staff_event_detail"),
    path("staff/events/<int:event_id>/delete/", views.staff_event_delete, name="staff_event_delete"),
    path("staff/batch/<int:batch_id>/mapping/", views.staff_batch_mapping, name="staff_batch_mapping"),
    path("staff/batch/<int:batch_id>/generate/", views.staff_batch_generate, name="staff_batch_generate"),
    path("staff/batch/<int:batch_id>/edit/", views.staff_batch_edit, name="staff_batch_edit"),
    path("staff/batch/<int:batch_id>/delete/", views.staff_batch_delete, name="staff_batch_delete"),
    path("staff/events/<int:event_id>/certificates/", views.staff_event_certificates, name="staff_event_certificates"),
    path("staff/events/<int:event_id>/certificates/download/", views.staff_event_certificates_download, name="staff_event_certificates_download"),
    path("staff/certificates/<int:cert_id>/download/", views.staff_certificate_download, name="staff_certificate_download"),
    path("staff/certificates/<int:cert_id>/edit/", views.staff_certificate_edit, name="staff_certificate_edit"),
    path("staff/certificates/<int:cert_id>/delete/", views.staff_certificate_delete, name="staff_certificate_delete"),
]

if settings.DEBUG:
    urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
    urlpatterns += [
        path("assets/<path:path>", static_serve, {"document_root": settings.BASE_DIR / "static/landing/assets"}),
        path("favicon.ico", static_serve, {"document_root": settings.BASE_DIR / "static/landing", "path": "favicon.ico"}),
        path("robots.txt", static_serve, {"document_root": settings.BASE_DIR / "static/landing", "path": "robots.txt"}),
        path("placeholder.svg", static_serve, {"document_root": settings.BASE_DIR / "static/landing", "path": "placeholder.svg"}),
    ]
