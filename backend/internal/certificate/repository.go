package certificate

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Repository defines data access for certificate operations.
type Repository interface {
	CreateCertificate(ctx context.Context, params sqlcdb.CreateCertificateParams) (sqlcdb.Certificate, error)
	GetCertificateByID(ctx context.Context, id int64) (sqlcdb.Certificate, error)
	GetCertificateByCode(ctx context.Context, code string) (sqlcdb.Certificate, error)
	GetCertificateByCodeWithDetails(ctx context.Context, code string) (sqlcdb.GetCertificateByCodeWithDetailsRow, error)
	ListCertificatesByEvent(ctx context.Context, eventID int64, limit, offset int32) ([]sqlcdb.Certificate, error)
	CountCertificatesByEvent(ctx context.Context, eventID int64) (int64, error)
	UpdateCertificateStatus(ctx context.Context, id int64, status, revokedReason string) (sqlcdb.Certificate, error)
	DeleteCertificate(ctx context.Context, id int64) error
	SearchCertificatesByIIN(ctx context.Context, iin string) ([]sqlcdb.SearchCertificatesByIINRow, error)
}

type pgRepository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new certificate repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{q: sqlcdb.New(pool)}
}

func (r *pgRepository) CreateCertificate(ctx context.Context, params sqlcdb.CreateCertificateParams) (sqlcdb.Certificate, error) {
	cert, err := r.q.CreateCertificate(ctx, params)
	if err != nil {
		return sqlcdb.Certificate{}, fmt.Errorf("create certificate: %w", err)
	}
	return cert, nil
}

func (r *pgRepository) GetCertificateByID(ctx context.Context, id int64) (sqlcdb.Certificate, error) {
	cert, err := r.q.GetCertificateByID(ctx, id)
	if err != nil {
		return sqlcdb.Certificate{}, fmt.Errorf("get certificate: %w", err)
	}
	return cert, nil
}

func (r *pgRepository) GetCertificateByCode(ctx context.Context, code string) (sqlcdb.Certificate, error) {
	cert, err := r.q.GetCertificateByCode(ctx, code)
	if err != nil {
		return sqlcdb.Certificate{}, fmt.Errorf("get certificate by code: %w", err)
	}
	return cert, nil
}

func (r *pgRepository) GetCertificateByCodeWithDetails(ctx context.Context, code string) (sqlcdb.GetCertificateByCodeWithDetailsRow, error) {
	row, err := r.q.GetCertificateByCodeWithDetails(ctx, code)
	if err != nil {
		return sqlcdb.GetCertificateByCodeWithDetailsRow{}, fmt.Errorf("get certificate by code with details: %w", err)
	}
	return row, nil
}

func (r *pgRepository) ListCertificatesByEvent(ctx context.Context, eventID int64, limit, offset int32) ([]sqlcdb.Certificate, error) {
	certs, err := r.q.ListCertificatesByEvent(ctx, sqlcdb.ListCertificatesByEventParams{
		EventID: eventID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list certificates: %w", err)
	}
	return certs, nil
}

func (r *pgRepository) CountCertificatesByEvent(ctx context.Context, eventID int64) (int64, error) {
	count, err := r.q.CountCertificatesByEvent(ctx, eventID)
	if err != nil {
		return 0, fmt.Errorf("count certificates: %w", err)
	}
	return count, nil
}

func (r *pgRepository) UpdateCertificateStatus(ctx context.Context, id int64, status, revokedReason string) (sqlcdb.Certificate, error) {
	cert, err := r.q.UpdateCertificateStatus(ctx, sqlcdb.UpdateCertificateStatusParams{
		ID:            id,
		Status:        pgtype.Text{String: status, Valid: true},
		RevokedReason: pgtype.Text{String: revokedReason, Valid: revokedReason != ""},
	})
	if err != nil {
		return sqlcdb.Certificate{}, fmt.Errorf("update certificate status: %w", err)
	}
	return cert, nil
}

func (r *pgRepository) DeleteCertificate(ctx context.Context, id int64) error {
	if err := r.q.DeleteCertificate(ctx, id); err != nil {
		return fmt.Errorf("delete certificate: %w", err)
	}
	return nil
}

func (r *pgRepository) SearchCertificatesByIIN(ctx context.Context, iin string) ([]sqlcdb.SearchCertificatesByIINRow, error) {
	results, err := r.q.SearchCertificatesByIIN(ctx, pgtype.Text{String: iin, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("search certificates: %w", err)
	}
	return results, nil
}
