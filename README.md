# Jetistik MVP (Django)
MVP веб‑сервис для автоматизации генерации дипломов/сертификатов из PPTX + XLSX/CSV, с QR‑верификацией.

## Возможности (MVP)
- Админ:
  - Создание Event
  - Загрузка PPTX шаблона (на 1-й слайд) с токенами `fname`, `fschool`, `fclass`, `fplace`, `fteacher`, `fnomination`, `fid`
  - Загрузка XLSX/CSV участников
  - Авто‑mapping колонок ↔ токенов (можно поправить)
  - Генерация PDF (PPTX → PDF через LibreOffice headless)
  - Просмотр сертификатов + revoke/unrevoke
- Публично:
  - `/` ввод ИИН → `/my` список сертификатов → скачивание PDF
  - `/verify/<code>` проверка подлинности (VALID/REVOKED/NOT_FOUND), ИИН маской

## Быстрый старт (Docker)
1) Установите Docker и Docker Compose
2) В корне проекта:
```bash
cp .env.example .env
docker compose up --build
```
3) Миграции и суперюзер (в отдельном терминале):
```bash
docker compose exec web python manage.py migrate
docker compose exec web python manage.py createsuperuser
```
4) Откройте:
- Админка: `http://localhost:8000/admin/`
- Публично: `http://localhost:8000/`

## Production deployment (VPS/hosting)
Используйте отдельный compose-файл и production env.

1) Подготовка:
```bash
cp .env.prod.example .env.prod
```
Заполните в `.env.prod`:
- `SECRET_KEY`
- `ALLOWED_HOSTS`
- `CSRF_TRUSTED_ORIGINS`
- `DATABASE_URL` / `POSTGRES_*`
- `EMAIL_HOST_USER` / `EMAIL_HOST_PASSWORD`

2) Запуск:
```bash
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
```

3) Создание админа:
```bash
docker compose -f docker-compose.prod.yml --env-file .env.prod exec web python manage.py createsuperuser
```

Сервисы в production:
- `web` (gunicorn, авто `migrate` + `collectstatic`)
- `worker` (celery)
- `db` (postgres)
- `redis`

## Без Docker (локально)
Зависимости:
- Python 3.11+
- Redis (для Celery) **или** включите `CELERY_TASK_ALWAYS_EAGER=True`
- LibreOffice (для PPTX → PDF)

```bash
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
python manage.py migrate
python manage.py createsuperuser
python manage.py runserver
# в отдельном терминале:
celery -A config worker -l info
```

## LibreOffice
Конвертация PPTX → PDF делается командой:
`soffice --headless --convert-to pdf --outdir <outdir> <pptx>`

В Docker LibreOffice ставится внутри образа.

## Шаблон PPTX
- На 1‑м слайде токены `fname`, `fschool`, `fclass`, `fplace`, `fteacher`, `fnomination`, `fid`
- Для QR:
  - предпочтительно: фигура с **name = "QR"**
  - fallback: текстовый маркер `fqr` (будет заменён картинкой)

## Примечания по MVP
- Для упрощения скачивание PDF контролируется через **session iin**: пользователь вводит ИИН → в session сохраняется iin → `/download/<code>` проверяет совпадение.
- Rate limit: 10 запросов/мин на IP для формы поиска.
- Логи действий: AuditLog (upload/generate/revoke).

---
Если хочешь — я могу:
- добавить полноценный прогресс‑бар (polling + websocket),
- S3 storage + presigned URLs,
- многослайдовые шаблоны,
- массовую генерацию в фоне с retry и DLQ.
