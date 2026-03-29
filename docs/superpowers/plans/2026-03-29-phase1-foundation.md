# Phase 1: Foundation — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reorganize the repo (v1/ folder), create CLAUDE.md, scaffold Go backend + SvelteKit frontend, wire Docker Compose, and establish the full database schema — so all subsequent phases have a working foundation to build on.

**Architecture:** Modular monolith Go backend serving JSON API on :8080, SvelteKit frontend on :5173 (dev) / :3000 (prod). PostgreSQL 16, Redis 7, MinIO, Gotenberg as Docker services. goose for migrations, sqlc for query generation.

**Tech Stack:** Go 1.23+ (chi, pgxpool, goose, sqlc), SvelteKit 2 (TypeScript, Tailwind CSS 4, Svelte 5), PostgreSQL 16, Redis 7, MinIO, Gotenberg 8, Docker Compose.

**Spec:** `docs/superpowers/specs/2026-03-29-jetistik-v2-design.md`

**Phases Overview:**
- **Phase 1 (this plan):** Foundation — repo reorg, scaffolds, Docker, DB schema
- Phase 2: Auth & Users — JWT, profiles, login/register
- Phase 3: Core Business — orgs, events, templates, batches, certificates, staff UI
- Phase 4: Roles & Dashboards — student, teacher, admin
- Phase 5: Workers & Storage — MinIO integration, Asynq, Gotenberg, SSE
- Phase 6: Migration & Deploy — v1→v2 script, Ansible/Terraform

---

## File Map

### Files to Create

```
CLAUDE.md                                    # master instructions for the project
Makefile                                     # dev commands (up, down, migrate, sqlc, etc.)
docker-compose.yml                           # dev: postgres, redis, minio, gotenberg, backend, frontend
docker-compose.prod.yml                      # prod: same + worker, no dev volumes
.env.example                                 # v2 env template (replaces v1 .env.example)

backend/
├── Dockerfile                               # multi-stage Go build
├── go.mod                                   # Go module: github.com/jetistik/backend (or local)
├── go.sum
├── .air.toml                                # hot reload config for air
├── sqlc.yaml                                # sqlc configuration
├── cmd/
│   └── server/
│       └── main.go                          # entry point: config, DB, router, server start
├── internal/
│   └── platform/
│       ├── config/
│       │   └── config.go                    # env-based config struct + loader
│       ├── db/
│       │   └── db.go                        # pgxpool connection + healthcheck
│       ├── response/
│       │   └── response.go                  # JSON response helpers (OK, Error, Paginated)
│       └── middleware/
│           ├── requestid.go                 # X-Request-ID middleware
│           ├── logger.go                    # structured logging middleware (slog)
│           └── cors.go                      # CORS middleware
├── migrations/
│   └── 00001_initial_schema.sql             # full DB schema from spec
└── queries/
    └── health.sql                           # simple health check query

frontend/
├── Dockerfile                               # Node build for SvelteKit
├── package.json
├── svelte.config.js                         # SvelteKit config (adapter-node)
├── vite.config.ts                           # Vite config
├── tailwind.config.ts                       # Tailwind v4 with Sovereign Ledger tokens
├── tsconfig.json
├── src/
│   ├── app.html                             # HTML shell
│   ├── app.css                              # Tailwind imports + design system CSS vars
│   ├── routes/
│   │   ├── +layout.svelte                   # root layout (fonts, global styles)
│   │   └── +page.svelte                     # placeholder landing page
│   └── lib/
│       ├── api/
│       │   └── client.ts                    # fetch wrapper (base URL, error handling)
│       └── i18n/
│           ├── index.ts                     # t() function, language store, switcher
│           ├── kz.ts                        # Kazakh translations (skeleton)
│           ├── ru.ts                        # Russian translations (skeleton)
│           └── en.ts                        # English translations (skeleton)
```

### Files to Move

All existing top-level files/directories (except `.git`, `.gitignore`, `.claude`, `docs/`, `stitch/`) move into `v1/`.

### Files to Modify

```
.gitignore                                   # add Go, SvelteKit, MinIO ignores
```

---

### Task 1: Move existing code into v1/

**Files:**
- Move: all top-level files/dirs (except `.git`, `.gitignore`, `.claude`, `docs/`, `stitch/`) → `v1/`
- Modify: `.gitignore`

- [ ] **Step 1: Create v1/ directory and move files**

```bash
mkdir v1

# Move all v1 files and directories
git mv config v1/
git mv core v1/
git mv deploy v1/
git mv docker-compose.yml v1/
git mv docker-compose.prod.yml v1/
git mv Dockerfile v1/
git mv frontend v1/
git mv jetistik_mvp v1/
git mv manage.py v1/
git mv README.md v1/
git mv requirements.txt v1/
git mv scripts v1/
git mv static v1/
git mv templates v1/
git mv .dockerignore v1/
git mv .env.example v1/
git mv .env.prod.example v1/
```

- [ ] **Step 2: Update .gitignore for v2 stack**

Replace the contents of `.gitignore` with:

```gitignore
# OS
.DS_Store
*.log

# === V1 (legacy Django) ===
v1/media/
v1/staticfiles/
v1/__pycache__/
v1/**/__pycache__/
v1/*.pyc
v1/db.sqlite3
v1/.venv/
v1/venv/
v1/frontend/**/node_modules/
v1/frontend/**/.vite/

# === V2 Backend (Go) ===
backend/tmp/
backend/server

# === V2 Frontend (SvelteKit) ===
frontend/node_modules/
frontend/.svelte-kit/
frontend/build/

# === Infrastructure ===
deploy/terraform/.terraform/
deploy/terraform/.terraform.lock.hcl
deploy/terraform/terraform.tfstate*
deploy/terraform/terraform.tfvars
deploy/ansible/group_vars/jetistik.yml

# === Environment ===
.env
.env.prod
.env.local

# === IDE / Tools ===
.idea/
.vscode/
.claude/

# === Archives ===
*.zip
*.sqlite3
```

- [ ] **Step 3: Verify the move**

Run: `ls -la` at project root
Expected: only `.git/`, `.gitignore`, `.claude/`, `docs/`, `stitch/`, `v1/` remain at top level.

Run: `ls v1/`
Expected: `config/`, `core/`, `deploy/`, `docker-compose.yml`, `docker-compose.prod.yml`, `Dockerfile`, `frontend/`, `jetistik_mvp/`, `manage.py`, `README.md`, `requirements.txt`, `scripts/`, `static/`, `templates/`, `.dockerignore`, `.env.example`, `.env.prod.example`

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "Move existing code to v1/ for v2 rewrite"
```

---

### Task 2: Create CLAUDE.md

**Files:**
- Create: `CLAUDE.md`

- [ ] **Step 1: Write CLAUDE.md**

```markdown
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

## Environment Variables

See .env.example for the full list.
```

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "Add CLAUDE.md with v2 project instructions"
```

---

### Task 3: Create Docker Compose and env files

**Files:**
- Create: `docker-compose.yml`
- Create: `docker-compose.prod.yml`
- Create: `.env.example`

- [ ] **Step 1: Create dev docker-compose.yml**

```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: jetistik
      POSTGRES_USER: jetistik
      POSTGRES_PASSWORD: dev-password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U jetistik"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data

  gotenberg:
    image: gotenberg/gotenberg:8
    ports:
      - "3000:3000"

volumes:
  postgres_data:
  minio_data:
```

Note: backend and frontend services are excluded from dev compose intentionally. During development, run `air` and `npm run dev` locally for faster hot reload. The compose file provides only infrastructure services.

- [ ] **Step 2: Create production docker-compose.prod.yml**

```yaml
services:
  db:
    image: postgres:16-alpine
    restart: always
    env_file: .env.prod
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U jetistik"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    restart: always
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  minio:
    image: minio/minio:latest
    restart: always
    command: server /data
    env_file: .env.prod
    volumes:
      - minio_data:/data

  gotenberg:
    image: gotenberg/gotenberg:8
    restart: always

  backend:
    build: ./backend
    restart: always
    env_file: .env.prod
    ports:
      - "8080:8080"
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
      minio:
        condition: service_started
      gotenberg:
        condition: service_started

  worker:
    build: ./backend
    restart: always
    command: /app/server --worker
    env_file: .env.prod
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
      minio:
        condition: service_started
      gotenberg:
        condition: service_started

  frontend:
    build: ./frontend
    restart: always
    ports:
      - "3001:3000"

volumes:
  postgres_data:
  minio_data:
```

- [ ] **Step 3: Create .env.example**

```env
# Database
DATABASE_URL=postgres://jetistik:dev-password@localhost:5432/jetistik?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379/0

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=jetistik
MINIO_USE_SSL=false

# Gotenberg
GOTENBERG_URL=http://localhost:3000

# JWT
JWT_SECRET=dev-secret-change-me-in-production
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# App
PUBLIC_BASE_URL=http://localhost:5173
APP_PORT=8080
APP_ENV=development

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=Jetistik <noreply@jetistik.kz>
ORGANIZER_REQUEST_TO=yerzhan@blackboard.kz
```

- [ ] **Step 4: Verify compose starts infrastructure**

```bash
docker compose up -d
docker compose ps
```

Expected: db, redis, minio, gotenberg all show "running" or "healthy".

- [ ] **Step 5: Commit**

```bash
git add docker-compose.yml docker-compose.prod.yml .env.example
git commit -m "Add Docker Compose for v2 infrastructure services"
```

---

### Task 4: Scaffold Go backend

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/internal/platform/config/config.go`
- Create: `backend/internal/platform/db/db.go`
- Create: `backend/internal/platform/response/response.go`
- Create: `backend/Dockerfile`
- Create: `backend/.air.toml`

- [ ] **Step 1: Initialize Go module**

```bash
mkdir -p backend/cmd/server
mkdir -p backend/internal/platform/{config,db,response,middleware}
mkdir -p backend/migrations
mkdir -p backend/queries
cd backend && go mod init jetistik && cd ..
```

- [ ] **Step 2: Create config loader**

Create `backend/internal/platform/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppPort      string
	AppEnv       string
	DatabaseURL  string
	RedisURL     string
	PublicBaseURL string

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool

	GotenbergURL string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	SMTPHost   string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	SMTPFrom   string
	OrgRequestTo string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppPort:      envOr("APP_PORT", "8080"),
		AppEnv:       envOr("APP_ENV", "development"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisURL:     envOr("REDIS_URL", "redis://localhost:6379/0"),
		PublicBaseURL: envOr("PUBLIC_BASE_URL", "http://localhost:5173"),

		MinioEndpoint:  envOr("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: envOr("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: envOr("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    envOr("MINIO_BUCKET", "jetistik"),
		MinioUseSSL:    os.Getenv("MINIO_USE_SSL") == "true",

		GotenbergURL: envOr("GOTENBERG_URL", "http://localhost:3000"),

		JWTSecret: os.Getenv("JWT_SECRET"),

		SMTPHost:     envOr("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     envOr("SMTP_PORT", "587"),
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPass:     os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     envOr("SMTP_FROM", "Jetistik <noreply@jetistik.kz>"),
		OrgRequestTo: envOr("ORGANIZER_REQUEST_TO", "yerzhan@blackboard.kz"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	var err error
	cfg.JWTAccessTTL, err = time.ParseDuration(envOr("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	cfg.JWTRefreshTTL, err = time.ParseDuration(envOr("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	return cfg, nil
}

func (c *Config) IsDev() bool {
	return c.AppEnv == "development"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 3: Create database connection**

Create `backend/internal/platform/db/db.go`:

```go
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
```

- [ ] **Step 4: Create response helpers**

Create `backend/internal/platform/response/response.go`:

```go
package response

import (
	"encoding/json"
	"net/http"
)

type envelope struct {
	Data interface{} `json:"data,omitempty"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

type paginatedEnvelope struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Data: data})
}

func Paginated(w http.ResponseWriter, data interface{}, p Pagination) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(paginatedEnvelope{Data: data, Pagination: p})
}

func Error(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorEnvelope{
		Error: errorBody{Code: code, Message: message},
	})
}

func ValidationError(w http.ResponseWriter, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(errorEnvelope{
		Error: errorBody{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: details,
		},
	})
}
```

- [ ] **Step 5: Create main.go entry point**

Create `backend/cmd/server/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/platform/config"
	"jetistik/internal/platform/db"
	"jetistik/internal/platform/response"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	if cfg.IsDev() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
	slog.SetDefault(logger)

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	slog.Info("connected to database")

	r := chi.NewRouter()

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		err := pool.Ping(r.Context())
		if err != nil {
			response.Error(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database is not reachable")
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	return srv.Shutdown(shutdownCtx)
}
```

- [ ] **Step 6: Install Go dependencies**

```bash
cd backend
go get github.com/go-chi/chi/v5
go get github.com/jackc/pgx/v5
go mod tidy
cd ..
```

- [ ] **Step 7: Create backend Dockerfile**

Create `backend/Dockerfile`:

```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/server .
COPY --from=builder /build/migrations ./migrations

EXPOSE 8080
CMD ["/app/server"]
```

- [ ] **Step 8: Create air config for hot reload**

Create `backend/.air.toml`:

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/server ./cmd/server"
  bin = "./tmp/server"
  include_ext = ["go"]
  exclude_dir = ["tmp", "migrations", "queries"]
  delay = 1000

[log]
  time = false

[misc]
  clean_on_exit = true
```

- [ ] **Step 9: Verify backend compiles**

```bash
cd backend && go build ./cmd/server && cd ..
```

Expected: no errors, binary created.

- [ ] **Step 10: Test health endpoint manually**

Ensure Docker Compose is running (`docker compose up -d`), then:

```bash
cd backend
DATABASE_URL="postgres://jetistik:dev-password@localhost:5432/jetistik?sslmode=disable" go run ./cmd/server &
sleep 2
curl -s http://localhost:8080/api/v1/health | python3 -m json.tool
kill %1
cd ..
```

Expected output:
```json
{
    "data": {
        "status": "ok"
    }
}
```

- [ ] **Step 11: Commit**

```bash
git add backend/
git commit -m "Scaffold Go backend with health endpoint"
```

---

### Task 5: Create database migration

**Files:**
- Create: `backend/migrations/00001_initial_schema.sql`
- Create: `backend/sqlc.yaml`
- Create: `backend/queries/health.sql`

- [ ] **Step 1: Install goose and sqlc**

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

Verify both are available:
```bash
goose --version
sqlc version
```

- [ ] **Step 2: Create initial migration**

Create `backend/migrations/00001_initial_schema.sql`:

```sql
-- +goose Up

CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(150) UNIQUE NOT NULL,
    email       VARCHAR(254) UNIQUE,
    password    TEXT NOT NULL,
    iin         VARCHAR(12),
    role        VARCHAR(20) NOT NULL,
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

-- +goose Down

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS teacher_students;
DROP TABLE IF EXISTS certificates;
DROP TABLE IF EXISTS participant_rows;
DROP TABLE IF EXISTS import_batches;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;
```

- [ ] **Step 3: Run migration**

```bash
cd backend
goose -dir migrations postgres "postgres://jetistik:dev-password@localhost:5432/jetistik?sslmode=disable" up
cd ..
```

Expected: `OK    00001_initial_schema.sql`

- [ ] **Step 4: Verify tables exist**

```bash
docker compose exec db psql -U jetistik -c "\dt"
```

Expected: all 10 tables listed (users, organizations, organization_members, events, templates, import_batches, participant_rows, certificates, teacher_students, audit_logs, refresh_tokens).

- [ ] **Step 5: Create sqlc config**

Create `backend/sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/"
    schema: "migrations/"
    gen:
      go:
        package: "sqlcdb"
        out: "internal/sqlcdb"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_empty_slices: true
```

- [ ] **Step 6: Create health query and generate sqlc**

Create `backend/queries/health.sql`:

```sql
-- name: HealthCheck :one
SELECT 1 AS ok;
```

Run sqlc:
```bash
cd backend && sqlc generate && cd ..
```

Expected: `backend/internal/sqlcdb/` directory created with generated Go files.

- [ ] **Step 7: Commit**

```bash
git add backend/migrations/ backend/sqlc.yaml backend/queries/ backend/internal/sqlcdb/
git commit -m "Add database schema migration and sqlc setup"
```

---

### Task 6: Create middleware (request-id, logger, CORS)

**Files:**
- Create: `backend/internal/platform/middleware/requestid.go`
- Create: `backend/internal/platform/middleware/logger.go`
- Create: `backend/internal/platform/middleware/cors.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Create request-id middleware**

Create `backend/internal/platform/middleware/requestid.go`:

```go
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
```

- [ ] **Step 2: Create logger middleware**

Create `backend/internal/platform/middleware/logger.go`:

```go
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", GetRequestID(r.Context()),
		)
	})
}
```

- [ ] **Step 3: Create CORS middleware**

Create `backend/internal/platform/middleware/cors.go`:

```go
package middleware

import (
	"net/http"
)

func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 4: Wire middleware into main.go**

Replace the router section in `backend/cmd/server/main.go`. The full updated file:

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/platform/config"
	"jetistik/internal/platform/db"
	"jetistik/internal/platform/middleware"
	"jetistik/internal/platform/response"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	if cfg.IsDev() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
	slog.SetDefault(logger)

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	slog.Info("connected to database")

	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS(cfg.PublicBaseURL))

	// Routes
	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		err := pool.Ping(r.Context())
		if err != nil {
			response.Error(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database is not reachable")
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	return srv.Shutdown(shutdownCtx)
}
```

- [ ] **Step 5: Install uuid dependency**

```bash
cd backend
go get github.com/google/uuid
go mod tidy
cd ..
```

- [ ] **Step 6: Verify build**

```bash
cd backend && go build ./cmd/server && cd ..
```

Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add backend/
git commit -m "Add request-id, logger, and CORS middleware"
```

---

### Task 7: Create Makefile

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Create Makefile**

```makefile
.PHONY: up down migrate migrate-new sqlc backend frontend test-backend test-frontend lint

DB_URL ?= postgres://jetistik:dev-password@localhost:5432/jetistik?sslmode=disable

## Infrastructure
up:
	docker compose up -d

down:
	docker compose down

## Database
migrate:
	cd backend && goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	cd backend && goose -dir migrations postgres "$(DB_URL)" down

migrate-new:
	cd backend && goose -dir migrations create $(N) sql

migrate-status:
	cd backend && goose -dir migrations postgres "$(DB_URL)" status

## Code generation
sqlc:
	cd backend && sqlc generate

## Development
backend:
	cd backend && air

frontend:
	cd frontend && npm run dev

## Testing
test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npm test

## Linting
lint:
	cd backend && go vet ./...
	cd frontend && npm run check
```

- [ ] **Step 2: Verify make commands**

```bash
make migrate-status
```

Expected: shows migration 00001_initial_schema applied.

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "Add Makefile with dev commands"
```

---

### Task 8: Scaffold SvelteKit frontend

**Files:**
- Create: `frontend/` (SvelteKit project)

- [ ] **Step 1: Create SvelteKit project**

```bash
npx sv create frontend --template minimal --types ts --no-add-ons --no-install
```

If `sv` is not available, use:
```bash
npm create svelte@latest frontend -- --template minimal --types ts
```

Choose: Skeleton project, TypeScript, no additional options.

- [ ] **Step 2: Install dependencies**

```bash
cd frontend
npm install
npm install -D tailwindcss @tailwindcss/vite
npm install -D @fontsource/manrope @fontsource/inter
cd ..
```

- [ ] **Step 3: Configure Tailwind**

Update `frontend/vite.config.ts`:

```ts
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()]
});
```

Create `frontend/src/app.css`:

```css
@import "tailwindcss";
@import "@fontsource/manrope/400.css";
@import "@fontsource/manrope/600.css";
@import "@fontsource/manrope/700.css";
@import "@fontsource/inter/400.css";
@import "@fontsource/inter/500.css";
@import "@fontsource/inter/600.css";

/* Sovereign Ledger Design System */
@theme {
  /* Surface hierarchy */
  --color-surface: #f7f9fb;
  --color-surface-low: #f2f4f6;
  --color-surface-lowest: #ffffff;
  --color-surface-high: #e6e8ea;

  /* Primary */
  --color-primary: #004ac6;
  --color-primary-container: #2563eb;
  --color-primary-fixed: #dbeafe;
  --color-on-primary: #ffffff;
  --color-on-primary-container: #1e3a5f;

  /* Error */
  --color-error: #ba1a1a;
  --color-error-container: #ffdad6;
  --color-on-error-container: #410002;

  /* Text */
  --color-on-surface: #191c1e;
  --color-on-surface-variant: #6b7280;
  --color-outline-variant: #c4c7cc;

  /* Surface tint */
  --color-surface-tint: #004ac6;

  /* Font families */
  --font-display: "Manrope", sans-serif;
  --font-body: "Inter", sans-serif;

  /* Border radius */
  --radius-md: 0.375rem;
  --radius-lg: 0.5rem;
  --radius-full: 9999px;
}

/* Base styles */
body {
  font-family: var(--font-body);
  color: var(--color-on-surface);
  background-color: var(--color-surface);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-display);
}
```

- [ ] **Step 4: Create root layout**

Create `frontend/src/routes/+layout.svelte`:

```svelte
<script>
  import "../app.css";
  let { children } = $props();
</script>

{@render children()}
```

- [ ] **Step 5: Create placeholder landing page**

Replace `frontend/src/routes/+page.svelte`:

```svelte
<div class="min-h-screen bg-surface flex items-center justify-center">
  <div class="text-center">
    <h1 class="text-4xl font-bold font-display text-on-surface mb-4">
      Jetistik
    </h1>
    <p class="text-on-surface-variant font-body text-lg">
      v2 — coming soon
    </p>
    <div class="mt-8 inline-block bg-gradient-to-br from-primary to-primary-container text-on-primary px-6 py-3 rounded-lg font-display font-semibold">
      Sovereign Ledger
    </div>
  </div>
</div>
```

- [ ] **Step 6: Configure adapter-node for production**

```bash
cd frontend && npm install -D @sveltejs/adapter-node && cd ..
```

Update `frontend/svelte.config.js`:

```js
import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			out: 'build'
		})
	}
};

export default config;
```

- [ ] **Step 7: Create frontend Dockerfile**

Create `frontend/Dockerfile`:

```dockerfile
FROM node:22-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:22-alpine

WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000
ENV PORT=3000
CMD ["node", "build"]
```

- [ ] **Step 8: Verify frontend runs**

```bash
cd frontend && npm run dev -- --port 5173 &
sleep 3
curl -s http://localhost:5173 | grep -o "Jetistik"
kill %1
cd ..
```

Expected: "Jetistik" found in HTML output.

- [ ] **Step 9: Verify frontend builds**

```bash
cd frontend && npm run build && cd ..
```

Expected: no errors, `frontend/build/` directory created.

- [ ] **Step 10: Commit**

```bash
git add frontend/
git commit -m "Scaffold SvelteKit frontend with Sovereign Ledger design system"
```

---

### Task 9: Create i18n system

**Files:**
- Create: `frontend/src/lib/i18n/index.ts`
- Create: `frontend/src/lib/i18n/kz.ts`
- Create: `frontend/src/lib/i18n/ru.ts`
- Create: `frontend/src/lib/i18n/en.ts`

- [ ] **Step 1: Create translation files**

Create `frontend/src/lib/i18n/kz.ts`:

```ts
export default {
  "app.name": "Jetistik",
  "app.tagline": "QR-верификациясы бар сертификаттар",
  "nav.home": "Басты бет",
  "nav.howItWorks": "Қалай жұмыс істейді",
  "nav.verification": "Верификация",
  "nav.faq": "FAQ",
  "nav.forOrganizers": "Ұйымдастырушыларға",
  "nav.login": "Кіру",
  "nav.register": "Тіркелу",
  "nav.logout": "Шығу",
  "common.loading": "Жүктелуде...",
  "common.save": "Сақтау",
  "common.cancel": "Болдырмау",
  "common.delete": "Жою",
  "common.edit": "Өзгерту",
  "common.download": "Жүктеу",
  "common.search": "Іздеу",
  "common.filter": "Сүзгі",
  "common.back": "Артқа",
  "landing.title": "Сертификаттарды тауып, бір минутта жүктеп алыңыз",
  "landing.subtitle": "Цифрлық сертификаттарды QR-код арқылы тексеру платформасы",
} as const;
```

Create `frontend/src/lib/i18n/ru.ts`:

```ts
export default {
  "app.name": "Jetistik",
  "app.tagline": "Сертификаты с QR-верификацией",
  "nav.home": "Главная",
  "nav.howItWorks": "Как это работает",
  "nav.verification": "Верификация",
  "nav.faq": "FAQ",
  "nav.forOrganizers": "Организаторам",
  "nav.login": "Войти",
  "nav.register": "Регистрация",
  "nav.logout": "Выйти",
  "common.loading": "Загрузка...",
  "common.save": "Сохранить",
  "common.cancel": "Отмена",
  "common.delete": "Удалить",
  "common.edit": "Редактировать",
  "common.download": "Скачать",
  "common.search": "Поиск",
  "common.filter": "Фильтр",
  "common.back": "Назад",
  "landing.title": "Найдите и скачайте сертификаты за минуту",
  "landing.subtitle": "Платформа верификации цифровых сертификатов через QR-код",
} as const;
```

Create `frontend/src/lib/i18n/en.ts`:

```ts
export default {
  "app.name": "Jetistik",
  "app.tagline": "Certificates with QR Verification",
  "nav.home": "Home",
  "nav.howItWorks": "How it works",
  "nav.verification": "Verification",
  "nav.faq": "FAQ",
  "nav.forOrganizers": "For Organizers",
  "nav.login": "Log in",
  "nav.register": "Register",
  "nav.logout": "Log out",
  "common.loading": "Loading...",
  "common.save": "Save",
  "common.cancel": "Cancel",
  "common.delete": "Delete",
  "common.edit": "Edit",
  "common.download": "Download",
  "common.search": "Search",
  "common.filter": "Filter",
  "common.back": "Back",
  "landing.title": "Find and download your certificates in a minute",
  "landing.subtitle": "Digital certificate verification platform via QR code",
} as const;
```

- [ ] **Step 2: Create i18n store and t() function**

Create `frontend/src/lib/i18n/index.ts`:

```ts
import { writable, derived } from "svelte/store";
import { browser } from "$app/environment";
import kz from "./kz";
import ru from "./ru";
import en from "./en";

export type Language = "kz" | "ru" | "en";
export type TranslationKey = keyof typeof kz;

const translations: Record<Language, Record<string, string>> = { kz, ru, en };

function getInitialLanguage(): Language {
  if (browser) {
    const stored = localStorage.getItem("jetistik-lang");
    if (stored && stored in translations) return stored as Language;
  }
  return "kz";
}

export const language = writable<Language>(getInitialLanguage());

if (browser) {
  language.subscribe((lang) => {
    localStorage.setItem("jetistik-lang", lang);
    document.documentElement.lang = lang === "kz" ? "kk" : lang;
  });
}

export const t = derived(language, ($lang) => {
  return (key: TranslationKey): string => {
    return translations[$lang]?.[key] ?? translations.kz[key] ?? key;
  };
});

export function setLanguage(lang: Language) {
  language.set(lang);
}
```

- [ ] **Step 3: Verify i18n imports compile**

```bash
cd frontend && npm run check && cd ..
```

Expected: no type errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/i18n/
git commit -m "Add i18n system with KZ/RU/EN translations"
```

---

### Task 10: Create API client

**Files:**
- Create: `frontend/src/lib/api/client.ts`

- [ ] **Step 1: Create API client**

Create `frontend/src/lib/api/client.ts`:

```ts
const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: unknown
  ) {
    super(message);
    this.name = "ApiError";
  }
}

interface ApiResponse<T> {
  data: T;
}

interface ApiErrorResponse {
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
}

interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    per_page: number;
    total: number;
  };
}

let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken(): string | null {
  return accessToken;
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as ApiErrorResponse | null;
    throw new ApiError(
      res.status,
      body?.error?.code ?? "UNKNOWN",
      body?.error?.message ?? res.statusText,
      body?.error?.details
    );
  }

  if (res.status === 204) return undefined as T;

  return res.json() as Promise<T>;
}

export const api = {
  get<T>(path: string): Promise<ApiResponse<T>> {
    return request(path);
  },

  post<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return request(path, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    });
  },

  patch<T>(path: string, body: unknown): Promise<ApiResponse<T>> {
    return request(path, {
      method: "PATCH",
      body: JSON.stringify(body),
    });
  },

  delete<T>(path: string): Promise<ApiResponse<T>> {
    return request(path, { method: "DELETE" });
  },

  upload<T>(path: string, formData: FormData): Promise<ApiResponse<T>> {
    return request(path, {
      method: "POST",
      body: formData,
      headers: {},
    });
  },
};

export type { ApiResponse, PaginatedResponse };
```

- [ ] **Step 2: Verify it compiles**

```bash
cd frontend && npm run check && cd ..
```

Expected: no type errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/api/
git commit -m "Add typed API client with JWT support"
```

---

### Task 11: End-to-end smoke test

**Files:** none new — validates everything works together.

- [ ] **Step 1: Ensure Docker Compose is running**

```bash
docker compose up -d
docker compose ps
```

Expected: db, redis, minio, gotenberg all healthy/running.

- [ ] **Step 2: Run migration**

```bash
make migrate
```

Expected: migration applied (or already applied).

- [ ] **Step 3: Start backend**

```bash
cd backend
DATABASE_URL="postgres://jetistik:dev-password@localhost:5432/jetistik?sslmode=disable" go run ./cmd/server &
sleep 2
```

- [ ] **Step 4: Test health endpoint**

```bash
curl -s http://localhost:8080/api/v1/health
```

Expected: `{"data":{"status":"ok"}}`

- [ ] **Step 5: Check CORS headers**

```bash
curl -s -I -X OPTIONS http://localhost:8080/api/v1/health \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: GET"
```

Expected: `Access-Control-Allow-Origin: http://localhost:5173` in response headers.

- [ ] **Step 6: Check request-id header**

```bash
curl -s -I http://localhost:8080/api/v1/health | grep X-Request-ID
```

Expected: `X-Request-ID: <uuid>` present.

- [ ] **Step 7: Start frontend**

```bash
kill %1  # stop backend
cd frontend && npm run dev -- --port 5173 &
sleep 3
curl -s http://localhost:5173 | grep "Sovereign Ledger"
kill %1
cd ..
```

Expected: "Sovereign Ledger" found in HTML.

- [ ] **Step 8: Verify build artifacts**

```bash
cd backend && go build -o /dev/null ./cmd/server && cd ..
cd frontend && npm run build && cd ..
```

Expected: both build without errors.

- [ ] **Step 9: Final commit with any cleanup**

If any files changed during testing:
```bash
git add -A
git commit -m "Foundation phase complete: backend + frontend scaffolds verified"
```

---

## Summary

After completing all 11 tasks, the project has:

1. **v1/ folder** — all legacy code preserved with git history
2. **CLAUDE.md** — master instructions for the project
3. **Docker Compose** — dev (db, redis, minio, gotenberg) + prod (+ backend, worker, frontend)
4. **Go backend** — compiling, serving /health, with middleware (request-id, logger, CORS)
5. **Database** — full schema applied via goose, sqlc configured and generating
6. **SvelteKit frontend** — running with Sovereign Ledger design system, Tailwind, i18n (KZ/RU/EN)
7. **API client** — typed fetch wrapper with JWT support
8. **Makefile** — all common dev commands

**Next phase:** Phase 2 — Auth & Users (JWT login/register, user profiles, password hashing, role middleware).
