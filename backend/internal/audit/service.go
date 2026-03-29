package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	"jetistik/internal/sqlcdb"
)

// Service handles audit log business logic.
type Service struct {
	repo Repository
}

// NewService creates a new audit service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Log records an audit event. It never returns an error to callers — failures are logged.
func (s *Service) Log(ctx context.Context, actorID int64, action, objectType, objectID string, meta map[string]interface{}) {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		metaJSON = []byte("{}")
	}

	_, err = s.repo.CreateAuditLog(ctx, sqlcdb.CreateAuditLogParams{
		ActorID:    pgtype.Int8{Int64: actorID, Valid: actorID > 0},
		Action:     action,
		ObjectType: pgtype.Text{String: objectType, Valid: objectType != ""},
		ObjectID:   pgtype.Text{String: objectID, Valid: objectID != ""},
		Meta:       metaJSON,
	})
	if err != nil {
		slog.Error("failed to record audit log", "action", action, "error", err)
	}
}

// List returns paginated audit logs.
func (s *Service) List(ctx context.Context, page, perPage int, actionFilter string) ([]AuditLogResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	if actionFilter != "" {
		rows, err := s.repo.ListAuditLogsByAction(ctx, actionFilter, int32(perPage), int32(offset))
		if err != nil {
			return nil, 0, fmt.Errorf("list audit logs: %w", err)
		}
		result := make([]AuditLogResponse, len(rows))
		for i, r := range rows {
			result[i] = toAuditResponseFromAction(r)
		}
		// For filtered results, we use len as an approximate count
		return result, int64(len(rows)), nil
	}

	rows, err := s.repo.ListAuditLogs(ctx, int32(perPage), int32(offset))
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	total, err := s.repo.CountAuditLogs(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}
	result := make([]AuditLogResponse, len(rows))
	for i, r := range rows {
		result[i] = toAuditResponse(r)
	}
	return result, total, nil
}

func toAuditResponse(r sqlcdb.ListAuditLogsRow) AuditLogResponse {
	resp := AuditLogResponse{
		ID:            r.ID,
		Action:        r.Action,
		ObjectType:    r.ObjectType.String,
		ObjectID:      r.ObjectID.String,
		ActorUsername: r.ActorUsername.String,
		CreatedAt:     r.CreatedAt.Time,
	}
	if r.ActorID.Valid {
		id := r.ActorID.Int64
		resp.ActorID = &id
	}
	if len(r.Meta) > 0 {
		var meta map[string]interface{}
		if err := json.Unmarshal(r.Meta, &meta); err == nil {
			resp.Meta = meta
		}
	}
	return resp
}

func toAuditResponseFromAction(r sqlcdb.ListAuditLogsByActionRow) AuditLogResponse {
	resp := AuditLogResponse{
		ID:            r.ID,
		Action:        r.Action,
		ObjectType:    r.ObjectType.String,
		ObjectID:      r.ObjectID.String,
		ActorUsername: r.ActorUsername.String,
		CreatedAt:     r.CreatedAt.Time,
	}
	if r.ActorID.Valid {
		id := r.ActorID.Int64
		resp.ActorID = &id
	}
	if len(r.Meta) > 0 {
		var meta map[string]interface{}
		if err := json.Unmarshal(r.Meta, &meta); err == nil {
			resp.Meta = meta
		}
	}
	return resp
}
