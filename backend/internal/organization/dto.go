package organization

import "time"

// --- Requests ---

type CreateOrganizationRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type UpdateOrganizationRequest struct {
	Name   *string `json:"name"`
	Domain *string `json:"domain"`
	Status *string `json:"status"`
}

type AddMemberRequest struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

// --- Responses ---

type OrganizationResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain,omitempty"`
	LogoPath  string    `json:"logo_path,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MemberResponse struct {
	ID             int64     `json:"id"`
	OrganizationID int64     `json:"organization_id"`
	UserID         int64     `json:"user_id"`
	Username       string    `json:"username"`
	Email          string    `json:"email,omitempty"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
}

// --- Validation ---

func (r CreateOrganizationRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Name == "" {
		errs["name"] = "name is required"
	}
	return errs
}

func (r UpdateOrganizationRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Status != nil && *r.Status != "" {
		if *r.Status != "active" && *r.Status != "inactive" {
			errs["status"] = "status must be active or inactive"
		}
	}
	return errs
}

func (r AddMemberRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.UserID == 0 {
		errs["user_id"] = "user_id is required"
	}
	if r.Role == "" {
		r.Role = "member"
	}
	if r.Role != "member" && r.Role != "admin" {
		errs["role"] = "role must be member or admin"
	}
	return errs
}
