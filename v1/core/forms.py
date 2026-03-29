from django import forms
from django.contrib.auth.forms import UserCreationForm
from django.contrib.auth.models import User
from .models import Event, Template, ImportBatch, Certificate, TeacherStudent, Organization

class IINSearchForm(forms.Form):
    iin = forms.CharField(label="ИИН", max_length=12, min_length=12)

    def clean_iin(self):
        iin = self.cleaned_data["iin"].strip()
        if not iin.isdigit() or len(iin) != 12:
            raise forms.ValidationError("ИИН должен быть из 12 цифр.")
        return iin

class EventForm(forms.ModelForm):
    class Meta:
        model = Event
        fields = ["title", "date", "city", "description"]

class TemplateUploadForm(forms.ModelForm):
    class Meta:
        model = Template
        fields = ["pptx_file"]

class ImportUploadForm(forms.ModelForm):
    class Meta:
        model = ImportBatch
        fields = ["file"]

class BatchEditForm(forms.ModelForm):
    class Meta:
        model = ImportBatch
        fields = ["file"]

class CertificateEditForm(forms.ModelForm):
    def __init__(self, *args, tokens=None, payload=None, **kwargs):
        super().__init__(*args, **kwargs)
        tokens = tokens or []
        payload = payload or {}
        for tok in tokens:
            if tok in ("fqr", "qr"):
                continue
            self.fields[f"tok_{tok}"] = forms.CharField(
                required=False,
                label=tok,
                initial=payload.get(tok, "")
            )

    class Meta:
        model = Certificate
        fields = ["name", "iin", "status", "revoked_reason"]

class TeacherAddStudentForm(forms.Form):
    student_iin = forms.CharField(label="ИИН ученика", max_length=12, min_length=12)

    def clean_student_iin(self):
        iin = self.cleaned_data["student_iin"].strip()
        if not iin.isdigit() or len(iin) != 12:
            raise forms.ValidationError("ИИН должен быть из 12 цифр.")
        return iin

class RegistrationForm(UserCreationForm):
    ROLE_CHOICES = [
        ("user_student", "Ученик"),
        ("user_teacher", "Учитель"),
    ]
    role = forms.ChoiceField(label="Роль", choices=ROLE_CHOICES)
    iin = forms.CharField(label="ИИН", max_length=12, min_length=12)

    class Meta:
        model = User
        fields = ("username", "role", "iin", "password1", "password2")

    def clean_iin(self):
        iin = self.cleaned_data["iin"].strip()
        if not iin.isdigit() or len(iin) != 12:
            raise forms.ValidationError("ИИН должен быть из 12 цифр.")
        return iin

class ProfileIINForm(forms.Form):
    iin = forms.CharField(label="ИИН", max_length=12, min_length=12)

    def clean_iin(self):
        iin = self.cleaned_data["iin"].strip()
        if not iin.isdigit() or len(iin) != 12:
            raise forms.ValidationError("ИИН должен быть из 12 цифр.")
        return iin

class OrganizationRegistrationForm(UserCreationForm):
    org_name = forms.CharField(label="Название организации", max_length=255)
    org_logo = forms.ImageField(label="Логотип", required=False)

    class Meta:
        model = User
        fields = ("username", "org_name", "org_logo", "password1", "password2")

class MappingForm(forms.Form):
    # token -> column mapping will be added dynamically
    pass
