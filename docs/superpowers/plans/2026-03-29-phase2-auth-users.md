# Phase 2: Auth & Users — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement JWT authentication (access + refresh tokens), user registration (student/teacher + organization), login/logout, password hashing (bcrypt + Django PBKDF2 migration verifier), role-based middleware, rate limiting, user profile CRUD, teacher-student links, and frontend auth pages — so all subsequent phases can build on a fully working auth system.

**Architecture:** Auth module (`internal/auth/`) handles JWT issuance, password verification, and auth endpoints. User module (`internal/user/`) handles profile CRUD and teacher-student links. Both follow handler -> service -> repository pattern. JWT middleware extracts user claims and injects them into context. Role middleware checks claims against required roles. Rate limiter uses in-memory token bucket per IP.

**Tech Stack:** Go 1.24+ (chi, pgxpool, sqlc, golang-jwt/jwt/v5, golang.org/x/crypto/bcrypt), SvelteKit 2 (TypeScript, Tailwind CSS 4, Svelte 5), PostgreSQL 16.

**Spec:** `docs/superpowers/specs/2026-03-29-jetistik-v2-design.md`

**Phases Overview:**
- Phase 1: Foundation — repo reorg, scaffolds, Docker, DB schema (done)
- **Phase 2 (this plan):** Auth & Users — JWT, profiles, login/register
- Phase 3: Core Business — orgs, events, templates, batches, certificates, staff UI
- Phase 4: Roles & Dashboards — student, teacher, admin
- Phase 5: Workers & Storage — MinIO integration, Asynq, Gotenberg, SSE
- Phase 6: Migration & Deploy — v1->v2 script, Ansible/Terraform

---

## File Map

### Files to Create

```
backend/
├── internal/
│   ├── auth/
│   │   ├── handler.go          # HTTP handlers: login, register, refresh, logout
│   │   ├── service.go          # business logic: token issuance, password verify
│   │   ├── repository.go       # data access interface
│   │   ├── dto.go              # request/response structs
│   │   ├── jwt.go              # JWT creation + parsing helpers
│   │   ├── password.go         # bcrypt hash + Django PBKDF2 verifier
│   │   └── password_test.go    # password hashing tests
│   ├── user/
│   │   ├── handler.go          # HTTP handlers: profile CRUD, teacher-student
│   │   ├── service.go          # business logic
│   │   ├── repository.go       # data access interface
│   │   └── dto.go              # request/response structs
│   └── platform/
│       └── middleware/
│           ├── auth.go         # JWT extraction middleware
│           ├── role.go         # role-based access control middleware
│           └── ratelimit.go    # in-memory rate limiter middleware
├── queries/
│   ├── users.sql               # sqlc queries for users table
│   └── refresh_tokens.sql      # sqlc queries for refresh_tokens table

frontend/
├── src/
│   ├── lib/
│   │   ├── stores/
│   │   │   └── auth.ts         # auth store: user state, tokens, role
│   │   └── api/
│   │       └── client.ts       # (modify) add token refresh interceptor
│   ├── routes/
│   │   ├── (auth)/
│   │   │   ├── +layout.svelte  # auth pages layout (centered card)
│   │   │   ├── login/
│   │   │   │   └── +page.svelte  # login page
│   │   │   ├── register/
│   │   │   │   └── +page.svelte  # register page
│   │   │   └── logout/
│   │   │       └── +page.svelte  # logout page
│   │   └── (app)/
│   │       └── +layout.svelte    # protected layout (auth guard)
```

### Files to Modify

```
backend/cmd/server/main.go          # wire auth + user modules, add routes
backend/go.mod                       # add golang-jwt/jwt/v5 dependency
frontend/src/lib/api/client.ts       # add refresh token interceptor
frontend/src/lib/i18n/kz.ts          # add auth translation keys
frontend/src/lib/i18n/ru.ts          # add auth translation keys
frontend/src/lib/i18n/en.ts          # add auth translation keys
```

---

### Task 1: Add Go dependencies for JWT and password hashing

**Files:**
- Modify: `backend/go.mod`

- [ ] **Step 1: Install golang-jwt**

```bash
cd backend
go get github.com/golang-jwt/jwt/v5
go mod tidy
cd ..
```

Note: `golang.org/x/crypto` is already an indirect dependency (from pgx). It will become direct when we import `bcrypt` and `pbkdf2`.

- [ ] **Step 2: Verify build still works**

```bash
cd backend && go build ./cmd/server && cd ..
```

- [ ] **Step 3: Commit**

```bash
git add backend/go.mod backend/go.sum
git commit -m "Add golang-jwt dependency for auth module"
```

---

### Task 2: Create sqlc queries for users and refresh tokens

**Files:**
- Create: `backend/queries/users.sql`
- Create: `backend/queries/refresh_tokens.sql`

- [ ] **Step 1: Create users queries**

Create `backend/queries/users.sql`:

```sql
-- name: CreateUser :one
INSERT INTO users (username, email, password, iin, role, language)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, username, email, password, iin, role, is_active, language, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, username, email, password, iin, role, is_active, language, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, password, iin, role, is_active, language, created_at, updated_at
FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, password, iin, role, is_active, language, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByIIN :one
SELECT id, username, email, password, iin, role, is_active, language, created_at, updated_at
FROM users
WHERE iin = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password = $2, updated_at = now() WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET email = COALESCE(sqlc.narg('email'), email),
    iin = COALESCE(sqlc.narg('iin'), iin),
    language = COALESCE(sqlc.narg('language'), language),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, username, email, iin, role, is_active, language, created_at, updated_at;

-- name: ListUsers :many
SELECT id, username, email, iin, role, is_active, language, created_at, updated_at
FROM users
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT count(*) FROM users;

-- name: UsernameExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);
```

- [ ] **Step 2: Create refresh_tokens queries**

Create `backend/queries/refresh_tokens.sql`:

```sql
-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_hash, expires_at, created_at;

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, created_at
FROM refresh_tokens
WHERE token_hash = $1 AND expires_at > now();

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE token_hash = $1;

-- name: DeleteRefreshTokensByUser :exec
DELETE FROM refresh_tokens WHERE user_id = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens WHERE expires_at <= now();
```

- [ ] **Step 3: Generate sqlc code**

```bash
cd backend && sqlc generate && cd ..
```

Expected: `backend/internal/sqlcdb/` updated with `users.sql.go` and `refresh_tokens.sql.go`.

- [ ] **Step 4: Verify generated code compiles**

```bash
cd backend && go build ./... && cd ..
```

- [ ] **Step 5: Commit**

```bash
git add backend/queries/ backend/internal/sqlcdb/
git commit -m "Add sqlc queries for users and refresh tokens"
```

---

### Task 3: Create password hashing module (bcrypt + Django PBKDF2)

**Files:**
- Create: `backend/internal/auth/password.go`
- Create: `backend/internal/auth/password_test.go`

- [ ] **Step 1: Create password.go**

Create `backend/internal/auth/password.go`:

```go
package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

const bcryptCost = 12

// HashPassword creates a bcrypt hash of the given password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword checks a password against a stored hash.
// It supports both bcrypt hashes and Django PBKDF2 hashes
// (format: pbkdf2_sha256$<iterations>$<salt>$<hash>).
// Returns (matches bool, needsRehash bool).
func VerifyPassword(password, storedHash string) (bool, bool) {
	// Try bcrypt first
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err == nil {
		return true, false
	}

	// Try Django PBKDF2
	if strings.HasPrefix(storedHash, "pbkdf2_sha256$") {
		if verifyDjangoPBKDF2(password, storedHash) {
			return true, true // matches, but needs rehash to bcrypt
		}
	}

	return false, false
}

// verifyDjangoPBKDF2 verifies a password against a Django-format PBKDF2 hash.
// Django format: pbkdf2_sha256$<iterations>$<salt>$<base64_hash>
func verifyDjangoPBKDF2(password, djangoHash string) bool {
	parts := strings.SplitN(djangoHash, "$", 4)
	if len(parts) != 4 {
		return false
	}

	algorithm := parts[0]
	if algorithm != "pbkdf2_sha256" {
		return false
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}

	salt := parts[2]
	expectedHash := parts[3]

	derived := pbkdf2.Key([]byte(password), []byte(salt), iterations, sha256.Size, sha256.New)
	computedHash := base64.StdEncoding.EncodeToString(derived)

	return computedHash == expectedHash
}
```

- [ ] **Step 2: Create password_test.go**

Create `backend/internal/auth/password_test.go`:

```go
package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == "testpassword123" {
		t.Fatal("HashPassword returned plaintext")
	}
}

func TestVerifyPassword_Bcrypt(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	matches, needsRehash := VerifyPassword("mypassword", hash)
	if !matches {
		t.Error("expected password to match")
	}
	if needsRehash {
		t.Error("bcrypt hash should not need rehash")
	}

	matches, _ = VerifyPassword("wrongpassword", hash)
	if matches {
		t.Error("expected wrong password to not match")
	}
}

func TestVerifyPassword_DjangoPBKDF2(t *testing.T) {
	// This is a real Django-generated hash for password "testpass123"
	// Generated with: make_password("testpass123") in Django 4.2
	// Algorithm: pbkdf2_sha256, iterations: 720000, salt: test_salt_value
	// We generate a known hash for testing instead of using a Django-generated one.
	password := "testpass123"

	// Create a Django-format hash manually for testing
	djangoHash := createTestDjangoHash(password, "testsalt123", 260000)

	matches, needsRehash := VerifyPassword(password, djangoHash)
	if !matches {
		t.Error("expected Django PBKDF2 password to match")
	}
	if !needsRehash {
		t.Error("Django PBKDF2 hash should need rehash to bcrypt")
	}

	matches, _ = VerifyPassword("wrongpassword", djangoHash)
	if matches {
		t.Error("expected wrong password to not match Django hash")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	matches, needsRehash := VerifyPassword("password", "invalid_hash")
	if matches {
		t.Error("expected invalid hash to not match")
	}
	if needsRehash {
		t.Error("expected invalid hash to not need rehash")
	}
}

// createTestDjangoHash creates a Django-compatible PBKDF2 hash for testing.
func createTestDjangoHash(password, salt string, iterations int) string {
	import (
		"crypto/sha256"
		"encoding/base64"
		"fmt"

		"golang.org/x/crypto/pbkdf2"
	)

	derived := pbkdf2.Key([]byte(password), []byte(salt), iterations, sha256.Size, sha256.New)
	hash := base64.StdEncoding.EncodeToString(derived)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", iterations, salt, hash)
}
```

Wait — Go does not allow imports inside functions. Fix the test helper:

Create `backend/internal/auth/password_test.go`:

```go
package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	"golang.org/x/crypto/pbkdf2"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == "testpassword123" {
		t.Fatal("HashPassword returned plaintext")
	}
}

func TestVerifyPassword_Bcrypt(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	matches, needsRehash := VerifyPassword("mypassword", hash)
	if !matches {
		t.Error("expected password to match")
	}
	if needsRehash {
		t.Error("bcrypt hash should not need rehash")
	}

	matches, _ = VerifyPassword("wrongpassword", hash)
	if matches {
		t.Error("expected wrong password to not match")
	}
}

func TestVerifyPassword_DjangoPBKDF2(t *testing.T) {
	password := "testpass123"
	djangoHash := createTestDjangoHash(password, "testsalt123", 260000)

	matches, needsRehash := VerifyPassword(password, djangoHash)
	if !matches {
		t.Error("expected Django PBKDF2 password to match")
	}
	if !needsRehash {
		t.Error("Django PBKDF2 hash should need rehash to bcrypt")
	}

	matches, _ = VerifyPassword("wrongpassword", djangoHash)
	if matches {
		t.Error("expected wrong password to not match Django hash")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	matches, needsRehash := VerifyPassword("password", "invalid_hash")
	if matches {
		t.Error("expected invalid hash to not match")
	}
	if needsRehash {
		t.Error("expected invalid hash to not need rehash")
	}
}

func createTestDjangoHash(password, salt string, iterations int) string {
	derived := pbkdf2.Key([]byte(password), []byte(salt), iterations, sha256.Size, sha256.New)
	hash := base64.StdEncoding.EncodeToString(derived)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", iterations, salt, hash)
}
```

- [ ] **Step 3: Run tests**

```bash
cd backend && go test ./internal/auth/ -v && cd ..
```

Expected: all 4 tests pass.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/auth/password.go backend/internal/auth/password_test.go
git commit -m "Add password hashing with bcrypt and Django PBKDF2 verifier"
```

---

### Task 4: Create JWT helpers

**Files:**
- Create: `backend/internal/auth/jwt.go`

- [ ] **Step 1: Create jwt.go**

Create `backend/internal/auth/jwt.go`:

```go
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the JWT payload for access tokens.
type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"sub"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair contains an access token and a raw refresh token.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// GenerateAccessToken creates a signed JWT access token.
func GenerateAccessToken(userID int64, username, role, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "jetistik",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// ParseAccessToken validates and parses a JWT access token.
func ParseAccessToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GenerateRefreshToken creates a cryptographically random refresh token string.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashRefreshToken creates a SHA-256 hash of a refresh token for storage.
func HashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/auth/ && cd ..
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/auth/jwt.go
git commit -m "Add JWT access and refresh token helpers"
```

---

### Task 5: Create auth DTOs

**Files:**
- Create: `backend/internal/auth/dto.go`

- [ ] **Step 1: Create dto.go**

Create `backend/internal/auth/dto.go`:

```go
package auth

import "time"

// --- Requests ---

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IIN      string `json:"iin"`
	Role     string `json:"role"`
	Language string `json:"language"`
}

type RegisterOrgRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	OrgName  string `json:"org_name"`
	Language string `json:"language"`
}

// --- Responses ---

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	IIN       string    `json:"iin,omitempty"`
	Role      string    `json:"role"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Validation ---

func (r LoginRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Username == "" {
		errs["username"] = "username is required"
	}
	if r.Password == "" {
		errs["password"] = "password is required"
	}
	return errs
}

func (r RegisterRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Username == "" {
		errs["username"] = "username is required"
	}
	if len(r.Username) < 3 {
		errs["username"] = "username must be at least 3 characters"
	}
	if r.Password == "" {
		errs["password"] = "password is required"
	}
	if len(r.Password) < 8 {
		errs["password"] = "password must be at least 8 characters"
	}
	if r.Role == "" {
		errs["role"] = "role is required"
	}
	if r.Role != "" && r.Role != "student" && r.Role != "teacher" {
		errs["role"] = "role must be student or teacher"
	}
	if r.IIN != "" && len(r.IIN) != 12 {
		errs["iin"] = "IIN must be exactly 12 digits"
	}
	if r.Language == "" {
		r.Language = "kz"
	}
	return errs
}

func (r RegisterOrgRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Username == "" {
		errs["username"] = "username is required"
	}
	if len(r.Username) < 3 {
		errs["username"] = "username must be at least 3 characters"
	}
	if r.Email == "" {
		errs["email"] = "email is required"
	}
	if r.Password == "" {
		errs["password"] = "password is required"
	}
	if len(r.Password) < 8 {
		errs["password"] = "password must be at least 8 characters"
	}
	if r.OrgName == "" {
		errs["org_name"] = "organization name is required"
	}
	return errs
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/auth/dto.go
git commit -m "Add auth request/response DTOs with validation"
```

---

### Task 6: Create auth repository

**Files:**
- Create: `backend/internal/auth/repository.go`

- [ ] **Step 1: Create repository.go**

Create `backend/internal/auth/repository.go`:

```go
package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Repository defines data access for auth operations.
type Repository interface {
	CreateUser(ctx context.Context, params sqlcdb.CreateUserParams) (sqlcdb.User, error)
	GetUserByUsername(ctx context.Context, username string) (sqlcdb.User, error)
	GetUserByID(ctx context.Context, id int64) (sqlcdb.User, error)
	UpdateUserPassword(ctx context.Context, id int64, password string) error
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateRefreshToken(ctx context.Context, params sqlcdb.CreateRefreshTokenParams) (sqlcdb.RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, hash string) (sqlcdb.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, hash string) error
	DeleteRefreshTokensByUser(ctx context.Context, userID int64) error
}

type pgRepository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new auth repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{q: sqlcdb.New(pool)}
}

func (r *pgRepository) CreateUser(ctx context.Context, params sqlcdb.CreateUserParams) (sqlcdb.User, error) {
	user, err := r.q.CreateUser(ctx, params)
	if err != nil {
		return sqlcdb.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *pgRepository) GetUserByUsername(ctx context.Context, username string) (sqlcdb.User, error) {
	user, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		return sqlcdb.User{}, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (r *pgRepository) GetUserByID(ctx context.Context, id int64) (sqlcdb.User, error) {
	user, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return sqlcdb.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *pgRepository) UpdateUserPassword(ctx context.Context, id int64, password string) error {
	err := r.q.UpdateUserPassword(ctx, sqlcdb.UpdateUserPasswordParams{
		ID:       id,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

func (r *pgRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	exists, err := r.q.UsernameExists(ctx, username)
	if err != nil {
		return false, fmt.Errorf("check username exists: %w", err)
	}
	return exists, nil
}

func (r *pgRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	exists, err := r.q.EmailExists(ctx, pgtype.Text{String: email, Valid: email != ""})
	if err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}
	return exists, nil
}

func (r *pgRepository) CreateRefreshToken(ctx context.Context, params sqlcdb.CreateRefreshTokenParams) (sqlcdb.RefreshToken, error) {
	token, err := r.q.CreateRefreshToken(ctx, params)
	if err != nil {
		return sqlcdb.RefreshToken{}, fmt.Errorf("create refresh token: %w", err)
	}
	return token, nil
}

func (r *pgRepository) GetRefreshTokenByHash(ctx context.Context, hash string) (sqlcdb.RefreshToken, error) {
	token, err := r.q.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return sqlcdb.RefreshToken{}, fmt.Errorf("get refresh token: %w", err)
	}
	return token, nil
}

func (r *pgRepository) DeleteRefreshToken(ctx context.Context, hash string) error {
	err := r.q.DeleteRefreshToken(ctx, hash)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

func (r *pgRepository) DeleteRefreshTokensByUser(ctx context.Context, userID int64) error {
	err := r.q.DeleteRefreshTokensByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("delete user refresh tokens: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/auth/ && cd ..
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/auth/repository.go
git commit -m "Add auth repository with user and refresh token data access"
```

---

### Task 7: Create auth service

**Files:**
- Create: `backend/internal/auth/service.go`

- [ ] **Step 1: Create service.go**

Create `backend/internal/auth/service.go`:

```go
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"jetistik/internal/sqlcdb"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotActive      = errors.New("user account is not active")
	ErrUsernameExists     = errors.New("username already taken")
	ErrEmailExists        = errors.New("email already taken")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

// Service handles auth business logic.
type Service struct {
	repo      Repository
	secret    string
	accessTTL time.Duration
	refreshTTL time.Duration
}

// NewService creates a new auth service.
func NewService(repo Repository, secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Login authenticates a user and returns a token pair.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, string, error) {
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("login: %w", err)
	}

	if !user.IsActive.Bool {
		return nil, "", ErrUserNotActive
	}

	matches, needsRehash := VerifyPassword(req.Password, user.Password)
	if !matches {
		return nil, "", ErrInvalidCredentials
	}

	// Rehash Django PBKDF2 to bcrypt on successful login
	if needsRehash {
		newHash, err := HashPassword(req.Password)
		if err != nil {
			slog.Error("failed to rehash password", "user_id", user.ID, "error", err)
		} else {
			if err := s.repo.UpdateUserPassword(ctx, user.ID, newHash); err != nil {
				slog.Error("failed to update rehashed password", "user_id", user.ID, "error", err)
			} else {
				slog.Info("rehashed Django password to bcrypt", "user_id", user.ID)
			}
		}
	}

	return s.issueTokens(ctx, user)
}

// Register creates a new user and returns a token pair.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, string, error) {
	exists, err := s.repo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, "", fmt.Errorf("register: %w", err)
	}
	if exists {
		return nil, "", ErrUsernameExists
	}

	if req.Email != "" {
		emailExists, err := s.repo.EmailExists(ctx, req.Email)
		if err != nil {
			return nil, "", fmt.Errorf("register: %w", err)
		}
		if emailExists {
			return nil, "", ErrEmailExists
		}
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, "", fmt.Errorf("register: %w", err)
	}

	lang := req.Language
	if lang == "" {
		lang = "kz"
	}

	user, err := s.repo.CreateUser(ctx, sqlcdb.CreateUserParams{
		Username: req.Username,
		Email:    pgtype.Text{String: req.Email, Valid: req.Email != ""},
		Password: hashedPassword,
		Iin:      pgtype.Text{String: req.IIN, Valid: req.IIN != ""},
		Role:     req.Role,
		Language: pgtype.Text{String: lang, Valid: true},
	})
	if err != nil {
		return nil, "", fmt.Errorf("register: %w", err)
	}

	return s.issueTokens(ctx, user)
}

// Refresh validates a refresh token and issues a new token pair.
func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*AuthResponse, string, error) {
	hash := HashRefreshToken(rawRefreshToken)

	storedToken, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", ErrInvalidRefreshToken
		}
		return nil, "", fmt.Errorf("refresh: %w", err)
	}

	// Delete the used refresh token (rotation)
	if err := s.repo.DeleteRefreshToken(ctx, hash); err != nil {
		slog.Error("failed to delete used refresh token", "error", err)
	}

	user, err := s.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, "", fmt.Errorf("refresh: %w", err)
	}

	if !user.IsActive.Bool {
		return nil, "", ErrUserNotActive
	}

	return s.issueTokens(ctx, user)
}

// Logout deletes the refresh token.
func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}
	hash := HashRefreshToken(rawRefreshToken)
	return s.repo.DeleteRefreshToken(ctx, hash)
}

// issueTokens creates a new access + refresh token pair for a user.
func (s *Service) issueTokens(ctx context.Context, user sqlcdb.User) (*AuthResponse, string, error) {
	accessToken, err := GenerateAccessToken(user.ID, user.Username, user.Role, s.secret, s.accessTTL)
	if err != nil {
		return nil, "", fmt.Errorf("issue tokens: %w", err)
	}

	rawRefresh, err := GenerateRefreshToken()
	if err != nil {
		return nil, "", fmt.Errorf("issue tokens: %w", err)
	}

	refreshHash := HashRefreshToken(rawRefresh)
	_, err = s.repo.CreateRefreshToken(ctx, sqlcdb.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(s.refreshTTL), Valid: true},
	})
	if err != nil {
		return nil, "", fmt.Errorf("issue tokens: %w", err)
	}

	resp := &AuthResponse{
		AccessToken: accessToken,
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email.String,
			IIN:       maskIIN(user.Iin.String),
			Role:      user.Role,
			Language:  user.Language.String,
			CreatedAt: user.CreatedAt.Time,
		},
	}

	return resp, rawRefresh, nil
}

// maskIIN masks an IIN for display: 990512345678 -> 9905****5678
func maskIIN(iin string) string {
	if len(iin) != 12 {
		return iin
	}
	return iin[:4] + "****" + iin[8:]
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/auth/ && cd ..
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/auth/service.go
git commit -m "Add auth service with login, register, refresh, and logout"
```

---

### Task 8: Create auth handler

**Files:**
- Create: `backend/internal/auth/handler.go`

- [ ] **Step 1: Create handler.go**

Create `backend/internal/auth/handler.go`:

```go
package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/platform/response"
)

// Handler holds auth HTTP handlers.
type Handler struct {
	svc        *Service
	refreshTTL time.Duration
	secureCookie bool
}

// NewHandler creates a new auth handler.
func NewHandler(svc *Service, refreshTTL time.Duration, secureCookie bool) *Handler {
	return &Handler{
		svc:          svc,
		refreshTTL:   refreshTTL,
		secureCookie: secureCookie,
	}
}

// Routes registers auth routes on the given router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	return r
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	resp, rawRefresh, err := h.svc.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			response.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
		case errors.Is(err, ErrUserNotActive):
			response.Error(w, http.StatusForbidden, "USER_INACTIVE", "user account is not active")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL", "login failed")
		}
		return
	}

	h.setRefreshCookie(w, rawRefresh)
	response.JSON(w, http.StatusOK, resp)
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	resp, rawRefresh, err := h.svc.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUsernameExists):
			response.Error(w, http.StatusConflict, "USERNAME_EXISTS", "username already taken")
		case errors.Is(err, ErrEmailExists):
			response.Error(w, http.StatusConflict, "EMAIL_EXISTS", "email already taken")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL", "registration failed")
		}
		return
	}

	h.setRefreshCookie(w, rawRefresh)
	response.JSON(w, http.StatusCreated, resp)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		response.Error(w, http.StatusUnauthorized, "NO_REFRESH_TOKEN", "refresh token not found")
		return
	}

	resp, rawRefresh, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRefreshToken):
			h.clearRefreshCookie(w)
			response.Error(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "invalid or expired refresh token")
		case errors.Is(err, ErrUserNotActive):
			h.clearRefreshCookie(w)
			response.Error(w, http.StatusForbidden, "USER_INACTIVE", "user account is not active")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL", "token refresh failed")
		}
		return
	}

	h.setRefreshCookie(w, rawRefresh)
	response.JSON(w, http.StatusOK, resp)
}

// Logout handles POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		_ = h.svc.Logout(r.Context(), cookie.Value)
	}

	h.clearRefreshCookie(w)
	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.refreshTTL.Seconds()),
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/auth/ && cd ..
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/auth/handler.go
git commit -m "Add auth HTTP handlers for login, register, refresh, logout"
```

---

### Task 9: Create JWT middleware

**Files:**
- Create: `backend/internal/platform/middleware/auth.go`

- [ ] **Step 1: Create auth middleware**

Create `backend/internal/platform/middleware/auth.go`:

```go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"jetistik/internal/auth"
	"jetistik/internal/platform/response"
)

type userContextKey struct{}

// UserClaims holds the authenticated user's JWT claims in the request context.
type UserClaims struct {
	UserID   int64
	Username string
	Role     string
}

// JWTAuth creates middleware that extracts and validates JWT access tokens.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				response.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization header is required")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "authorization header must be Bearer <token>")
				return
			}

			claims, err := auth.ParseAccessToken(parts[1], secret)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired access token")
				return
			}

			uc := UserClaims{
				UserID:   claims.UserID,
				Username: claims.Username,
				Role:     claims.Role,
			}

			ctx := context.WithValue(r.Context(), userContextKey{}, uc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUser extracts the authenticated user claims from the request context.
func GetUser(ctx context.Context) (UserClaims, bool) {
	uc, ok := ctx.Value(userContextKey{}).(UserClaims)
	return uc, ok
}

// MustGetUser extracts user claims or panics. Use only after JWTAuth middleware.
func MustGetUser(ctx context.Context) UserClaims {
	uc, ok := GetUser(ctx)
	if !ok {
		panic("middleware: MustGetUser called without JWTAuth middleware")
	}
	return uc
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/platform/middleware/ && cd ..
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/platform/middleware/auth.go
git commit -m "Add JWT authentication middleware"
```

---

### Task 10: Create role-based middleware

**Files:**
- Create: `backend/internal/platform/middleware/role.go`

- [ ] **Step 1: Create role middleware**

Create `backend/internal/platform/middleware/role.go`:

```go
package middleware

import (
	"net/http"

	"jetistik/internal/platform/response"
)

// RequireRole creates middleware that checks whether the authenticated user
// has one of the allowed roles. Must be used after JWTAuth middleware.
func RequireRole(allowed ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		roleSet[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uc, ok := GetUser(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authentication required")
				return
			}

			if !roleSet[uc.Role] {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/platform/middleware/ && cd ..
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/platform/middleware/role.go
git commit -m "Add role-based access control middleware"
```

---

### Task 11: Create rate limiting middleware

**Files:**
- Create: `backend/internal/platform/middleware/ratelimit.go`

- [ ] **Step 1: Create rate limiter**

Create `backend/internal/platform/middleware/ratelimit.go`:

```go
package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"jetistik/internal/platform/response"
)

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// RateLimit creates middleware that limits requests per IP using a token bucket.
// rate is the number of requests allowed per interval.
func RateLimit(rate int, interval time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	visitors := make(map[string]*visitor)
	maxTokens := float64(rate)
	refillRate := maxTokens / interval.Seconds()

	// Background cleanup of stale visitors
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			// Check X-Forwarded-For for proxied requests
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			mu.Lock()
			v, exists := visitors[ip]
			now := time.Now()

			if !exists {
				v = &visitor{tokens: maxTokens, lastSeen: now}
				visitors[ip] = v
			}

			// Refill tokens based on elapsed time
			elapsed := now.Sub(v.lastSeen).Seconds()
			v.tokens += elapsed * refillRate
			if v.tokens > maxTokens {
				v.tokens = maxTokens
			}
			v.lastSeen = now

			if v.tokens < 1 {
				mu.Unlock()
				w.Header().Set("Retry-After", "60")
				response.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, please try again later")
				return
			}

			v.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/platform/middleware/ && cd ..
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/platform/middleware/ratelimit.go
git commit -m "Add in-memory rate limiting middleware with token bucket"
```

---

### Task 12: Create user DTOs and repository

**Files:**
- Create: `backend/internal/user/dto.go`
- Create: `backend/internal/user/repository.go`

- [ ] **Step 1: Create user dto.go**

Create `backend/internal/user/dto.go`:

```go
package user

import "time"

// --- Requests ---

type UpdateProfileRequest struct {
	Email    *string `json:"email"`
	IIN      *string `json:"iin"`
	Language *string `json:"language"`
}

type AddStudentRequest struct {
	StudentIIN string `json:"student_iin"`
}

// --- Responses ---

type ProfileResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	IIN       string    `json:"iin,omitempty"`
	Role      string    `json:"role"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
}

type TeacherStudentResponse struct {
	ID         int64     `json:"id"`
	TeacherID  int64     `json:"teacher_id"`
	StudentIIN string    `json:"student_iin"`
	CreatedAt  time.Time `json:"created_at"`
}

// --- Validation ---

func (r UpdateProfileRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.IIN != nil && *r.IIN != "" && len(*r.IIN) != 12 {
		errs["iin"] = "IIN must be exactly 12 digits"
	}
	if r.Language != nil && *r.Language != "" {
		if *r.Language != "kz" && *r.Language != "ru" && *r.Language != "en" {
			errs["language"] = "language must be kz, ru, or en"
		}
	}
	return errs
}

func (r AddStudentRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.StudentIIN == "" {
		errs["student_iin"] = "student IIN is required"
	}
	if r.StudentIIN != "" && len(r.StudentIIN) != 12 {
		errs["student_iin"] = "IIN must be exactly 12 digits"
	}
	return errs
}
```

- [ ] **Step 2: Create user repository.go**

Create `backend/internal/user/repository.go`:

```go
package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Repository defines data access for user operations.
type Repository interface {
	GetUserByID(ctx context.Context, id int64) (sqlcdb.User, error)
	UpdateUserProfile(ctx context.Context, params sqlcdb.UpdateUserProfileParams) (sqlcdb.User, error)
	ListTeacherStudents(ctx context.Context, teacherID int64) ([]sqlcdb.TeacherStudent, error)
	AddTeacherStudent(ctx context.Context, teacherID int64, studentIIN string) (sqlcdb.TeacherStudent, error)
	RemoveTeacherStudent(ctx context.Context, teacherID int64, studentIIN string) error
}

type pgRepository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new user repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{q: sqlcdb.New(pool)}
}

func (r *pgRepository) GetUserByID(ctx context.Context, id int64) (sqlcdb.User, error) {
	user, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return sqlcdb.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (r *pgRepository) UpdateUserProfile(ctx context.Context, params sqlcdb.UpdateUserProfileParams) (sqlcdb.User, error) {
	user, err := r.q.UpdateUserProfile(ctx, params)
	if err != nil {
		return sqlcdb.User{}, fmt.Errorf("update profile: %w", err)
	}
	return user, nil
}

func (r *pgRepository) ListTeacherStudents(ctx context.Context, teacherID int64) ([]sqlcdb.TeacherStudent, error) {
	students, err := r.q.ListTeacherStudents(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list teacher students: %w", err)
	}
	return students, nil
}

func (r *pgRepository) AddTeacherStudent(ctx context.Context, teacherID int64, studentIIN string) (sqlcdb.TeacherStudent, error) {
	ts, err := r.q.AddTeacherStudent(ctx, sqlcdb.AddTeacherStudentParams{
		TeacherID:  teacherID,
		StudentIin: studentIIN,
	})
	if err != nil {
		return sqlcdb.TeacherStudent{}, fmt.Errorf("add teacher student: %w", err)
	}
	return ts, nil
}

func (r *pgRepository) RemoveTeacherStudent(ctx context.Context, teacherID int64, studentIIN string) error {
	err := r.q.RemoveTeacherStudent(ctx, sqlcdb.RemoveTeacherStudentParams{
		TeacherID:  teacherID,
		StudentIin: studentIIN,
	})
	if err != nil {
		return fmt.Errorf("remove teacher student: %w", err)
	}
	return nil
}
```

- [ ] **Step 3: Add teacher_students queries**

Add to `backend/queries/users.sql` (append at the end):

```sql
-- name: ListTeacherStudents :many
SELECT id, teacher_id, student_iin, created_at
FROM teacher_students
WHERE teacher_id = $1
ORDER BY created_at DESC;

-- name: AddTeacherStudent :one
INSERT INTO teacher_students (teacher_id, student_iin)
VALUES ($1, $2)
RETURNING id, teacher_id, student_iin, created_at;

-- name: RemoveTeacherStudent :exec
DELETE FROM teacher_students
WHERE teacher_id = $1 AND student_iin = $2;
```

- [ ] **Step 4: Regenerate sqlc**

```bash
cd backend && sqlc generate && cd ..
```

- [ ] **Step 5: Verify compilation**

```bash
cd backend && go build ./internal/user/ && cd ..
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/user/dto.go backend/internal/user/repository.go backend/queries/users.sql backend/internal/sqlcdb/
git commit -m "Add user module DTOs, repository, and teacher-student queries"
```

---

### Task 13: Create user service and handler

**Files:**
- Create: `backend/internal/user/service.go`
- Create: `backend/internal/user/handler.go`

- [ ] **Step 1: Create user service.go**

Create `backend/internal/user/service.go`:

```go
package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"jetistik/internal/sqlcdb"
)

// Service handles user business logic.
type Service struct {
	repo Repository
}

// NewService creates a new user service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetProfile returns the user's profile.
func (s *Service) GetProfile(ctx context.Context, userID int64) (*ProfileResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return toProfileResponse(user), nil
}

// UpdateProfile updates the user's profile fields.
func (s *Service) UpdateProfile(ctx context.Context, userID int64, req UpdateProfileRequest) (*ProfileResponse, error) {
	params := sqlcdb.UpdateUserProfileParams{
		ID: userID,
	}
	if req.Email != nil {
		params.Email = pgtype.Text{String: *req.Email, Valid: true}
	}
	if req.IIN != nil {
		params.Iin = pgtype.Text{String: *req.IIN, Valid: true}
	}
	if req.Language != nil {
		params.Language = pgtype.Text{String: *req.Language, Valid: true}
	}

	user, err := s.repo.UpdateUserProfile(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return toProfileResponse(user), nil
}

// ListStudents returns the teacher's linked students.
func (s *Service) ListStudents(ctx context.Context, teacherID int64) ([]TeacherStudentResponse, error) {
	students, err := s.repo.ListTeacherStudents(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}

	result := make([]TeacherStudentResponse, len(students))
	for i, ts := range students {
		result[i] = TeacherStudentResponse{
			ID:         ts.ID,
			TeacherID:  ts.TeacherID,
			StudentIIN: ts.StudentIin,
			CreatedAt:  ts.CreatedAt.Time,
		}
	}
	return result, nil
}

// AddStudent links a student IIN to the teacher.
func (s *Service) AddStudent(ctx context.Context, teacherID int64, studentIIN string) (*TeacherStudentResponse, error) {
	ts, err := s.repo.AddTeacherStudent(ctx, teacherID, studentIIN)
	if err != nil {
		return nil, fmt.Errorf("add student: %w", err)
	}
	return &TeacherStudentResponse{
		ID:         ts.ID,
		TeacherID:  ts.TeacherID,
		StudentIIN: ts.StudentIin,
		CreatedAt:  ts.CreatedAt.Time,
	}, nil
}

// RemoveStudent unlinks a student IIN from the teacher.
func (s *Service) RemoveStudent(ctx context.Context, teacherID int64, studentIIN string) error {
	return s.repo.RemoveTeacherStudent(ctx, teacherID, studentIIN)
}

func toProfileResponse(u sqlcdb.User) *ProfileResponse {
	iin := u.Iin.String
	if len(iin) == 12 {
		iin = iin[:4] + "****" + iin[8:]
	}
	return &ProfileResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email.String,
		IIN:       iin,
		Role:      u.Role,
		Language:  u.Language.String,
		CreatedAt: u.CreatedAt.Time,
	}
}
```

- [ ] **Step 2: Create user handler.go**

Create `backend/internal/user/handler.go`:

```go
package user

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/platform/middleware"
	"jetistik/internal/platform/response"
)

// Handler holds user HTTP handlers.
type Handler struct {
	svc *Service
}

// NewHandler creates a new user handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ProfileRoutes registers profile-related routes.
func (h *Handler) ProfileRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetProfile)
	r.Patch("/", h.UpdateProfile)
	return r
}

// TeacherStudentRoutes registers teacher-student routes.
func (h *Handler) TeacherStudentRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireRole("teacher"))
	r.Get("/", h.ListStudents)
	r.Post("/", h.AddStudent)
	r.Delete("/{iin}", h.RemoveStudent)
	return r
}

// GetProfile handles GET /api/v1/profile
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())

	profile, err := h.svc.GetProfile(r.Context(), uc.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to get profile")
		return
	}

	response.JSON(w, http.StatusOK, profile)
}

// UpdateProfile handles PATCH /api/v1/profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	profile, err := h.svc.UpdateProfile(r.Context(), uc.UserID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to update profile")
		return
	}

	response.JSON(w, http.StatusOK, profile)
}

// ListStudents handles GET /api/v1/teacher/students
func (h *Handler) ListStudents(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())

	students, err := h.svc.ListStudents(r.Context(), uc.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to list students")
		return
	}

	response.JSON(w, http.StatusOK, students)
}

// AddStudent handles POST /api/v1/teacher/students
func (h *Handler) AddStudent(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())

	var req AddStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	ts, err := h.svc.AddStudent(r.Context(), uc.UserID, req.StudentIIN)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to add student")
		return
	}

	response.JSON(w, http.StatusCreated, ts)
}

// RemoveStudent handles DELETE /api/v1/teacher/students/{iin}
func (h *Handler) RemoveStudent(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	iin := chi.URLParam(r, "iin")

	if err := h.svc.RemoveStudent(r.Context(), uc.UserID, iin); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to remove student")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 3: Verify compilation**

```bash
cd backend && go build ./internal/user/ && cd ..
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/user/service.go backend/internal/user/handler.go
git commit -m "Add user service and handler for profile CRUD and teacher-students"
```

---

### Task 14: Wire auth and user modules into main.go

**Files:**
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Update main.go with auth and user routes**

Replace `backend/cmd/server/main.go` with:

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

	"jetistik/internal/auth"
	"jetistik/internal/platform/config"
	"jetistik/internal/platform/db"
	"jetistik/internal/platform/middleware"
	"jetistik/internal/platform/response"
	"jetistik/internal/user"
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

	// Wire modules
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	authHandler := auth.NewHandler(authSvc, cfg.JWTRefreshTTL, !cfg.IsDev())

	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS(cfg.PublicBaseURL))

	r.Route("/api/v1", func(r chi.Router) {
		// Health
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			err := pool.Ping(r.Context())
			if err != nil {
				response.Error(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database is not reachable")
				return
			}
			response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		// Public auth routes (rate-limited)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RateLimit(10, time.Minute))
			r.Mount("/auth", authHandler.Routes())
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWTSecret))

			r.Mount("/profile", userHandler.ProfileRoutes())
			r.Mount("/teacher/students", userHandler.TeacherStudentRoutes())
		})
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

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./cmd/server && cd ..
```

- [ ] **Step 3: Test the full server manually**

Start infrastructure and run:

```bash
cd backend
JWT_SECRET=dev-secret-change-me DATABASE_URL="postgres://jetistik:dev-password@localhost:5432/jetistik?sslmode=disable" go run ./cmd/server &
sleep 2

# Test registration
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123","role":"student"}' | python3 -m json.tool

# Test login
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}' | python3 -m json.tool

kill %1
cd ..
```

Expected: register returns 201 with access_token and user data. Login returns 200 with access_token and user data. Both responses set a `refresh_token` cookie.

- [ ] **Step 4: Commit**

```bash
git add backend/cmd/server/main.go
git commit -m "Wire auth and user modules into server with routes and middleware"
```

---

### Task 15: Add auth translations to i18n

**Files:**
- Modify: `frontend/src/lib/i18n/kz.ts`
- Modify: `frontend/src/lib/i18n/ru.ts`
- Modify: `frontend/src/lib/i18n/en.ts`

- [ ] **Step 1: Update kz.ts**

Add the following keys to `frontend/src/lib/i18n/kz.ts` (add to the existing object):

```typescript
// Auth
"auth.login": "Кіру",
"auth.register": "Тіркелу",
"auth.logout": "Шығу",
"auth.username": "Пайдаланушы аты",
"auth.password": "Құпиясөз",
"auth.email": "Электрондық пошта",
"auth.iin": "ЖСН",
"auth.role": "Рөлі",
"auth.role.student": "Студент",
"auth.role.teacher": "Мұғалім",
"auth.org_name": "Ұйым атауы",
"auth.register_as": "Тіркелу ретінде",
"auth.register_org": "Ұйым тіркеу",
"auth.have_account": "Аккаунтыңыз бар ма?",
"auth.no_account": "Аккаунтыңыз жоқ па?",
"auth.login_action": "Кіру",
"auth.register_action": "Тіркелу",
"auth.logging_out": "Сіз жүйеден шығып жатырсыз...",
"auth.logged_out": "Сіз жүйеден шықтыңыз",
"auth.error.invalid_credentials": "Қате пайдаланушы аты немесе құпиясөз",
"auth.error.username_exists": "Бұл пайдаланушы аты бос емес",
"auth.error.email_exists": "Бұл электрондық пошта тіркелген",
```

- [ ] **Step 2: Update ru.ts**

Add the following keys to `frontend/src/lib/i18n/ru.ts`:

```typescript
// Auth
"auth.login": "Вход",
"auth.register": "Регистрация",
"auth.logout": "Выход",
"auth.username": "Имя пользователя",
"auth.password": "Пароль",
"auth.email": "Электронная почта",
"auth.iin": "ИИН",
"auth.role": "Роль",
"auth.role.student": "Студент",
"auth.role.teacher": "Преподаватель",
"auth.org_name": "Название организации",
"auth.register_as": "Зарегистрироваться как",
"auth.register_org": "Регистрация организации",
"auth.have_account": "Уже есть аккаунт?",
"auth.no_account": "Нет аккаунта?",
"auth.login_action": "Войти",
"auth.register_action": "Зарегистрироваться",
"auth.logging_out": "Выполняется выход...",
"auth.logged_out": "Вы вышли из системы",
"auth.error.invalid_credentials": "Неверное имя пользователя или пароль",
"auth.error.username_exists": "Имя пользователя уже занято",
"auth.error.email_exists": "Электронная почта уже зарегистрирована",
```

- [ ] **Step 3: Update en.ts**

Add the following keys to `frontend/src/lib/i18n/en.ts`:

```typescript
// Auth
"auth.login": "Log in",
"auth.register": "Register",
"auth.logout": "Log out",
"auth.username": "Username",
"auth.password": "Password",
"auth.email": "Email",
"auth.iin": "IIN",
"auth.role": "Role",
"auth.role.student": "Student",
"auth.role.teacher": "Teacher",
"auth.org_name": "Organization name",
"auth.register_as": "Register as",
"auth.register_org": "Register organization",
"auth.have_account": "Already have an account?",
"auth.no_account": "Don't have an account?",
"auth.login_action": "Log in",
"auth.register_action": "Register",
"auth.logging_out": "Logging out...",
"auth.logged_out": "You have been logged out",
"auth.error.invalid_credentials": "Invalid username or password",
"auth.error.username_exists": "Username already taken",
"auth.error.email_exists": "Email already registered",
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/i18n/
git commit -m "Add auth translation keys for KZ, RU, EN"
```

---

### Task 16: Create auth store

**Files:**
- Create: `frontend/src/lib/stores/auth.ts`

- [ ] **Step 1: Create auth.ts store**

Create `frontend/src/lib/stores/auth.ts`:

```typescript
import { writable, derived } from "svelte/store";
import { browser } from "$app/environment";
import { api, setAccessToken, type ApiResponse } from "$lib/api/client";

export interface AuthUser {
  id: number;
  username: string;
  email?: string;
  iin?: string;
  role: string;
  language: string;
  created_at: string;
}

interface AuthState {
  user: AuthUser | null;
  loading: boolean;
}

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>({
    user: null,
    loading: true,
  });

  return {
    subscribe,

    async login(username: string, password: string): Promise<void> {
      const res = await api.post<{
        access_token: string;
        user: AuthUser;
      }>("/api/v1/auth/login", { username, password });

      setAccessToken(res.data.access_token);
      set({ user: res.data.user, loading: false });
    },

    async register(data: {
      username: string;
      password: string;
      email?: string;
      iin?: string;
      role: string;
      language?: string;
    }): Promise<void> {
      const res = await api.post<{
        access_token: string;
        user: AuthUser;
      }>("/api/v1/auth/register", data);

      setAccessToken(res.data.access_token);
      set({ user: res.data.user, loading: false });
    },

    async refresh(): Promise<boolean> {
      try {
        const res = await api.post<{
          access_token: string;
          user: AuthUser;
        }>("/api/v1/auth/refresh");

        setAccessToken(res.data.access_token);
        set({ user: res.data.user, loading: false });
        return true;
      } catch {
        setAccessToken(null);
        set({ user: null, loading: false });
        return false;
      }
    },

    async logout(): Promise<void> {
      try {
        await api.post("/api/v1/auth/logout");
      } catch {
        // Ignore errors on logout
      }
      setAccessToken(null);
      set({ user: null, loading: false });
    },

    reset() {
      setAccessToken(null);
      set({ user: null, loading: false });
    },
  };
}

export const auth = createAuthStore();

export const isAuthenticated = derived(auth, ($auth) => $auth.user !== null);
export const currentUser = derived(auth, ($auth) => $auth.user);
export const userRole = derived(auth, ($auth) => $auth.user?.role ?? null);
export const isLoading = derived(auth, ($auth) => $auth.loading);
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/lib/stores/auth.ts
git commit -m "Add auth store with login, register, refresh, logout"
```

---

### Task 17: Update API client with token refresh interceptor

**Files:**
- Modify: `frontend/src/lib/api/client.ts`

- [ ] **Step 1: Update client.ts to handle 401 with refresh**

Replace `frontend/src/lib/api/client.ts` with:

```typescript
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
let refreshPromise: Promise<boolean> | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken(): string | null {
  return accessToken;
}

async function refreshAccessToken(): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/auth/refresh`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
    });

    if (!res.ok) {
      accessToken = null;
      return false;
    }

    const body = (await res.json()) as ApiResponse<{ access_token: string }>;
    accessToken = body.data.access_token;
    return true;
  } catch {
    accessToken = null;
    return false;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  retry = true
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

  if (res.status === 401 && retry && !path.includes("/auth/")) {
    // Deduplicate concurrent refresh calls
    if (!refreshPromise) {
      refreshPromise = refreshAccessToken().finally(() => {
        refreshPromise = null;
      });
    }

    const refreshed = await refreshPromise;
    if (refreshed) {
      return request<T>(path, options, false);
    }
  }

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

- [ ] **Step 2: Commit**

```bash
git add frontend/src/lib/api/client.ts
git commit -m "Add automatic token refresh interceptor to API client"
```

---

### Task 18: Create auth layout and login page

**Files:**
- Create: `frontend/src/routes/(auth)/+layout.svelte`
- Create: `frontend/src/routes/(auth)/login/+page.svelte`

- [ ] **Step 1: Create auth layout**

Create `frontend/src/routes/(auth)/+layout.svelte`:

```svelte
<script>
  let { children } = $props();
</script>

<div class="min-h-screen bg-surface flex items-center justify-center p-4">
  <div class="w-full max-w-md">
    <div class="text-center mb-8">
      <h1 class="font-display text-2xl font-bold text-on-surface">Jetistik</h1>
      <p class="text-on-surface-variant text-sm mt-1">
        Digital certificate platform
      </p>
    </div>
    <div class="bg-surface-lowest rounded-lg p-8 shadow-[0_4px_40px_rgba(0,74,198,0.04)]">
      {@render children()}
    </div>
  </div>
</div>
```

- [ ] **Step 2: Create login page**

Create `frontend/src/routes/(auth)/login/+page.svelte`:

```svelte
<script lang="ts">
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth";
  import { t } from "$lib/i18n";
  import { ApiError } from "$lib/api/client";

  let username = $state("");
  let password = $state("");
  let error = $state("");
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = "";
    loading = true;

    try {
      await auth.login(username, password);
      goto("/");
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === "INVALID_CREDENTIALS") {
          error = $t("auth.error.invalid_credentials");
        } else {
          error = err.message;
        }
      } else {
        error = "An unexpected error occurred";
      }
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>{$t("auth.login")} — Jetistik</title>
</svelte:head>

<form onsubmit={handleSubmit}>
  <h2 class="font-display text-xl font-semibold text-on-surface mb-6">
    {$t("auth.login")}
  </h2>

  {#if error}
    <div class="bg-error-container text-on-error-container rounded-md p-3 mb-4 text-sm">
      {error}
    </div>
  {/if}

  <div class="space-y-4">
    <div>
      <label for="username" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.username")}
      </label>
      <input
        id="username"
        type="text"
        bind:value={username}
        required
        autocomplete="username"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
        placeholder={$t("auth.username")}
      />
    </div>

    <div>
      <label for="password" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.password")}
      </label>
      <input
        id="password"
        type="password"
        bind:value={password}
        required
        autocomplete="current-password"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
        placeholder={$t("auth.password")}
      />
    </div>
  </div>

  <button
    type="submit"
    disabled={loading}
    class="w-full mt-6 py-2.5 px-4 rounded-md text-on-primary font-medium text-sm
           bg-gradient-to-br from-primary to-primary-container
           hover:opacity-90 disabled:opacity-50
           transition-opacity cursor-pointer disabled:cursor-not-allowed"
  >
    {loading ? "..." : $t("auth.login_action")}
  </button>

  <p class="text-center text-sm text-on-surface-variant mt-4">
    {$t("auth.no_account")}
    <a href="/register" class="text-primary font-medium hover:underline">
      {$t("auth.register")}
    </a>
  </p>
</form>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/routes/\(auth\)/
git commit -m "Add auth layout and login page"
```

---

### Task 19: Create register page

**Files:**
- Create: `frontend/src/routes/(auth)/register/+page.svelte`

- [ ] **Step 1: Create register page**

Create `frontend/src/routes/(auth)/register/+page.svelte`:

```svelte
<script lang="ts">
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth";
  import { t, language } from "$lib/i18n";
  import { ApiError } from "$lib/api/client";

  let username = $state("");
  let email = $state("");
  let password = $state("");
  let iin = $state("");
  let role = $state<"student" | "teacher">("student");
  let error = $state("");
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = "";
    loading = true;

    try {
      let currentLang: string = "kz";
      language.subscribe((v) => (currentLang = v))();

      await auth.register({
        username,
        password,
        email: email || undefined,
        iin: iin || undefined,
        role,
        language: currentLang,
      });
      goto("/");
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === "USERNAME_EXISTS") {
          error = $t("auth.error.username_exists");
        } else if (err.code === "EMAIL_EXISTS") {
          error = $t("auth.error.email_exists");
        } else {
          error = err.message;
        }
      } else {
        error = "An unexpected error occurred";
      }
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>{$t("auth.register")} — Jetistik</title>
</svelte:head>

<form onsubmit={handleSubmit}>
  <h2 class="font-display text-xl font-semibold text-on-surface mb-6">
    {$t("auth.register")}
  </h2>

  {#if error}
    <div class="bg-error-container text-on-error-container rounded-md p-3 mb-4 text-sm">
      {error}
    </div>
  {/if}

  <div class="space-y-4">
    <div>
      <label for="username" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.username")} *
      </label>
      <input
        id="username"
        type="text"
        bind:value={username}
        required
        minlength="3"
        autocomplete="username"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
      />
    </div>

    <div>
      <label for="email" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.email")}
      </label>
      <input
        id="email"
        type="email"
        bind:value={email}
        autocomplete="email"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
      />
    </div>

    <div>
      <label for="password" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.password")} *
      </label>
      <input
        id="password"
        type="password"
        bind:value={password}
        required
        minlength="8"
        autocomplete="new-password"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
      />
    </div>

    <div>
      <label for="iin" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.iin")}
      </label>
      <input
        id="iin"
        type="text"
        bind:value={iin}
        maxlength="12"
        pattern="[0-9]{12}"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
        placeholder="123456789012"
      />
    </div>

    <div>
      <label class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.register_as")} *
      </label>
      <div class="flex gap-3">
        <button
          type="button"
          onclick={() => (role = "student")}
          class="flex-1 py-2.5 rounded-md text-sm font-medium transition-colors
                 {role === 'student'
                   ? 'bg-gradient-to-br from-primary to-primary-container text-on-primary'
                   : 'bg-surface-low text-on-surface-variant hover:text-on-surface'}"
        >
          {$t("auth.role.student")}
        </button>
        <button
          type="button"
          onclick={() => (role = "teacher")}
          class="flex-1 py-2.5 rounded-md text-sm font-medium transition-colors
                 {role === 'teacher'
                   ? 'bg-gradient-to-br from-primary to-primary-container text-on-primary'
                   : 'bg-surface-low text-on-surface-variant hover:text-on-surface'}"
        >
          {$t("auth.role.teacher")}
        </button>
      </div>
    </div>
  </div>

  <button
    type="submit"
    disabled={loading}
    class="w-full mt-6 py-2.5 px-4 rounded-md text-on-primary font-medium text-sm
           bg-gradient-to-br from-primary to-primary-container
           hover:opacity-90 disabled:opacity-50
           transition-opacity cursor-pointer disabled:cursor-not-allowed"
  >
    {loading ? "..." : $t("auth.register_action")}
  </button>

  <p class="text-center text-sm text-on-surface-variant mt-4">
    {$t("auth.have_account")}
    <a href="/login" class="text-primary font-medium hover:underline">
      {$t("auth.login")}
    </a>
  </p>
</form>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/routes/\(auth\)/register/
git commit -m "Add registration page with role selection"
```

---

### Task 20: Create logout page

**Files:**
- Create: `frontend/src/routes/(auth)/logout/+page.svelte`

- [ ] **Step 1: Create logout page**

Create `frontend/src/routes/(auth)/logout/+page.svelte`:

```svelte
<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth";
  import { t } from "$lib/i18n";

  let done = $state(false);

  onMount(async () => {
    await auth.logout();
    done = true;
    setTimeout(() => goto("/login"), 1500);
  });
</script>

<svelte:head>
  <title>{$t("auth.logout")} — Jetistik</title>
</svelte:head>

<div class="text-center py-8">
  {#if done}
    <p class="text-on-surface font-medium">{$t("auth.logged_out")}</p>
    <p class="text-on-surface-variant text-sm mt-2">
      Redirecting to login...
    </p>
  {:else}
    <p class="text-on-surface-variant">{$t("auth.logging_out")}</p>
  {/if}
</div>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/routes/\(auth\)/logout/
git commit -m "Add logout page with automatic redirect"
```

---

### Task 21: Create protected app layout with auth guard

**Files:**
- Create: `frontend/src/routes/(app)/+layout.svelte`

- [ ] **Step 1: Create app layout with auth guard**

Create `frontend/src/routes/(app)/+layout.svelte`:

```svelte
<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { auth, isAuthenticated, isLoading } from "$lib/stores/auth";

  let { children } = $props();
  let initialized = $state(false);

  onMount(async () => {
    // Try to refresh token on initial load
    const ok = await auth.refresh();
    if (!ok) {
      goto("/login");
      return;
    }
    initialized = true;
  });

  // Redirect if auth state changes to unauthenticated after init
  $effect(() => {
    if (initialized && !$isLoading && !$isAuthenticated) {
      goto("/login");
    }
  });
</script>

{#if !initialized}
  <div class="min-h-screen bg-surface flex items-center justify-center">
    <div class="text-on-surface-variant">Loading...</div>
  </div>
{:else}
  {@render children()}
{/if}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/routes/\(app\)/
git commit -m "Add protected app layout with auth guard and token refresh"
```

---

### Task 22: End-to-end verification

- [ ] **Step 1: Verify all backend code compiles**

```bash
cd backend && go build ./... && cd ..
```

- [ ] **Step 2: Run all backend tests**

```bash
cd backend && go test ./... -v && cd ..
```

Expected: password tests pass, no other test files yet.

- [ ] **Step 3: Verify frontend builds**

```bash
cd frontend && npm run check && cd ..
```

- [ ] **Step 4: Manual integration test**

Start Docker services and the backend:

```bash
docker compose up -d
cd backend
goose -dir migrations postgres "postgres://jetistik:dev-password@localhost:5432/jetistik?sslmode=disable" up
JWT_SECRET=dev-secret DATABASE_URL="postgres://jetistik:dev-password@localhost:5432/jetistik?sslmode=disable" go run ./cmd/server &
sleep 2
```

Test the full auth flow:

```bash
# Register
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"e2euser","password":"e2epass123","role":"student","iin":"990512345678"}' | python3 -m json.tool

# Login and capture token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"e2euser","password":"e2epass123"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])")

# Get profile
curl -s http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool

# Update profile
curl -s -X PATCH http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"language":"ru"}' | python3 -m json.tool

# Test unauthorized access
curl -s http://localhost:8080/api/v1/profile | python3 -m json.tool

# Test rate limiting (run 12 times rapidly)
for i in $(seq 1 12); do
  curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"x","password":"x"}'
  echo ""
done

kill %1
```

Expected:
- Register returns 201 with user data and access_token
- Login returns 200 with access_token
- Profile returns 200 with user data (IIN masked as `9905****5678`)
- Profile update returns 200 with language changed to `ru`
- Unauthorized profile access returns 401
- Rate limiting: first 10 requests return 401, last 2 return 429

- [ ] **Step 5: Commit (if any final fixes needed)**

Only commit if fixes were needed. The plan should work without fixes if followed exactly.
