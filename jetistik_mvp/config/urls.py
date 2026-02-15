from django.contrib import admin
from django.urls import path
from django.conf import settings
from django.conf.urls.static import static

from core import views

urlpatterns = [
    path("admin/", admin.site.urls),

    # Public
    path("", views.home, name="home"),
    path("my/", views.my_certificates, name="my_certificates"),
    path("verify/<str:code>/", views.verify, name="verify"),
    path("download/<str:code>/", views.download_certificate, name="download_certificate"),

    # Staff dashboard (минимальный)
    path("staff/events/", views.staff_events, name="staff_events"),
    path("staff/events/<int:event_id>/", views.staff_event_detail, name="staff_event_detail"),
    path("staff/batch/<int:batch_id>/mapping/", views.staff_batch_mapping, name="staff_batch_mapping"),
    path("staff/batch/<int:batch_id>/generate/", views.staff_batch_generate, name="staff_batch_generate"),
]

if settings.DEBUG:
    urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
