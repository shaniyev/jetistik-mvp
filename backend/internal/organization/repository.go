package organization

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Repository defines data access for organization operations.
type Repository interface {
	CreateOrganization(ctx context.Context, params sqlcdb.CreateOrganizationParams) (sqlcdb.Organization, error)
	GetOrganizationByID(ctx context.Context, id int64) (sqlcdb.Organization, error)
	ListOrganizations(ctx context.Context, limit, offset int32) ([]sqlcdb.Organization, error)
	CountOrganizations(ctx context.Context) (int64, error)
	UpdateOrganization(ctx context.Context, params sqlcdb.UpdateOrganizationParams) (sqlcdb.Organization, error)
	DeleteOrganization(ctx context.Context, id int64) error
	ListOrganizationMembers(ctx context.Context, orgID int64) ([]sqlcdb.ListOrganizationMembersRow, error)
	AddOrganizationMember(ctx context.Context, params sqlcdb.AddOrganizationMemberParams) (sqlcdb.OrganizationMember, error)
	RemoveOrganizationMember(ctx context.Context, orgID, userID int64) error
	GetUserOrganization(ctx context.Context, userID int64) (sqlcdb.Organization, error)
}

type pgRepository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new organization repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{q: sqlcdb.New(pool)}
}

func (r *pgRepository) CreateOrganization(ctx context.Context, params sqlcdb.CreateOrganizationParams) (sqlcdb.Organization, error) {
	org, err := r.q.CreateOrganization(ctx, params)
	if err != nil {
		return sqlcdb.Organization{}, fmt.Errorf("create organization: %w", err)
	}
	return org, nil
}

func (r *pgRepository) GetOrganizationByID(ctx context.Context, id int64) (sqlcdb.Organization, error) {
	org, err := r.q.GetOrganizationByID(ctx, id)
	if err != nil {
		return sqlcdb.Organization{}, fmt.Errorf("get organization: %w", err)
	}
	return org, nil
}

func (r *pgRepository) ListOrganizations(ctx context.Context, limit, offset int32) ([]sqlcdb.Organization, error) {
	orgs, err := r.q.ListOrganizations(ctx, sqlcdb.ListOrganizationsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	return orgs, nil
}

func (r *pgRepository) CountOrganizations(ctx context.Context) (int64, error) {
	count, err := r.q.CountOrganizations(ctx)
	if err != nil {
		return 0, fmt.Errorf("count organizations: %w", err)
	}
	return count, nil
}

func (r *pgRepository) UpdateOrganization(ctx context.Context, params sqlcdb.UpdateOrganizationParams) (sqlcdb.Organization, error) {
	org, err := r.q.UpdateOrganization(ctx, params)
	if err != nil {
		return sqlcdb.Organization{}, fmt.Errorf("update organization: %w", err)
	}
	return org, nil
}

func (r *pgRepository) DeleteOrganization(ctx context.Context, id int64) error {
	if err := r.q.DeleteOrganization(ctx, id); err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}
	return nil
}

func (r *pgRepository) ListOrganizationMembers(ctx context.Context, orgID int64) ([]sqlcdb.ListOrganizationMembersRow, error) {
	members, err := r.q.ListOrganizationMembers(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	return members, nil
}

func (r *pgRepository) AddOrganizationMember(ctx context.Context, params sqlcdb.AddOrganizationMemberParams) (sqlcdb.OrganizationMember, error) {
	member, err := r.q.AddOrganizationMember(ctx, params)
	if err != nil {
		return sqlcdb.OrganizationMember{}, fmt.Errorf("add member: %w", err)
	}
	return member, nil
}

func (r *pgRepository) RemoveOrganizationMember(ctx context.Context, orgID, userID int64) error {
	if err := r.q.RemoveOrganizationMember(ctx, sqlcdb.RemoveOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
	}); err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	return nil
}

func (r *pgRepository) GetUserOrganization(ctx context.Context, userID int64) (sqlcdb.Organization, error) {
	org, err := r.q.GetUserOrganization(ctx, userID)
	if err != nil {
		return sqlcdb.Organization{}, fmt.Errorf("get user organization: %w", err)
	}
	return org, nil
}
