package event

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Repository defines data access for event operations.
type Repository interface {
	CreateEvent(ctx context.Context, params sqlcdb.CreateEventParams) (sqlcdb.Event, error)
	GetEventByID(ctx context.Context, id int64) (sqlcdb.Event, error)
	ListEventsByOrganization(ctx context.Context, orgID int64, limit, offset int32) ([]sqlcdb.Event, error)
	CountEventsByOrganization(ctx context.Context, orgID int64) (int64, error)
	UpdateEvent(ctx context.Context, params sqlcdb.UpdateEventParams) (sqlcdb.Event, error)
	DeleteEvent(ctx context.Context, id int64) error
}

type pgRepository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new event repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{q: sqlcdb.New(pool)}
}

func (r *pgRepository) CreateEvent(ctx context.Context, params sqlcdb.CreateEventParams) (sqlcdb.Event, error) {
	event, err := r.q.CreateEvent(ctx, params)
	if err != nil {
		return sqlcdb.Event{}, fmt.Errorf("create event: %w", err)
	}
	return event, nil
}

func (r *pgRepository) GetEventByID(ctx context.Context, id int64) (sqlcdb.Event, error) {
	event, err := r.q.GetEventByID(ctx, id)
	if err != nil {
		return sqlcdb.Event{}, fmt.Errorf("get event: %w", err)
	}
	return event, nil
}

func (r *pgRepository) ListEventsByOrganization(ctx context.Context, orgID int64, limit, offset int32) ([]sqlcdb.Event, error) {
	events, err := r.q.ListEventsByOrganization(ctx, sqlcdb.ListEventsByOrganizationParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return events, nil
}

func (r *pgRepository) CountEventsByOrganization(ctx context.Context, orgID int64) (int64, error) {
	count, err := r.q.CountEventsByOrganization(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("count events: %w", err)
	}
	return count, nil
}

func (r *pgRepository) UpdateEvent(ctx context.Context, params sqlcdb.UpdateEventParams) (sqlcdb.Event, error) {
	event, err := r.q.UpdateEvent(ctx, params)
	if err != nil {
		return sqlcdb.Event{}, fmt.Errorf("update event: %w", err)
	}
	return event, nil
}

func (r *pgRepository) DeleteEvent(ctx context.Context, id int64) error {
	if err := r.q.DeleteEvent(ctx, id); err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}
