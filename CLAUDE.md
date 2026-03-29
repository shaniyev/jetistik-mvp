# Jetistik v2

Platform for generating and verifying digital certificates with QR codes.
Organizations upload PPTX templates + participant data (CSV/XLSX),
the system generates personalized PDF certificates with QR verification.

## Stack

- **Backend:** Go 1.23+, chi (router), sqlc (queries), goose (migrations), pgxpool (PostgreSQL)
- **Frontend:** SvelteKit, Tailwind CSS, TypeScript
- **Database:** PostgreSQL 16
- **File Storage:** MinIO (S3-compatible)
- **Task Queue:** Asynq (Redis-based)
- **PDF Generation:** Gotenberg (PPTX -> PDF conversion)
- **Auth:** JWT (access token in memory, refresh token in httpOnly cookie)
- **Real-time:** SSE (Server-Sent Events) for generation progress
- **I18n:** Kazakh, Russian, English

## Project Structure

```
v1/                    # legacy Django MVP (read-only reference)
backend/               # Go modular monolith
  cmd/server/          # entry point
  internal/            # modules (auth, event, certificate, etc.)
  migrations/          # goose SQL
  queries/             # sqlc SQL
frontend/              # SvelteKit
  src/routes/          # file-based routing
  src/lib/             # components, API client, stores, i18n
stitch/                # design concepts (reference)
deploy/                # Terraform + Ansible
```

## Backend Architecture

Modular monolith. Each module in `internal/`:
```
module/
  handler.go      # HTTP handlers
  service.go      # business logic
  repository.go   # data access interface
  dto.go          # request/response structs
```

Dependencies: handler -> service -> repository.
Inter-module communication through interfaces.
Wiring in cmd/server/main.go.

## Dev Commands

```bash
make up              # start all Docker services
make down            # stop all services
make migrate         # run goose migrations
make migrate-new N=  # create new migration: make migrate-new N=add_users
make sqlc            # regenerate sqlc code
make backend         # run Go backend with air (hot reload)
make frontend        # run SvelteKit dev server
make test-backend    # run Go tests
make test-frontend   # run frontend tests
make lint            # run linters
```

## Code Conventions

### Go
- Standard gofmt/goimports
- Errors: return, don't panic. `if err != nil { return fmt.Errorf("context: %w", err) }`
- context.Context is the first argument in service/repository functions
- Logging: slog (structured)
- HTTP responses: `{"data": ...}` success, `{"error": {"code": "...", "message": "..."}}` errors
- Validation on handler layer via dto struct tags

### SvelteKit
- TypeScript strict mode
- Components: PascalCase files (StatusBadge.svelte)
- Styles: Tailwind utility classes, "Sovereign Ledger" design system
- API calls through $lib/api/ -- never raw fetch
- Stores for global state (auth, i18n)
- SSR for public pages (SEO), SPA for dashboards

### SQL
- snake_case for tables and columns
- Migrations: YYYYMMDDHHMMSS_description.sql
- sqlc queries: GetUserByID, ListCertificatesByIIN, CreateEvent

## Design System

Reference: stitch/jetistik_sovereign/DESIGN.md

Key rules:
- No-Line Rule: no 1px borders, separation via background tone shifts
- Fonts: Manrope (headlines) + Inter (body)
- Surface hierarchy: #f7f9fb -> #f2f4f6 -> #ffffff -> #e6e8ea
- Primary: #004ac6 -> #2563eb gradient (135deg)
- No pure black: use #191c1e (on_surface)
- Corners: 0.375rem (md) or 0.5rem (lg)
- Shadows: ambient, 4% opacity, primary-tinted, 40px blur

## Roles

| Role    | Access                                             |
|---------|----------------------------------------------------|
| admin   | Full access. Manage orgs, users, everything        |
| staff   | Events, templates, batches, certs of own org       |
| teacher | Own certs + linked students' certs                 |
| student | Own certs only                                     |
| public  | Verify by code, search by IIN (rate-limited)       |

## v1 Reference

v1/ contains the legacy Django code. Key files for porting business logic:
- v1/core/models.py -- data schema
- v1/core/views.py -- all business logic
- v1/core/utils.py -- PPTX processing, QR generation, PDF conversion
- v1/core/tasks.py -- Celery async tasks

URL /verify/{code} MUST keep the same format (old QR codes on printed certificates).

## Security

- JWT access: 15 min, in memory (not localStorage)
- JWT refresh: 7 days, httpOnly secure cookie
- Rate limiting: 10 req/min per IP for public endpoints
- IIN masking: display as 9905****52
- CORS: frontend domain only
- File uploads: validate MIME type + size
- HTTPS only in production

## Deploy

```bash
# v2 deploy (from project root)
cd deploy/ansible
ansible-playbook -i inventory.ini playbook.yml

# v1 migration (standalone binary)
cd backend
go build -o migrate-v1 ./cmd/migrate-v1
./migrate-v1 \
  --source-db "postgres://user:pass@v1-host:5432/jetistik" \
  --target-db "postgres://user:pass@v2-host:5432/jetistik" \
  --minio-endpoint "localhost:9000" \
  --minio-key "minioadmin" \
  --minio-secret "minioadmin" \
  --minio-bucket "jetistik" \
  --media-dir "/path/to/v1/media"

# Dry run (no writes, just prints plan)
./migrate-v1 --dry-run ...
```

Deploy config: `deploy/ansible/group_vars/jetistik-v2.yml.example`
Production env template: `deploy/ansible/roles/app/templates/env.prod.j2`

## Environment Variables

See .env.example for the full list.
