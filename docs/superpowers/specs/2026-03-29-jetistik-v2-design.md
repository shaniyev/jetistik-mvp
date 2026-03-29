# Jetistik v2 — Design Specification

## 1. Overview

Jetistik v2 is a complete rewrite of the certificate generation and QR verification platform from Django/React to Go/SvelteKit. The platform enables organizations to upload PPTX templates + participant data (CSV/XLSX) and automatically generate personalized PDF certificates with QR verification codes.

**Key drivers for v2:**
- Full stack migration: Django → Go, React → SvelteKit
- Custom admin console replacing Django Admin
- Improved architecture (modular monolith)
- S3-compatible file storage (MinIO)
- Real-time generation progress (SSE)
- Production data migration (1700+ certificates)

---

## 2. Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Backend | Go 1.23+, chi | HTTP API, modular monolith |
| Database | PostgreSQL 16 | Primary data store |
| DB Queries | sqlc | SQL → type-safe Go code generation |
| DB Migrations | goose | SQL-based schema migrations |
| Frontend | SvelteKit, TypeScript | SSR landing + SPA dashboards |
| Styling | Tailwind CSS | Utility-first, "Sovereign Ledger" design system |
| File Storage | MinIO | S3-compatible, media files only |
| Task Queue | Asynq (Redis) | Async PDF generation |
| PDF Conversion | Gotenberg | PPTX → PDF via isolated container |
| Auth | JWT | Access (memory) + Refresh (httpOnly cookie) |
| Real-time | SSE | Generation progress streaming |
| I18n | KZ, RU, EN | Trilingual support |
| Deploy | Terraform + Ansible + Docker Compose | OpenStack IaC |

---

## 3. Project Structure

```
jetistik-mvp/
├── v1/                    # legacy Django MVP (read-only reference)
├── backend/               # Go modular monolith
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── auth/          # JWT, login, register, roles
│   │   ├── organization/  # CRUD organizations, members
│   │   ├── event/         # CRUD events
│   │   ├── template/      # PPTX upload, token extraction
│   │   ├── batch/         # CSV/XLSX import, mapping, generation trigger
│   │   ├── certificate/   # certificates, verification, QR
│   │   ├── user/          # profiles, teacher-student links
│   │   ├── audit/         # audit log
│   │   ├── admin/         # admin panel API
│   │   ├── storage/       # MinIO client
│   │   ├── worker/        # Asynq tasks (PDF generation)
│   │   ├── notification/  # email, SSE
│   │   └── platform/      # config, middleware, db, errors, i18n
│   ├── migrations/        # goose SQL migrations
│   ├── queries/           # sqlc SQL queries
│   ├── sqlc.yaml
│   ├── go.mod
│   └── go.sum
├── frontend/              # SvelteKit
│   ├── src/
│   │   ├── routes/
│   │   │   ├── (public)/  # landing, verify, FAQ (SSR)
│   │   │   ├── (auth)/    # login, register, logout
│   │   │   └── (app)/     # student, teacher, staff, admin (SPA)
│   │   ├── lib/
│   │   │   ├── api/       # Go API client
│   │   │   ├── components/
│   │   │   ├── stores/
│   │   │   └── i18n/
│   │   └── app.html
│   ├── svelte.config.js
│   ├── tailwind.config.ts
│   └── package.json
├── stitch/                # design concepts (reference)
├── deploy/
│   ├── terraform/         # OpenStack IaC
│   └── ansible/           # deployment playbooks
├── docker-compose.yml     # dev
├── docker-compose.prod.yml
├── CLAUDE.md
└── Makefile
```

---

## 4. Backend Architecture

### Modular Monolith

Each module in `internal/` follows a consistent structure:

```
module/
├── handler.go      # HTTP handlers (chi router)
├── service.go      # business logic
├── repository.go   # data access interface
└── dto.go          # request/response structs
```

**Dependency flow:** handler → service → repository

**Inter-module communication:** via interfaces. Module `batch` depends on `certificate.Service` (interface), not on the concrete implementation. Wiring happens in `cmd/server/main.go`.

### Platform Module (`internal/platform/`)

Shared infrastructure:
- `config/` — env-based configuration loading
- `db/` — PostgreSQL connection pool (pgxpool)
- `middleware/` — JWT, CORS, rate-limit, request-id, logging
- `errors/` — standardized API errors
- `response/` — JSON response helpers
- `validator/` — validation (IIN, email, etc.)

### Router (chi)

```go
r.Route("/api/v1", func(r chi.Router) {
    // public
    r.Group(func(r chi.Router) {
        r.Post("/auth/login", auth.Login)
        r.Post("/auth/register", auth.Register)
        r.Get("/verify/{code}", certificate.Verify)
    })
    // protected
    r.Group(func(r chi.Router) {
        r.Use(auth.JWTMiddleware)
        r.Mount("/events", event.Routes())
        r.Mount("/certificates", certificate.Routes())
        r.Mount("/organizations", organization.Routes())
    })
})
```

---

## 5. Database Schema

```sql
CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(150) UNIQUE NOT NULL,
    email       VARCHAR(254) UNIQUE,
    password    TEXT NOT NULL,
    iin         VARCHAR(12),
    role        VARCHAR(20) NOT NULL,       -- admin, staff, teacher, student
    is_active   BOOLEAN DEFAULT true,
    language    VARCHAR(2) DEFAULT 'kz',
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_users_iin ON users(iin);

CREATE TABLE organizations (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    domain      VARCHAR(255),
    logo_path   TEXT,
    status      VARCHAR(20) DEFAULT 'active',
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE organization_members (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id),
    user_id         BIGINT NOT NULL REFERENCES users(id),
    role            VARCHAR(20) DEFAULT 'member',
    created_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(organization_id, user_id)
);

CREATE TABLE events (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id),
    created_by      BIGINT REFERENCES users(id),
    title           VARCHAR(255) NOT NULL,
    date            DATE,
    city            VARCHAR(128) DEFAULT '',
    description     TEXT DEFAULT '',
    status          VARCHAR(20) DEFAULT 'active',
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE templates (
    id          BIGSERIAL PRIMARY KEY,
    event_id    BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    file_path   TEXT NOT NULL,
    tokens      JSONB DEFAULT '[]',
    created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE import_batches (
    id              BIGSERIAL PRIMARY KEY,
    event_id        BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    file_path       TEXT NOT NULL,
    status          VARCHAR(20) DEFAULT 'uploaded',
    rows_total      INT DEFAULT 0,
    rows_ok         INT DEFAULT 0,
    rows_failed     INT DEFAULT 0,
    mapping         JSONB DEFAULT '{}',
    tokens          JSONB DEFAULT '[]',
    report          JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE participant_rows (
    id          BIGSERIAL PRIMARY KEY,
    batch_id    BIGINT NOT NULL REFERENCES import_batches(id) ON DELETE CASCADE,
    iin         VARCHAR(12),
    name        VARCHAR(255),
    payload     JSONB DEFAULT '{}',
    status      VARCHAR(20) DEFAULT 'pending',
    error       TEXT DEFAULT '',
    created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_participant_rows_iin ON participant_rows(iin);
CREATE INDEX idx_participant_rows_batch ON participant_rows(batch_id);

CREATE TABLE certificates (
    id              BIGSERIAL PRIMARY KEY,
    event_id        BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    organization_id BIGINT REFERENCES organizations(id),
    iin             VARCHAR(12),
    name            VARCHAR(255),
    code            VARCHAR(64) UNIQUE NOT NULL,
    pdf_path        TEXT,
    status          VARCHAR(20) DEFAULT 'valid',
    revoked_reason  VARCHAR(255) DEFAULT '',
    payload         JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_certificates_iin ON certificates(iin);
CREATE INDEX idx_certificates_code ON certificates(code);
CREATE INDEX idx_certificates_event ON certificates(event_id);

CREATE TABLE teacher_students (
    id          BIGSERIAL PRIMARY KEY,
    teacher_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_iin VARCHAR(12) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now(),
    UNIQUE(teacher_id, student_iin)
);
CREATE INDEX idx_teacher_students_iin ON teacher_students(student_iin);

CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    actor_id    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action      VARCHAR(64) NOT NULL,
    object_type VARCHAR(64) DEFAULT '',
    object_id   VARCHAR(64) DEFAULT '',
    meta        JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id);

CREATE TABLE refresh_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
```

---

## 6. API Endpoints

### Auth
```
POST   /api/v1/auth/register
POST   /api/v1/auth/register/org
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
POST   /api/v1/auth/password-reset
POST   /api/v1/auth/password-reset/confirm
```

### Profile
```
GET    /api/v1/profile
PATCH  /api/v1/profile
GET    /api/v1/profile/qr
```

### Public
```
GET    /api/v1/verify/{code}
GET    /api/v1/certificates/search?iin=...
GET    /api/v1/certificates/{code}/download
POST   /api/v1/organizer-request
```

### Student
```
GET    /api/v1/student/certificates
GET    /api/v1/student/certificates/{id}/download
GET    /api/v1/student/certificates/download
```

### Teacher
```
GET    /api/v1/teacher/certificates
POST   /api/v1/teacher/students
DELETE /api/v1/teacher/students/{iin}
GET    /api/v1/teacher/students
GET    /api/v1/teacher/certificates/{id}/download
```

### Staff
```
GET    /api/v1/staff/events
POST   /api/v1/staff/events
GET    /api/v1/staff/events/{id}
PATCH  /api/v1/staff/events/{id}
DELETE /api/v1/staff/events/{id}
POST   /api/v1/staff/events/{id}/template
DELETE /api/v1/staff/events/{id}/template
POST   /api/v1/staff/events/{id}/batches
GET    /api/v1/staff/batches/{id}
PATCH  /api/v1/staff/batches/{id}/mapping
POST   /api/v1/staff/batches/{id}/generate
GET    /api/v1/staff/batches/{id}/progress          # SSE
DELETE /api/v1/staff/batches/{id}
GET    /api/v1/staff/events/{id}/certificates
GET    /api/v1/staff/events/{id}/certificates/download
GET    /api/v1/staff/certificates/{id}/download
PATCH  /api/v1/staff/certificates/{id}
DELETE /api/v1/staff/certificates/{id}
POST   /api/v1/staff/certificates/{id}/revoke
POST   /api/v1/staff/certificates/{id}/unrevoke
```

### Admin
```
GET    /api/v1/admin/organizations
POST   /api/v1/admin/organizations
GET    /api/v1/admin/organizations/{id}
PATCH  /api/v1/admin/organizations/{id}
DELETE /api/v1/admin/organizations/{id}
GET    /api/v1/admin/organizations/{id}/members
POST   /api/v1/admin/organizations/{id}/members
DELETE /api/v1/admin/organizations/{id}/members/{uid}
GET    /api/v1/admin/users
GET    /api/v1/admin/users/{id}
PATCH  /api/v1/admin/users/{id}
DELETE /api/v1/admin/users/{id}
GET    /api/v1/admin/events
GET    /api/v1/admin/certificates
GET    /api/v1/admin/batches
GET    /api/v1/admin/audit-log
GET    /api/v1/admin/audit-log/export
GET    /api/v1/admin/stats
GET    /api/v1/admin/profiles
GET    /api/v1/admin/teacher-students
```

### Response Format
```json
{ "data": { ... } }
{ "data": [...], "pagination": { "page": 1, "per_page": 20, "total": 142 } }
{ "error": { "code": "VALIDATION_ERROR", "message": "...", "details": [...] } }
```

### Middleware Chain
```
request → request-id → logger → CORS → rate-limit → JWT auth → role check → handler
```

---

## 7. Docker Compose

### Development
```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: jetistik
      POSTGRES_USER: jetistik
      POSTGRES_PASSWORD: dev-password
    ports: ["5432:5432"]
    volumes: [postgres_data:/var/lib/postgresql/data]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports: ["9000:9000", "9001:9001"]
    volumes: [minio_data:/data]

  gotenberg:
    image: gotenberg/gotenberg:8
    ports: ["3000:3000"]

  backend:
    build: ./backend
    env_file: .env
    ports: ["8080:8080"]
    depends_on: [db, redis, minio, gotenberg]
    volumes: [./backend:/app]

  frontend:
    build: ./frontend
    ports: ["5173:5173"]
    depends_on: [backend]
    volumes: [./frontend:/app, /app/node_modules]

volumes:
  postgres_data:
  minio_data:
```

### Production
```yaml
services:
  db:
    image: postgres:16-alpine
    restart: always
    env_file: .env.prod
    volumes: [postgres_data:/var/lib/postgresql/data]

  redis:
    image: redis:7-alpine
    restart: always

  minio:
    image: minio/minio:latest
    restart: always
    command: server /data
    env_file: .env.prod
    volumes: [minio_data:/data]

  gotenberg:
    image: gotenberg/gotenberg:8
    restart: always

  backend:
    build: ./backend
    restart: always
    env_file: .env.prod
    ports: ["8080:8080"]
    depends_on: [db, redis, minio, gotenberg]

  worker:
    build: ./backend
    restart: always
    command: /app/server --worker
    env_file: .env.prod
    depends_on: [db, redis, minio, gotenberg]

  frontend:
    build: ./frontend
    restart: always
    ports: ["3000:3000"]

volumes:
  postgres_data:
  minio_data:
```

Nginx (via Ansible): `/` → frontend:3000, `/api/` → backend:8080

---

## 8. Frontend Architecture (SvelteKit)

### Route Groups

| Group | Pages | Rendering | Auth |
|-------|-------|-----------|------|
| (public) | Landing, Verify, FAQ, Organizers | SSR | None |
| (auth) | Login, Register, Logout, Password Reset | SSR | None |
| (app)/student | My Certificates, Portfolio QR | SPA | role=student |
| (app)/teacher | My Students, Student Certificates | SPA | role=teacher |
| (app)/staff | Events, Event Detail, Mapping, Generate, Certificates, Audit | SPA | role=staff |
| (app)/admin | Organizations, Users, Events, Templates, Batches, Participants, Certificates, Audit, Profiles, Teacher-Students | SPA | role=admin |

### Key Libraries
- `$lib/api/` — typed API client with JWT refresh logic
- `$lib/components/ui/` — base components (Button, Input, Card, Badge, etc.)
- `$lib/components/layout/` — Sidebar, Topbar, Footer
- `$lib/components/data/` — DataTable, Pagination, Filters, StatusBadge
- `$lib/stores/auth.ts` — user state, tokens, role
- `$lib/stores/i18n.ts` — current language
- `$lib/i18n/` — kz.ts, ru.ts, en.ts translation files

### Design System: "The Sovereign Ledger"

Reference: `stitch/jetistik_sovereign/DESIGN.md`

- **No-Line Rule**: no 1px borders; separation via background tone shifts
- **Fonts**: Manrope (headlines) + Inter (body)
- **Surface hierarchy**: #f7f9fb → #f2f4f6 → #ffffff → #e6e8ea
- **Primary**: #004ac6 → #2563eb gradient (135deg)
- **No pure black**: use #191c1e (on_surface)
- **Corners**: 0.375rem (md) or 0.5rem (lg)
- **Shadows**: ambient, 4% opacity, primary-tinted, 40px blur

---

## 9. Migration Strategy (v1 → v2)

### Data Migration

| Source (v1) | Target (v2) | Notes |
|-------------|-------------|-------|
| Django auth_user | users | Password: PBKDF2 → dual verify, re-hash to bcrypt on first login |
| UserProfile | users.iin | Merged into users table |
| Django Groups | users.role | staff_org→staff, user_teacher→teacher, user_student→student |
| Organization | organizations | + logo files to MinIO |
| OrganizationUser | organization_members | OneToOne → ManyToMany |
| Event | events | Preserve created_by links |
| Template | templates | Files: media/ → MinIO |
| ImportBatch | import_batches | Files: media/ → MinIO |
| ParticipantRow | participant_rows | payload_json preserved |
| Certificate | certificates | 1700+ PDFs: media/ → MinIO (CRITICAL) |
| AuditLog | audit_logs | Full history preserved |
| TeacherStudent | teacher_students | Links preserved |

### Password Migration

Django stores passwords as `pbkdf2_sha256$<iterations>$<salt>$<hash>`. Go backend implements:
1. Dual verifier: try bcrypt first, fallback to Django PBKDF2
2. On successful PBKDF2 login → re-hash to bcrypt and update DB
3. Transparent to user — no password reset required

### File Migration (media/ → MinIO)

```
v1 media/                        → MinIO bucket: jetistik/
├── templates/%Y/%m/*.pptx       → jetistik/templates/...
├── certificates/%Y/%m/*.pdf     → jetistik/certificates/...
├── imports/%Y/%m/*.xlsx/.csv    → jetistik/imports/...
└── orgs/%Y/%m/*.png/.jpg        → jetistik/logos/...
```

Migration script reads FileField records from v1 PostgreSQL, uploads to MinIO, updates paths in v2 DB.

### URL Compatibility

`/verify/{code}` MUST keep the same format — printed QR codes on 1700+ certificates point to these URLs.

### Cutover Plan

1. v2 deployed and tested on staging
2. v1 set to maintenance mode (read-only)
3. Run DB migration script (pg_dump v1 → transform → pg_restore v2)
4. Run file migration script (media/ → MinIO)
5. Verify: record counts, random sample certificate downloads
6. DNS switch to v2
7. Smoke tests: login, verify old certificate, download PDF
8. v1 remains running 48h as fallback

### Error Handling
- File not found in media/ → log warning, skip file, continue migration, generate report of missing files at the end
- Duplicate IIN/code → skip row, log conflict, include in migration report (v1 has no duplicates on code due to UNIQUE constraint, but IIN can repeat across certificates)
- Password verification failure → user uses "forgot password" flow to set new password in v2

### Migration Script Location
- `backend/cmd/migrate-v1/main.go` — standalone Go binary
- Reads from v1 PostgreSQL directly (connection string as flag)
- Writes to v2 PostgreSQL + MinIO
- Outputs JSON report with counts and errors

---

## 10. Roles & Access Control

| Role | Access |
|------|--------|
| admin | Full access. Manage organizations, users, everything |
| staff | Events, templates, batches, certificates of own organization |
| teacher | Own certificates + linked students' certificates |
| student | Own certificates only |
| public | Verify by code, search by IIN (rate-limited 10/min) |

---

## 11. Security

- JWT access token: 15 min TTL, stored in memory (not localStorage)
- JWT refresh token: 7 days TTL, httpOnly secure cookie
- Rate limiting: 10 req/min per IP for public endpoints
- IIN masking: display as `9905****52`
- CORS: frontend domain only
- SQL injection: impossible (sqlc, parameterized queries)
- File uploads: MIME type + size validation
- HTTPS only in production

---

## 12. Environment Variables

```env
# Database
DATABASE_URL=postgres://jetistik:password@db:5432/jetistik

# Redis
REDIS_URL=redis://redis:6379/0

# MinIO
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=jetistik
MINIO_USE_SSL=false

# Gotenberg
GOTENBERG_URL=http://gotenberg:3000

# JWT
JWT_SECRET=change-me
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# App
PUBLIC_BASE_URL=https://jetistik.kz
APP_PORT=8080

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=Jetistik <noreply@jetistik.kz>
ORGANIZER_REQUEST_TO=yerzhan@blackboard.kz
```

---

## 13. Deployment

- **Infrastructure**: Terraform (OpenStack, kz-ast-1 region)
- **Configuration**: Ansible (roles: common, docker, app, nginx)
- **Runtime**: Docker Compose (postgres, redis, minio, gotenberg, backend, worker, frontend)
- **SSL**: Certbot + Cloudflare DNS validation
- **CI**: GitHub Actions (lint, test, build)
