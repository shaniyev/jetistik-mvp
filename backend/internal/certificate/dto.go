package certificate

import "time"

// --- Requests ---

type UpdateCertificateRequest struct {
	Name   *string `json:"name"`
	IIN    *string `json:"iin"`
	Status *string `json:"status"`
}

type RevokeRequest struct {
	Reason string `json:"reason"`
}

// --- Responses ---

type CertificateResponse struct {
	ID             int64                  `json:"id"`
	EventID        int64                  `json:"event_id"`
	OrganizationID *int64                 `json:"organization_id,omitempty"`
	IIN            string                 `json:"iin,omitempty"`
	Name           string                 `json:"name"`
	Code           string                 `json:"code"`
	PdfPath        string                 `json:"pdf_path,omitempty"`
	Status         string                 `json:"status"`
	RevokedReason  string                 `json:"revoked_reason,omitempty"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type VerifyResponse struct {
	Valid         bool                   `json:"valid"`
	Code          string                 `json:"code"`
	Name          string                 `json:"name"`
	IIN           string                 `json:"iin,omitempty"`
	EventTitle    string                 `json:"event_title,omitempty"`
	OrgName       string                 `json:"org_name,omitempty"`
	Status        string                 `json:"status"`
	RevokedReason string                 `json:"revoked_reason,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

type SearchResult struct {
	ID         int64     `json:"id"`
	EventID    int64     `json:"event_id"`
	IIN        string    `json:"iin"`
	Name       string    `json:"name"`
	Code       string    `json:"code"`
	Status     string    `json:"status"`
	EventTitle string    `json:"event_title"`
	OrgName    string    `json:"org_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// --- Validation ---

func (r RevokeRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Reason == "" {
		errs["reason"] = "revoke reason is required"
	}
	return errs
}
