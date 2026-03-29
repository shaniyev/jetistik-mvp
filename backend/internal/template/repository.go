package template

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Repository defines data access for template operations.
type Repository interface {
	CreateTemplate(ctx context.Context, params sqlcdb.CreateTemplateParams) (sqlcdb.Template, error)
	GetTemplateByEventID(ctx context.Context, eventID int64) (sqlcdb.Template, error)
	DeleteTemplatesByEventID(ctx context.Context, eventID int64) error
}

type pgRepository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new template repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{q: sqlcdb.New(pool)}
}

func (r *pgRepository) CreateTemplate(ctx context.Context, params sqlcdb.CreateTemplateParams) (sqlcdb.Template, error) {
	tmpl, err := r.q.CreateTemplate(ctx, params)
	if err != nil {
		return sqlcdb.Template{}, fmt.Errorf("create template: %w", err)
	}
	return tmpl, nil
}

func (r *pgRepository) GetTemplateByEventID(ctx context.Context, eventID int64) (sqlcdb.Template, error) {
	tmpl, err := r.q.GetTemplateByEventID(ctx, eventID)
	if err != nil {
		return sqlcdb.Template{}, fmt.Errorf("get template: %w", err)
	}
	return tmpl, nil
}

func (r *pgRepository) DeleteTemplatesByEventID(ctx context.Context, eventID int64) error {
	if err := r.q.DeleteTemplatesByEventID(ctx, eventID); err != nil {
		return fmt.Errorf("delete templates: %w", err)
	}
	return nil
}
