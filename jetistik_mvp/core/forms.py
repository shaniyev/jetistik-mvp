from django import forms
from .models import Event, Template, ImportBatch

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

class MappingForm(forms.Form):
    # token -> column mapping will be added dynamically
    pass
