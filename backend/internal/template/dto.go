package template

import "time"

// --- Responses ---

type TemplateResponse struct {
	ID        int64     `json:"id"`
	EventID   int64     `json:"event_id"`
	FilePath  string    `json:"file_path"`
	Tokens    []string  `json:"tokens"`
	CreatedAt time.Time `json:"created_at"`
}
