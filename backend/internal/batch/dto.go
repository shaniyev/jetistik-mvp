package batch

import "time"

// --- Requests ---

type UpdateMappingRequest struct {
	Mapping map[string]string `json:"mapping"`
}

// --- Responses ---

type BatchResponse struct {
	ID         int64             `json:"id"`
	EventID    int64             `json:"event_id"`
	FilePath   string            `json:"file_path"`
	Status     string            `json:"status"`
	RowsTotal  int               `json:"rows_total"`
	RowsOk     int               `json:"rows_ok"`
	RowsFailed int               `json:"rows_failed"`
	Mapping    map[string]string `json:"mapping,omitempty"`
	Tokens     []string          `json:"tokens,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type BatchUploadResponse struct {
	Batch          BatchResponse     `json:"batch"`
	Columns        []string          `json:"columns"`
	DefaultMapping map[string]string `json:"default_mapping"`
	PreviewRows    int               `json:"preview_rows"`
}

// --- Validation ---

func (r UpdateMappingRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if len(r.Mapping) == 0 {
		errs["mapping"] = "mapping is required"
	}
	return errs
}
