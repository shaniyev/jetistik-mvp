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
