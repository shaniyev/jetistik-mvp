package event

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"jetistik/internal/sqlcdb"
)

// Service handles event business logic.
type Service struct {
	repo Repository
}

// NewService creates a new event service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new event for the given organization.
func (s *Service) Create(ctx context.Context, orgID, createdBy int64, req CreateEventRequest) (*EventResponse, error) {
	var dateVal pgtype.Date
	if req.Date != "" {
		t, err := time.Parse("2006-01-02", req.Date)
		if err == nil {
			dateVal = pgtype.Date{Time: t, Valid: true}
		}
	}

	event, err := s.repo.CreateEvent(ctx, sqlcdb.CreateEventParams{
		OrganizationID: orgID,
		CreatedBy:      pgtype.Int8{Int64: createdBy, Valid: true},
		Title:          req.Title,
		Date:           dateVal,
		City:           pgtype.Text{String: req.City, Valid: req.City != ""},
		Description:    pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Status:         pgtype.Text{String: "active", Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	return toEventResponse(event), nil
}

// GetByID returns an event by its ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*EventResponse, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	return toEventResponse(event), nil
}

// ListByOrg returns paginated events for an organization.
func (s *Service) ListByOrg(ctx context.Context, orgID int64, page, perPage int) ([]EventResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	events, err := s.repo.ListEventsByOrganization(ctx, orgID, int32(perPage), int32(offset))
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	total, err := s.repo.CountEventsByOrganization(ctx, orgID)
	if err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	result := make([]EventResponse, len(events))
	for i, e := range events {
		result[i] = *toEventResponse(e)
	}
	return result, total, nil
}

// Update updates an event.
func (s *Service) Update(ctx context.Context, id int64, req UpdateEventRequest) (*EventResponse, error) {
	params := sqlcdb.UpdateEventParams{ID: id}
	if req.Title != nil {
		params.Title = pgtype.Text{String: *req.Title, Valid: true}
	}
	if req.Date != nil {
		t, err := time.Parse("2006-01-02", *req.Date)
		if err == nil {
			params.Date = pgtype.Date{Time: t, Valid: true}
		}
	}
	if req.City != nil {
		params.City = pgtype.Text{String: *req.City, Valid: true}
	}
	if req.Description != nil {
		params.Description = pgtype.Text{String: *req.Description, Valid: true}
	}
	if req.Status != nil {
		params.Status = pgtype.Text{String: *req.Status, Valid: true}
	}

	event, err := s.repo.UpdateEvent(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}
	return toEventResponse(event), nil
}

// Delete deletes an event.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteEvent(ctx, id)
}

func toEventResponse(e sqlcdb.Event) *EventResponse {
	resp := &EventResponse{
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
	return resp
}
