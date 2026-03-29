FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

# LibreOffice for PPTX->PDF conversion
RUN apt-get update && apt-get install -y --no-install-recommends     libreoffice     libreoffice-impress     fonts-dejavu     fontconfig     && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . /app

# Keep frontend build artifacts but exclude node_modules to avoid huge image layers
RUN rm -rf /app/frontend/jetistik-your-certificates-verified-94-main/node_modules || true

EXPOSE 8000
CMD ["python", "manage.py", "runserver", "0.0.0.0:8000"]
