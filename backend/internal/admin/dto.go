package admin

import "time"

// --- Responses ---

// StatsResponse holds platform-wide statistics.
type StatsResponse struct {
	TotalOrganizations int64 `json:"total_organizations"`
	TotalUsers         int64 `json:"total_users"`
	TotalCertificates  int64 `json:"total_certificates"`
	TotalEvents        int64 `json:"total_events"`
	CertificatesWeek   int64 `json:"certificates_this_week"`
	OrganizationsWeek  int64 `json:"organizations_this_week"`
}

// UserResponse represents a user in admin listings.
type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	IIN       string    `json:"iin,omitempty"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EventResponse represents an event in admin listings.
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

// CertificateResponse represents a certificate in admin listings.
type CertificateResponse struct {
	ID             int64     `json:"id"`
	EventID        int64     `json:"event_id"`
	OrganizationID *int64    `json:"organization_id,omitempty"`
	IIN            string    `json:"iin,omitempty"`
	Name           string    `json:"name,omitempty"`
	Code           string    `json:"code"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// --- Requests ---

// UpdateUserRequest holds fields that an admin can update on a user.
type UpdateUserRequest struct {
	Role     *string `json:"role"`
	IsActive *bool   `json:"is_active"`
	Email    *string `json:"email"`
	Language *string `json:"language"`
}

// Validate checks the update user request.
func (r UpdateUserRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Role != nil && *r.Role != "" {
		valid := map[string]bool{"admin": true, "staff": true, "teacher": true, "student": true}
		if !valid[*r.Role] {
			errs["role"] = "role must be admin, staff, teacher, or student"
		}
	}
	if r.Language != nil && *r.Language != "" {
		if *r.Language != "kz" && *r.Language != "ru" && *r.Language != "en" {
			errs["language"] = "language must be kz, ru, or en"
		}
	}
	return errs
}
