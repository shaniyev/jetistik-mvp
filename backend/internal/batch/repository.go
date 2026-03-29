package batch

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Repository defines data access for batch operations.
type Repository interface {
	CreateImportBatch(ctx context.Context, params sqlcdb.CreateImportBatchParams) (sqlcdb.ImportBatch, error)
	GetImportBatchByID(ctx context.Context, id int64) (sqlcdb.ImportBatch, error)
	ListImportBatchesByEvent(ctx context.Context, eventID int64) ([]sqlcdb.ImportBatch, error)
	UpdateImportBatchMapping(ctx context.Context, id int64, mapping []byte, status string) (sqlcdb.ImportBatch, error)
	UpdateImportBatchStatus(ctx context.Context, id int64, status string, rowsOk, rowsFailed int, report []byte) (sqlcdb.ImportBatch, error)
	DeleteImportBatch(ctx context.Context, id int64) error
	CreateParticipantRow(ctx context.Context, params sqlcdb.CreateParticipantRowParams) (sqlcdb.ParticipantRow, error)
	ListParticipantRowsByBatch(ctx context.Context, batchID int64) ([]sqlcdb.ParticipantRow, error)
}

type pgRepository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new batch repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{q: sqlcdb.New(pool)}
}

func (r *pgRepository) CreateImportBatch(ctx context.Context, params sqlcdb.CreateImportBatchParams) (sqlcdb.ImportBatch, error) {
	batch, err := r.q.CreateImportBatch(ctx, params)
	if err != nil {
		return sqlcdb.ImportBatch{}, fmt.Errorf("create import batch: %w", err)
	}
	return batch, nil
}

func (r *pgRepository) GetImportBatchByID(ctx context.Context, id int64) (sqlcdb.ImportBatch, error) {
	batch, err := r.q.GetImportBatchByID(ctx, id)
	if err != nil {
		return sqlcdb.ImportBatch{}, fmt.Errorf("get import batch: %w", err)
	}
	return batch, nil
}

func (r *pgRepository) ListImportBatchesByEvent(ctx context.Context, eventID int64) ([]sqlcdb.ImportBatch, error) {
	batches, err := r.q.ListImportBatchesByEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("list import batches: %w", err)
	}
	return batches, nil
}

func (r *pgRepository) UpdateImportBatchMapping(ctx context.Context, id int64, mapping []byte, status string) (sqlcdb.ImportBatch, error) {
	batch, err := r.q.UpdateImportBatchMapping(ctx, sqlcdb.UpdateImportBatchMappingParams{
		ID:      id,
		Mapping: mapping,
		Status:  pgtype.Text{String: status, Valid: true},
	})
	if err != nil {
		return sqlcdb.ImportBatch{}, fmt.Errorf("update batch mapping: %w", err)
	}
	return batch, nil
}

func (r *pgRepository) UpdateImportBatchStatus(ctx context.Context, id int64, status string, rowsOk, rowsFailed int, report []byte) (sqlcdb.ImportBatch, error) {
	batch, err := r.q.UpdateImportBatchStatus(ctx, sqlcdb.UpdateImportBatchStatusParams{
		ID:         id,
		Status:     pgtype.Text{String: status, Valid: true},
		RowsOk:     pgtype.Int4{Int32: int32(rowsOk), Valid: true},
		RowsFailed: pgtype.Int4{Int32: int32(rowsFailed), Valid: true},
		Report:     report,
	})
	if err != nil {
		return sqlcdb.ImportBatch{}, fmt.Errorf("update batch status: %w", err)
	}
	return batch, nil
}

func (r *pgRepository) DeleteImportBatch(ctx context.Context, id int64) error {
	if err := r.q.DeleteImportBatch(ctx, id); err != nil {
		return fmt.Errorf("delete import batch: %w", err)
	}
	return nil
}

func (r *pgRepository) CreateParticipantRow(ctx context.Context, params sqlcdb.CreateParticipantRowParams) (sqlcdb.ParticipantRow, error) {
	row, err := r.q.CreateParticipantRow(ctx, params)
	if err != nil {
		return sqlcdb.ParticipantRow{}, fmt.Errorf("create participant row: %w", err)
	}
	return row, nil
}

func (r *pgRepository) ListParticipantRowsByBatch(ctx context.Context, batchID int64) ([]sqlcdb.ParticipantRow, error) {
	rows, err := r.q.ListParticipantRowsByBatch(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("list participant rows: %w", err)
	}
	return rows, nil
}
