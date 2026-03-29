package audit

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Repository defines data access for audit operations.
type Repository interface {
	CreateAuditLog(ctx context.Context, params sqlcdb.CreateAuditLogParams) (sqlcdb.AuditLog, error)
	ListAuditLogs(ctx context.Context, limit, offset int32) ([]sqlcdb.ListAuditLogsRow, error)
	CountAuditLogs(ctx context.Context) (int64, error)
	ListAuditLogsByAction(ctx context.Context, action string, limit, offset int32) ([]sqlcdb.ListAuditLogsByActionRow, error)
}

type pgRepository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new audit repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{q: sqlcdb.New(pool)}
}

func (r *pgRepository) CreateAuditLog(ctx context.Context, params sqlcdb.CreateAuditLogParams) (sqlcdb.AuditLog, error) {
	log, err := r.q.CreateAuditLog(ctx, params)
	if err != nil {
		return sqlcdb.AuditLog{}, fmt.Errorf("create audit log: %w", err)
	}
	return log, nil
}

func (r *pgRepository) ListAuditLogs(ctx context.Context, limit, offset int32) ([]sqlcdb.ListAuditLogsRow, error) {
	logs, err := r.q.ListAuditLogs(ctx, sqlcdb.ListAuditLogsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	return logs, nil
}

func (r *pgRepository) CountAuditLogs(ctx context.Context) (int64, error) {
	count, err := r.q.CountAuditLogs(ctx)
	if err != nil {
		return 0, fmt.Errorf("count audit logs: %w", err)
	}
	return count, nil
}

func (r *pgRepository) ListAuditLogsByAction(ctx context.Context, action string, limit, offset int32) ([]sqlcdb.ListAuditLogsByActionRow, error) {
	logs, err := r.q.ListAuditLogsByAction(ctx, sqlcdb.ListAuditLogsByActionParams{
		Action: action,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list audit logs by action: %w", err)
	}
	return logs, nil
}
