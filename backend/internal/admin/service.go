package admin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Service handles admin business logic.
type Service struct {
	q *sqlcdb.Queries
}

// NewService creates a new admin service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{q: sqlcdb.New(pool)}
}

// Stats returns platform-wide statistics.
func (s *Service) Stats(ctx context.Context) (*StatsResponse, error) {
	totalOrgs, err := s.q.CountOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("count organizations: %w", err)
	}
	totalUsers, err := s.q.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	totalCerts, err := s.q.CountAllCertificates(ctx)
	if err != nil {
		return nil, fmt.Errorf("count certificates: %w", err)
	}
	totalEvents, err := s.q.CountAllEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("count events: %w", err)
	}
	certsWeek, err := s.q.CountCertificatesThisWeek(ctx)
	if err != nil {
		return nil, fmt.Errorf("count certificates this week: %w", err)
	}
	orgsWeek, err := s.q.CountOrganizationsThisWeek(ctx)
	if err != nil {
		return nil, fmt.Errorf("count organizations this week: %w", err)
	}

	return &StatsResponse{
		TotalOrganizations: totalOrgs,
		TotalUsers:         totalUsers,
		TotalCertificates:  totalCerts,
		TotalEvents:        totalEvents,
		CertificatesWeek:   certsWeek,
		OrganizationsWeek:  orgsWeek,
	}, nil
}

// ListEvents returns all events (not org-scoped) with pagination.
func (s *Service) ListEvents(ctx context.Context, page, perPage int) ([]EventResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	events, err := s.q.ListAllEvents(ctx, sqlcdb.ListAllEventsParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	total, err := s.q.CountAllEvents(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	result := make([]EventResponse, len(events))
	for i, e := range events {
		resp := EventResponse{
			ID:             e.ID,
			OrganizationID: e.OrganizationID,
			Title:          e.Title,
			City:           e.City.String,
			Description:    e.Description.String,
			Status:         e.Status.String,
			CreatedAt:      e.CreatedAt.Time,
			UpdatedAt:      e.UpdatedAt.Time,
		}
		if e.CreatedBy.Valid {
			id := e.CreatedBy.Int64
			resp.CreatedBy = &id
		}
		if e.Date.Valid {
			resp.Date = e.Date.Time.Format("2006-01-02")
		}
		result[i] = resp
	}
	return result, total, nil
}

// ListCertificates returns all certificates (not org-scoped) with pagination.
func (s *Service) ListCertificates(ctx context.Context, page, perPage int) ([]CertificateResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	certs, err := s.q.ListAllCertificates(ctx, sqlcdb.ListAllCertificatesParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list certificates: %w", err)
	}
	total, err := s.q.CountAllCertificates(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count certificates: %w", err)
	}

	result := make([]CertificateResponse, len(certs))
	for i, c := range certs {
		resp := CertificateResponse{
			ID:      c.ID,
			EventID: c.EventID,
			IIN:     c.Iin.String,
			Name:    c.Name.String,
			Code:    c.Code,
			Status:  c.Status.String,
			CreatedAt: c.CreatedAt.Time,
			UpdatedAt: c.UpdatedAt.Time,
		}
		if c.OrganizationID.Valid {
			id := c.OrganizationID.Int64
			resp.OrganizationID = &id
		}
		result[i] = resp
	}
	return result, total, nil
}

// ListUsers returns all users with pagination.
func (s *Service) ListUsers(ctx context.Context, page, perPage int) ([]UserResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	users, err := s.q.ListUsers(ctx, sqlcdb.ListUsersParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	total, err := s.q.CountUsers(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email.String,
			IIN:       u.Iin.String,
			Role:      u.Role,
			IsActive:  u.IsActive.Bool,
			Language:  u.Language.String,
			CreatedAt: u.CreatedAt.Time,
			UpdatedAt: u.UpdatedAt.Time,
		}
	}
	return result, total, nil
}

// GetUser returns a user by ID.
func (s *Service) GetUser(ctx context.Context, id int64) (*UserResponse, error) {
	u, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email.String,
		IIN:       u.Iin.String,
		Role:      u.Role,
		IsActive:  u.IsActive.Bool,
		Language:  u.Language.String,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}, nil
}

// UpdateUser updates a user's fields (admin action).
func (s *Service) UpdateUser(ctx context.Context, id int64, req UpdateUserRequest) (*UserResponse, error) {
	params := sqlcdb.UpdateUserAdminParams{ID: id}
	if req.Role != nil {
		params.Role = pgtype.Text{String: *req.Role, Valid: true}
	}
	if req.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *req.IsActive, Valid: true}
	}
	if req.Email != nil {
		params.Email = pgtype.Text{String: *req.Email, Valid: true}
	}
	if req.Language != nil {
		params.Language = pgtype.Text{String: *req.Language, Valid: true}
	}

	u, err := s.q.UpdateUserAdmin(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email.String,
		IIN:       u.Iin.String,
		Role:      u.Role,
		IsActive:  u.IsActive.Bool,
		Language:  u.Language.String,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}, nil
}

// DeleteUser deletes a user by ID.
func (s *Service) DeleteUser(ctx context.Context, id int64) error {
	if err := s.q.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
