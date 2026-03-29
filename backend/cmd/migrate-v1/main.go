package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	SourceDB      string
	TargetDB      string
	MinioEndpoint string
	MinioKey      string
	MinioSecret   string
	MinioBucket   string
	MinioSSL      bool
	MediaDir      string
	DryRun        bool
}

type Report struct {
	StartedAt    time.Time              `json:"started_at"`
	FinishedAt   time.Time              `json:"finished_at"`
	Tables       map[string]TableReport `json:"tables"`
	MissingFiles []string               `json:"missing_files"`
	Errors       []string               `json:"errors"`
}

type TableReport struct {
	Total    int      `json:"total"`
	Migrated int      `json:"migrated"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.SourceDB, "source-db", "", "v1 PostgreSQL connection string")
	flag.StringVar(&cfg.TargetDB, "target-db", "", "v2 PostgreSQL connection string")
	flag.StringVar(&cfg.MinioEndpoint, "minio-endpoint", "localhost:9000", "MinIO endpoint")
	flag.StringVar(&cfg.MinioKey, "minio-key", "", "MinIO access key")
	flag.StringVar(&cfg.MinioSecret, "minio-secret", "", "MinIO secret key")
	flag.StringVar(&cfg.MinioBucket, "minio-bucket", "jetistik", "MinIO bucket name")
	flag.BoolVar(&cfg.MinioSSL, "minio-ssl", false, "Use SSL for MinIO")
	flag.StringVar(&cfg.MediaDir, "media-dir", "", "Path to v1 media directory")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Print plan without executing")
	flag.Parse()

	if cfg.SourceDB == "" || cfg.TargetDB == "" || cfg.MediaDir == "" {
		slog.Error("required flags: --source-db, --target-db, --media-dir")
		os.Exit(1)
	}
	if cfg.MinioKey == "" || cfg.MinioSecret == "" {
		slog.Error("required flags: --minio-key, --minio-secret")
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg Config) error {
	ctx := context.Background()

	slog.Info("connecting to source database (v1)")
	srcPool, err := pgxpool.New(ctx, cfg.SourceDB)
	if err != nil {
		return fmt.Errorf("connect source db: %w", err)
	}
	defer srcPool.Close()

	slog.Info("connecting to target database (v2)")
	dstPool, err := pgxpool.New(ctx, cfg.TargetDB)
	if err != nil {
		return fmt.Errorf("connect target db: %w", err)
	}
	defer dstPool.Close()

	slog.Info("connecting to MinIO", "endpoint", cfg.MinioEndpoint)
	mc, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioKey, cfg.MinioSecret, ""),
		Secure: cfg.MinioSSL,
	})
	if err != nil {
		return fmt.Errorf("connect minio: %w", err)
	}

	// Ensure bucket exists
	exists, err := mc.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		slog.Info("created MinIO bucket", "bucket", cfg.MinioBucket)
	}

	report := &Report{
		StartedAt: time.Now(),
		Tables:    make(map[string]TableReport),
	}

	m := &migrator{
		src:     srcPool,
		dst:     dstPool,
		mc:      mc,
		bucket:  cfg.MinioBucket,
		media:   cfg.MediaDir,
		report:  report,
		dryRun:  cfg.DryRun,
	}

	steps := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"users", m.migrateUsers},
		{"organizations", m.migrateOrganizations},
		{"organization_members", m.migrateOrganizationMembers},
		{"events", m.migrateEvents},
		{"templates", m.migrateTemplates},
		{"import_batches", m.migrateImportBatches},
		{"participant_rows", m.migrateParticipantRows},
		{"certificates", m.migrateCertificates},
		{"teacher_students", m.migrateTeacherStudents},
		{"audit_logs", m.migrateAuditLogs},
	}

	for _, step := range steps {
		slog.Info("migrating", "table", step.name)
		if err := step.fn(ctx); err != nil {
			return fmt.Errorf("migrate %s: %w", step.name, err)
		}
		tr := report.Tables[step.name]
		slog.Info("completed", "table", step.name, "total", tr.Total, "migrated", tr.Migrated, "skipped", tr.Skipped)
	}

	// Reset sequences to max id + 1
	if !cfg.DryRun {
		if err := m.resetSequences(ctx); err != nil {
			slog.Warn("failed to reset sequences", "error", err)
			report.Errors = append(report.Errors, fmt.Sprintf("reset sequences: %v", err))
		}
	}

	report.FinishedAt = time.Now()

	// Output report
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	fmt.Println(string(reportJSON))

	slog.Info("migration complete",
		"duration", report.FinishedAt.Sub(report.StartedAt).Round(time.Second),
		"missing_files", len(report.MissingFiles),
		"errors", len(report.Errors),
	)

	return nil
}

type migrator struct {
	src    *pgxpool.Pool
	dst    *pgxpool.Pool
	mc     *minio.Client
	bucket string
	media  string
	report *Report
	dryRun bool
}

func (m *migrator) uploadFile(ctx context.Context, localPath, objectPath, contentType string) error {
	fullPath := filepath.Join(m.media, localPath)
	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.report.MissingFiles = append(m.report.MissingFiles, localPath)
			slog.Warn("file not found, skipping upload", "path", localPath)
			return nil
		}
		return fmt.Errorf("open file %s: %w", localPath, err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat file %s: %w", localPath, err)
	}

	_, err = m.mc.PutObject(ctx, m.bucket, objectPath, f, stat.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("upload %s to %s: %w", localPath, objectPath, err)
	}
	return nil
}

// migrateUsers migrates auth_user + core_userprofile + groups -> users
func (m *migrator) migrateUsers(ctx context.Context) error {
	query := `
		SELECT
			u.id, u.username, u.email, u.password, u.is_active, u.is_superuser, u.date_joined,
			COALESCE(p.iin, '') AS iin,
			COALESCE(
				(SELECT g.name FROM auth_user_groups ug
				 JOIN auth_group g ON g.id = ug.group_id
				 WHERE ug.user_id = u.id
				 LIMIT 1),
				''
			) AS group_name
		FROM auth_user u
		LEFT JOIN core_userprofile p ON p.user_id = u.id
		ORDER BY u.id
	`

	rows, err := m.src.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("query source users: %w", err)
	}
	defer rows.Close()

	tr := TableReport{}
	for rows.Next() {
		var (
			id          int64
			username    string
			email       *string
			password    string
			isActive    bool
			isSuperuser bool
			dateJoined  time.Time
			iin         string
			groupName   string
		)
		if err := rows.Scan(&id, &username, &email, &password, &isActive, &isSuperuser, &dateJoined, &iin, &groupName); err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("scan user row: %v", err))
			tr.Total++
			tr.Skipped++
			continue
		}
		tr.Total++

		role := mapGroupToRole(groupName, isSuperuser)

		if m.dryRun {
			slog.Info("would migrate user", "id", id, "username", username, "role", role)
			tr.Migrated++
			continue
		}

		_, err := m.dst.Exec(ctx, `
			INSERT INTO users (id, username, email, password, iin, role, is_active, language, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'ru', $8, $8)
			ON CONFLICT (id) DO NOTHING
		`, id, username, email, password, nilIfEmpty(iin), role, isActive, dateJoined)
		if err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("insert user %d (%s): %v", id, username, err))
			tr.Skipped++
			continue
		}
		tr.Migrated++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate users: %w", err)
	}

	m.report.Tables["users"] = tr
	return nil
}

func mapGroupToRole(groupName string, isSuperuser bool) string {
	if isSuperuser {
		return "admin"
	}
	switch groupName {
	case "staff_org":
		return "staff"
	case "user_teacher":
		return "teacher"
	case "user_student":
		return "student"
	default:
		return "student"
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// migrateOrganizations migrates core_organization -> organizations
func (m *migrator) migrateOrganizations(ctx context.Context) error {
	rows, err := m.src.Query(ctx, `SELECT id, name, logo, created_at FROM core_organization ORDER BY id`)
	if err != nil {
		return fmt.Errorf("query source organizations: %w", err)
	}
	defer rows.Close()

	tr := TableReport{}
	for rows.Next() {
		var (
			id        int64
			name      string
			logo      *string
			createdAt time.Time
		)
		if err := rows.Scan(&id, &name, &logo, &createdAt); err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("scan org row: %v", err))
			tr.Total++
			tr.Skipped++
			continue
		}
		tr.Total++

		var logoPath *string
		if logo != nil && *logo != "" {
			objectPath := fmt.Sprintf("logos/%d/%s", id, filepath.Base(*logo))
			if !m.dryRun {
				if err := m.uploadFile(ctx, *logo, objectPath, guessContentType(*logo)); err != nil {
					slog.Warn("failed to upload org logo", "org_id", id, "error", err)
				}
			}
			logoPath = &objectPath
		}

		if m.dryRun {
			slog.Info("would migrate organization", "id", id, "name", name)
			tr.Migrated++
			continue
		}

		_, err := m.dst.Exec(ctx, `
			INSERT INTO organizations (id, name, logo_path, status, created_at, updated_at)
			VALUES ($1, $2, $3, 'active', $4, $4)
			ON CONFLICT (id) DO NOTHING
		`, id, name, logoPath, createdAt)
		if err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("insert org %d: %v", id, err))
			tr.Skipped++
			continue
		}
		tr.Migrated++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate organizations: %w", err)
	}

	m.report.Tables["organizations"] = tr
	return nil
}

// migrateOrganizationMembers migrates core_organizationuser -> organization_members
func (m *migrator) migrateOrganizationMembers(ctx context.Context) error {
	rows, err := m.src.Query(ctx, `SELECT id, organization_id, user_id, created_at FROM core_organizationuser ORDER BY id`)
	if err != nil {
		return fmt.Errorf("query source org users: %w", err)
	}
	defer rows.Close()

	tr := TableReport{}
	for rows.Next() {
		var (
			id        int64
			orgID     int64
			userID    int64
			createdAt time.Time
		)
		if err := rows.Scan(&id, &orgID, &userID, &createdAt); err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("scan org member row: %v", err))
			tr.Total++
			tr.Skipped++
			continue
		}
		tr.Total++

		if m.dryRun {
			tr.Migrated++
			continue
		}

		_, err := m.dst.Exec(ctx, `
			INSERT INTO organization_members (id, organization_id, user_id, role, created_at)
			VALUES ($1, $2, $3, 'member', $4)
			ON CONFLICT (organization_id, user_id) DO NOTHING
		`, id, orgID, userID, createdAt)
		if err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("insert org member %d: %v", id, err))
			tr.Skipped++
			continue
		}
		tr.Migrated++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate org members: %w", err)
	}

	m.report.Tables["organization_members"] = tr
	return nil
}

// migrateEvents migrates core_event -> events
func (m *migrator) migrateEvents(ctx context.Context) error {
	rows, err := m.src.Query(ctx, `SELECT id, organization_id, created_by_id, title, date, city, description, created_at FROM core_event ORDER BY id`)
	if err != nil {
		return fmt.Errorf("query source events: %w", err)
	}
	defer rows.Close()

	tr := TableReport{}
	for rows.Next() {
		var (
			id          int64
			orgID       *int64
			createdByID *int64
			title       string
			date        *time.Time
			city        string
			description string
			createdAt   time.Time
		)
		if err := rows.Scan(&id, &orgID, &createdByID, &title, &date, &city, &description, &createdAt); err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("scan event row: %v", err))
			tr.Total++
			tr.Skipped++
			continue
		}
		tr.Total++

		if orgID == nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("event %d has no organization, skipping", id))
			tr.Skipped++
			continue
		}

		if m.dryRun {
			tr.Migrated++
			continue
		}

		_, err := m.dst.Exec(ctx, `
			INSERT INTO events (id, organization_id, created_by, title, date, city, description, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'active', $8, $8)
			ON CONFLICT (id) DO NOTHING
		`, id, *orgID, createdByID, title, date, city, description, createdAt)
		if err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("insert event %d: %v", id, err))
			tr.Skipped++
			continue
		}
		tr.Migrated++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate events: %w", err)
	}

	m.report.Tables["events"] = tr
	return nil
}

// migrateTemplates migrates core_template -> templates
func (m *migrator) migrateTemplates(ctx context.Context) error {
	rows, err := m.src.Query(ctx, `SELECT id, event_id, pptx_file, created_at FROM core_template ORDER BY id`)
	if err != nil {
		return fmt.Errorf("query source templates: %w", err)
	}
	defer rows.Close()

	tr := TableReport{}
	for rows.Next() {
		var (
			id        int64
			eventID   int64
			filePath  string
			createdAt time.Time
		)
		if err := rows.Scan(&id, &eventID, &filePath, &createdAt); err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("scan template row: %v", err))
			tr.Total++
			tr.Skipped++
			continue
		}
		tr.Total++

		objectPath := fmt.Sprintf("templates/%d/%s", eventID, filepath.Base(filePath))

		if !m.dryRun {
			if err := m.uploadFile(ctx, filePath, objectPath, "application/vnd.openxmlformats-officedocument.presentationml.presentation"); err != nil {
				slog.Warn("failed to upload template", "id", id, "error", err)
			}
		}

		if m.dryRun {
			tr.Migrated++
			continue
		}

		_, err := m.dst.Exec(ctx, `
			INSERT INTO templates (id, event_id, file_path, created_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO NOTHING
		`, id, eventID, objectPath, createdAt)
		if err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("insert template %d: %v", id, err))
			tr.Skipped++
			continue
		}
		tr.Migrated++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate templates: %w", err)
	}

	m.report.Tables["templates"] = tr
	return nil
}

// migrateImportBatches migrates core_importbatch -> import_batches
func (m *migrator) migrateImportBatches(ctx context.Context) error {
	rows, err := m.src.Query(ctx, `
		SELECT id, event_id, file, status, rows_total, rows_ok, rows_failed,
		       mapping_json, tokens_json, report_json, created_at
		FROM core_importbatch ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("query source import batches: %w", err)
	}
	defer rows.Close()

	tr := TableReport{}
	for rows.Next() {
		var (
			id         int64
			eventID    int64
			filePath   string
			status     string
			rowsTotal  int
			rowsOK     int
			rowsFailed int
			mappingRaw json.RawMessage
			tokensRaw  json.RawMessage
			reportRaw  json.RawMessage
			createdAt  time.Time
		)
		if err := rows.Scan(&id, &eventID, &filePath, &status, &rowsTotal, &rowsOK, &rowsFailed,
			&mappingRaw, &tokensRaw, &reportRaw, &createdAt); err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("scan batch row: %v", err))
			tr.Total++
			tr.Skipped++
			continue
		}
		tr.Total++

		objectPath := fmt.Sprintf("imports/%d/%s", eventID, filepath.Base(filePath))

		if !m.dryRun {
			ct := guessContentType(filePath)
			if err := m.uploadFile(ctx, filePath, objectPath, ct); err != nil {
				slog.Warn("failed to upload batch file", "id", id, "error", err)
			}
		}

		if m.dryRun {
			tr.Migrated++
			continue
		}

		_, err := m.dst.Exec(ctx, `
			INSERT INTO import_batches (id, event_id, file_path, status, rows_total, rows_ok, rows_failed,
			                            mapping, tokens, report, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
			ON CONFLICT (id) DO NOTHING
		`, id, eventID, objectPath, status, rowsTotal, rowsOK, rowsFailed,
			jsonOrDefault(mappingRaw, "{}"), jsonOrDefault(tokensRaw, "[]"), jsonOrDefault(reportRaw, "{}"), createdAt)
		if err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("insert batch %d: %v", id, err))
			tr.Skipped++
			continue
		}
		tr.Migrated++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate import batches: %w", err)
	}

	m.report.Tables["import_batches"] = tr
	return nil
}

// migrateParticipantRows migrates core_participantrow -> participant_rows
func (m *migrator) migrateParticipantRows(ctx context.Context) error {
	rows, err := m.src.Query(ctx, `
		SELECT id, batch_id, iin, name, payload_json, status, error
		FROM core_participantrow ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("query source participant rows: %w", err)
	}
	defer rows.Close()

	tr := TableReport{}
	batch := make([][]interface{}, 0, 500)

	flush := func() error {
		if len(batch) == 0 || m.dryRun {
			return nil
		}
		tx, err := m.dst.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback(ctx)

		for _, args := range batch {
			_, err := tx.Exec(ctx, `
				INSERT INTO participant_rows (id, batch_id, iin, name, payload, status, error, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, now())
				ON CONFLICT (id) DO NOTHING
			`, args...)
			if err != nil {
				tr.Errors = append(tr.Errors, fmt.Sprintf("insert participant row %v: %v", args[0], err))
				tr.Skipped++
				continue
			}
			tr.Migrated++
		}

		return tx.Commit(ctx)
	}

	for rows.Next() {
		var (
			id         int64
			batchID    int64
			iin        string
			name       string
			payloadRaw json.RawMessage
			status     string
			errText    string
		)
		if err := rows.Scan(&id, &batchID, &iin, &name, &payloadRaw, &status, &errText); err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("scan participant row: %v", err))
			tr.Total++
			tr.Skipped++
			continue
		}
		tr.Total++

		if m.dryRun {
			tr.Migrated++
			continue
		}

		batch = append(batch, []interface{}{id, batchID, iin, name, jsonOrDefault(payloadRaw, "{}"), status, errText})
		if len(batch) >= 500 {
			if err := flush(); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate participant rows: %w", err)
	}
	if err := flush(); err != nil {
		return err
	}

	m.report.Tables["participant_rows"] = tr
	return nil
}

// migrateCertificates migrates core_certificate -> certificates (CRITICAL: 1700+ certs)
func (m *migrator) migrateCertificates(ctx context.Context) error {
	rows, err := m.src.Query(ctx, `
		SELECT id, event_id, organization_id, iin, name, code, pdf, status, revoked_reason, payload_json, created_at
		FROM core_certificate ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("query source certificates: %w", err)
	}
	defer rows.Close()

	tr := TableReport{}
	for rows.Next() {
		var (
			id            int64
			eventID       int64
			orgID         *int64
			iin           string
			name          string
			code          string
			pdfPath       string
			status        string
			revokedReason string
			payloadRaw    json.RawMessage
			createdAt     time.Time
		)
		if err := rows.Scan(&id, &eventID, &orgID, &iin, &name, &code, &pdfPath, &status, &revokedReason, &payloadRaw, &createdAt); err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("scan certificate row: %v", err))
			tr.Total++
			tr.Skipped++
			continue
		}
		tr.Total++

		// Upload PDF to MinIO preserving event-based path
		objectPath := fmt.Sprintf("certificates/%d/%s.pdf", eventID, code)

		if !m.dryRun && pdfPath != "" {
			if err := m.uploadFile(ctx, pdfPath, objectPath, "application/pdf"); err != nil {
				slog.Warn("failed to upload certificate PDF", "id", id, "code", code, "error", err)
			}
		}

		if m.dryRun {
			tr.Migrated++
			continue
		}

		_, err := m.dst.Exec(ctx, `
			INSERT INTO certificates (id, event_id, organization_id, iin, name, code, pdf_path, status, revoked_reason, payload, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
			ON CONFLICT (id) DO NOTHING
		`, id, eventID, orgID, iin, name, code, objectPath, status, revokedReason,
			jsonOrDefault(payloadRaw, "{}"), createdAt)
		if err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("insert certificate %d (code=%s): %v", id, code, err))
			tr.Skipped++
			continue
		}
		tr.Migrated++

		if tr.Total%100 == 0 {
			slog.Info("certificates progress", "processed", tr.Total, "migrated", tr.Migrated)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate certificates: %w", err)
	}

	m.report.Tables["certificates"] = tr
	return nil
}

// migrateTeacherStudents migrates core_teacherstudent -> teacher_students
func (m *migrator) migrateTeacherStudents(ctx context.Context) error {
	rows, err := m.src.Query(ctx, `SELECT id, teacher_id, student_iin, created_at FROM core_teacherstudent ORDER BY id`)
	if err != nil {
		return fmt.Errorf("query source teacher-students: %w", err)
	}
	defer rows.Close()

	tr := TableReport{}
	for rows.Next() {
		var (
			id         int64
			teacherID  int64
			studentIIN string
			createdAt  time.Time
		)
		if err := rows.Scan(&id, &teacherID, &studentIIN, &createdAt); err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("scan teacher-student row: %v", err))
			tr.Total++
			tr.Skipped++
			continue
		}
		tr.Total++

		if m.dryRun {
			tr.Migrated++
			continue
		}

		_, err := m.dst.Exec(ctx, `
			INSERT INTO teacher_students (id, teacher_id, student_iin, created_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (teacher_id, student_iin) DO NOTHING
		`, id, teacherID, studentIIN, createdAt)
		if err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("insert teacher-student %d: %v", id, err))
			tr.Skipped++
			continue
		}
		tr.Migrated++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate teacher-students: %w", err)
	}

	m.report.Tables["teacher_students"] = tr
	return nil
}

// migrateAuditLogs migrates core_auditlog -> audit_logs
func (m *migrator) migrateAuditLogs(ctx context.Context) error {
	rows, err := m.src.Query(ctx, `SELECT id, actor_id, action, object_type, object_id, meta, created_at FROM core_auditlog ORDER BY id`)
	if err != nil {
		return fmt.Errorf("query source audit logs: %w", err)
	}
	defer rows.Close()

	tr := TableReport{}
	batch := make([][]interface{}, 0, 500)

	flush := func() error {
		if len(batch) == 0 || m.dryRun {
			return nil
		}
		tx, err := m.dst.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback(ctx)

		for _, args := range batch {
			_, err := tx.Exec(ctx, `
				INSERT INTO audit_logs (id, actor_id, action, object_type, object_id, meta, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (id) DO NOTHING
			`, args...)
			if err != nil {
				tr.Errors = append(tr.Errors, fmt.Sprintf("insert audit log %v: %v", args[0], err))
				tr.Skipped++
				continue
			}
			tr.Migrated++
		}

		return tx.Commit(ctx)
	}

	for rows.Next() {
		var (
			id         int64
			actorID    *int64
			action     string
			objectType string
			objectID   string
			metaRaw    json.RawMessage
			createdAt  time.Time
		)
		if err := rows.Scan(&id, &actorID, &action, &objectType, &objectID, &metaRaw, &createdAt); err != nil {
			tr.Errors = append(tr.Errors, fmt.Sprintf("scan audit log row: %v", err))
			tr.Total++
			tr.Skipped++
			continue
		}
		tr.Total++

		if m.dryRun {
			tr.Migrated++
			continue
		}

		batch = append(batch, []interface{}{id, actorID, action, objectType, objectID, jsonOrDefault(metaRaw, "{}"), createdAt})
		if len(batch) >= 500 {
			if err := flush(); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate audit logs: %w", err)
	}
	if err := flush(); err != nil {
		return err
	}

	m.report.Tables["audit_logs"] = tr
	return nil
}

// resetSequences sets all serial sequences to max(id) + 1
func (m *migrator) resetSequences(ctx context.Context) error {
	tables := []struct {
		table    string
		sequence string
	}{
		{"users", "users_id_seq"},
		{"organizations", "organizations_id_seq"},
		{"organization_members", "organization_members_id_seq"},
		{"events", "events_id_seq"},
		{"templates", "templates_id_seq"},
		{"import_batches", "import_batches_id_seq"},
		{"participant_rows", "participant_rows_id_seq"},
		{"certificates", "certificates_id_seq"},
		{"teacher_students", "teacher_students_id_seq"},
		{"audit_logs", "audit_logs_id_seq"},
	}

	for _, t := range tables {
		var maxID *int64
		err := m.dst.QueryRow(ctx, fmt.Sprintf("SELECT MAX(id) FROM %s", t.table)).Scan(&maxID)
		if err != nil {
			return fmt.Errorf("get max id from %s: %w", t.table, err)
		}
		if maxID == nil {
			continue
		}
		_, err = m.dst.Exec(ctx, fmt.Sprintf("SELECT setval('%s', $1)", t.sequence), *maxID)
		if err != nil {
			return fmt.Errorf("setval %s: %w", t.sequence, err)
		}
		slog.Info("reset sequence", "sequence", t.sequence, "value", *maxID)
	}
	return nil
}

// jsonOrDefault returns raw JSON if non-empty, otherwise the default value.
func jsonOrDefault(raw json.RawMessage, def string) string {
	if len(raw) == 0 || string(raw) == "null" {
		return def
	}
	return string(raw)
}

// guessContentType returns a MIME type based on file extension.
func guessContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".csv":
		return "text/csv"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

// Ensure io is used (for interface compliance in future extensions).
var _ = io.EOF
var _ = pgx.ErrNoRows
