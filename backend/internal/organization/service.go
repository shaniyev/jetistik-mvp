package organization

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"jetistik/internal/sqlcdb"
)

// Service handles organization business logic.
type Service struct {
	repo Repository
}

// NewService creates a new organization service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new organization.
func (s *Service) Create(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error) {
	org, err := s.repo.CreateOrganization(ctx, sqlcdb.CreateOrganizationParams{
		Name:     req.Name,
		Domain:   pgtype.Text{String: req.Domain, Valid: req.Domain != ""},
		LogoPath: pgtype.Text{},
		Status:   pgtype.Text{String: "active", Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create org: %w", err)
	}
	return toOrgResponse(org), nil
}

// GetByID returns an organization by its ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*OrganizationResponse, error) {
	org, err := s.repo.GetOrganizationByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get org: %w", err)
	}
	return toOrgResponse(org), nil
}

// List returns paginated organizations.
func (s *Service) List(ctx context.Context, page, perPage int) ([]OrganizationResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	orgs, err := s.repo.ListOrganizations(ctx, int32(perPage), int32(offset))
	if err != nil {
		return nil, 0, fmt.Errorf("list orgs: %w", err)
	}

	total, err := s.repo.CountOrganizations(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count orgs: %w", err)
	}

	result := make([]OrganizationResponse, len(orgs))
	for i, o := range orgs {
		result[i] = *toOrgResponse(o)
	}
	return result, total, nil
}

// Update updates an organization.
func (s *Service) Update(ctx context.Context, id int64, req UpdateOrganizationRequest) (*OrganizationResponse, error) {
	params := sqlcdb.UpdateOrganizationParams{ID: id}
	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Domain != nil {
		params.Domain = pgtype.Text{String: *req.Domain, Valid: true}
	}
	if req.Status != nil {
		params.Status = pgtype.Text{String: *req.Status, Valid: true}
	}

	org, err := s.repo.UpdateOrganization(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update org: %w", err)
	}
	return toOrgResponse(org), nil
}

// Delete deletes an organization.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteOrganization(ctx, id)
}

// ListMembers returns all members of an organization.
func (s *Service) ListMembers(ctx context.Context, orgID int64) ([]MemberResponse, error) {
	members, err := s.repo.ListOrganizationMembers(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	result := make([]MemberResponse, len(members))
	for i, m := range members {
		result[i] = MemberResponse{
			ID:             m.ID,
			OrganizationID: m.OrganizationID,
			UserID:         m.UserID,
			Username:       m.Username,
			Email:          m.Email.String,
			Role:           m.Role.String,
			CreatedAt:      m.CreatedAt.Time,
		}
	}
	return result, nil
}

// AddMember adds a user to an organization.
func (s *Service) AddMember(ctx context.Context, orgID int64, req AddMemberRequest) (*MemberResponse, error) {
	role := req.Role
	if role == "" {
		role = "member"
	}
	member, err := s.repo.AddOrganizationMember(ctx, sqlcdb.AddOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         req.UserID,
		Role:           pgtype.Text{String: role, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("add member: %w", err)
	}
	return &MemberResponse{
		ID:             member.ID,
		OrganizationID: member.OrganizationID,
		UserID:         member.UserID,
		Role:           member.Role.String,
		CreatedAt:      member.CreatedAt.Time,
	}, nil
}

// RemoveMember removes a user from an organization.
func (s *Service) RemoveMember(ctx context.Context, orgID, userID int64) error {
	return s.repo.RemoveOrganizationMember(ctx, orgID, userID)
}

// GetUserOrg returns the organization the user belongs to.
func (s *Service) GetUserOrg(ctx context.Context, userID int64) (*OrganizationResponse, error) {
	org, err := s.repo.GetUserOrganization(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user org: %w", err)
	}
	return toOrgResponse(org), nil
}

func toOrgResponse(o sqlcdb.Organization) *OrganizationResponse {
	return &OrganizationResponse{
		ID:        o.ID,
		Name:      o.Name,
		Domain:    o.Domain.String,
		LogoPath:  o.LogoPath.String,
		Status:    o.Status.String,
		CreatedAt: o.CreatedAt.Time,
		UpdatedAt: o.UpdatedAt.Time,
	}
}
