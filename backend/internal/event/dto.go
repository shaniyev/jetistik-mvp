package event

import "time"

// --- Requests ---

type CreateEventRequest struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	City        string `json:"city"`
	Description string `json:"description"`
}

type UpdateEventRequest struct {
	Title       *string `json:"title"`
	Date        *string `json:"date"`
	City        *string `json:"city"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

// --- Responses ---

type EventResponse struct {
	ID             int64     `json:"id"`
	OrganizationID int64     `json:"organization_id"`
	CreatedBy      *int64    `json:"created_by,omitempty"`
	Title          string    `json:"title"`
	Date           string    `json:"date,omitempty"`
	City           string    `json:"city,omitempty"`
	Description    string    `json:"description,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// --- Validation ---

func (r CreateEventRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Title == "" {
		errs["title"] = "title is required"
	}
	return errs
}

func (r UpdateEventRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Status != nil && *r.Status != "" {
		if *r.Status != "active" && *r.Status != "archived" {
			errs["status"] = "status must be active or archived"
		}
	}
	return errs
}
