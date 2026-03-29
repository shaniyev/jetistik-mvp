package template

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"jetistik/internal/sqlcdb"
	"jetistik/internal/storage"
)

// Service handles template business logic.
type Service struct {
	repo    Repository
	storage *storage.Client
}

// NewService creates a new template service.
func NewService(repo Repository, storage *storage.Client) *Service {
	return &Service{repo: repo, storage: storage}
}

// Upload handles PPTX template upload: stores file in MinIO, extracts tokens, saves record.
func (s *Service) Upload(ctx context.Context, eventID int64, filename string, fileData io.Reader, fileSize int64) (*TemplateResponse, error) {
	// Read file into memory for both upload and token extraction
	data, err := io.ReadAll(fileData)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// Extract tokens from PPTX
	tokens, err := ExtractTokensFromPPTX(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("extract tokens: %w", err)
	}

	// Upload to MinIO
	objectPath := storage.TemplatePath(eventID, filename)
	_, err = s.storage.Upload(ctx, objectPath, bytes.NewReader(data), int64(len(data)), "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	if err != nil {
		return nil, fmt.Errorf("upload template: %w", err)
	}

	// Delete previous templates for this event
	_ = s.repo.DeleteTemplatesByEventID(ctx, eventID)

	// Save template record
	tokensJSON, _ := json.Marshal(tokens)
	tmpl, err := s.repo.CreateTemplate(ctx, sqlcdb.CreateTemplateParams{
		EventID:  eventID,
		FilePath: objectPath,
		Tokens:   tokensJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("save template: %w", err)
	}

	return toTemplateResponse(tmpl, tokens), nil
}

// GetByEventID returns the template for an event.
func (s *Service) GetByEventID(ctx context.Context, eventID int64) (*TemplateResponse, error) {
	tmpl, err := s.repo.GetTemplateByEventID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}
	var tokens []string
	if len(tmpl.Tokens) > 0 {
		_ = json.Unmarshal(tmpl.Tokens, &tokens)
	}
	return toTemplateResponse(tmpl, tokens), nil
}

// Delete deletes a template and its file from MinIO.
func (s *Service) Delete(ctx context.Context, eventID int64) error {
	tmpl, err := s.repo.GetTemplateByEventID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}
	// Delete from MinIO
	_ = s.storage.Delete(ctx, tmpl.FilePath)
	// Delete from DB
	return s.repo.DeleteTemplatesByEventID(ctx, eventID)
}

func toTemplateResponse(t sqlcdb.Template, tokens []string) *TemplateResponse {
	if tokens == nil {
		tokens = []string{}
	}
	return &TemplateResponse{
		ID:        t.ID,
		EventID:   t.EventID,
		FilePath:  t.FilePath,
		Tokens:    tokens,
		CreatedAt: t.CreatedAt.Time,
	}
}
