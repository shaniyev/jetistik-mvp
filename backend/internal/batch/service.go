package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"jetistik/internal/sqlcdb"
	"jetistik/internal/storage"
)

// Service handles batch business logic.
type Service struct {
	repo    Repository
	storage *storage.Client
}

// NewService creates a new batch service.
func NewService(repo Repository, storage *storage.Client) *Service {
	return &Service{repo: repo, storage: storage}
}

// Upload handles CSV/XLSX batch upload: parses file, stores in MinIO, creates batch + participant rows.
func (s *Service) Upload(ctx context.Context, eventID int64, filename string, fileData io.Reader, fileSize int64, templateTokens []string) (*BatchUploadResponse, error) {
	// Read file into memory
	data, err := io.ReadAll(fileData)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// Parse file
	var result *ParseResult
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".csv"):
		result, err = ParseCSV(bytes.NewReader(data))
	case strings.HasSuffix(lower, ".xlsx"):
		result, err = ParseXLSX(bytes.NewReader(data))
	default:
		return nil, fmt.Errorf("unsupported file format: only CSV and XLSX are supported")
	}
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	if len(result.Rows) == 0 {
		return nil, fmt.Errorf("file contains no data rows")
	}

	// Upload to MinIO
	objectPath := storage.ImportPath(eventID, filename)
	_, err = s.storage.Upload(ctx, objectPath, bytes.NewReader(data), int64(len(data)), "application/octet-stream")
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	// Create default mapping
	defMapping := DefaultMapping(result.Columns, templateTokens)

	// Create batch record
	tokensJSON, _ := json.Marshal(result.Columns)
	batchRecord, err := s.repo.CreateImportBatch(ctx, sqlcdb.CreateImportBatchParams{
		EventID:   eventID,
		FilePath:  objectPath,
		Status:    pgtype.Text{String: "uploaded", Valid: true},
		RowsTotal: pgtype.Int4{Int32: int32(len(result.Rows)), Valid: true},
		Tokens:    tokensJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("create batch: %w", err)
	}

	// Create participant rows
	for _, row := range result.Rows {
		payloadJSON, _ := json.Marshal(row)
		name := row["name"]
		if name == "" {
			name = row["fullname"]
		}
		if name == "" {
			name = row["fio"]
		}
		if name == "" {
			name = row["\u0444\u0438\u043e"]
		}
		iin := row["iin"]
		if iin == "" {
			iin = row["\u0418\u0418\u041d"]
		}

		_, err := s.repo.CreateParticipantRow(ctx, sqlcdb.CreateParticipantRowParams{
			BatchID: batchRecord.ID,
			Iin:     pgtype.Text{String: iin, Valid: iin != ""},
			Name:    pgtype.Text{String: name, Valid: name != ""},
			Payload: payloadJSON,
			Status:  pgtype.Text{String: "pending", Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("create participant row: %w", err)
		}
	}

	return &BatchUploadResponse{
		Batch:          *toBatchResponse(batchRecord),
		Columns:        result.Columns,
		DefaultMapping: defMapping,
		PreviewRows:    len(result.Rows),
	}, nil
}

// GetByID returns a batch by its ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*BatchResponse, error) {
	batchRecord, err := s.repo.GetImportBatchByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get batch: %w", err)
	}
	return toBatchResponse(batchRecord), nil
}

// ListByEvent returns all batches for an event.
func (s *Service) ListByEvent(ctx context.Context, eventID int64) ([]BatchResponse, error) {
	batches, err := s.repo.ListImportBatchesByEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("list batches: %w", err)
	}
	result := make([]BatchResponse, len(batches))
	for i, b := range batches {
		result[i] = *toBatchResponse(b)
	}
	return result, nil
}

// UpdateMapping saves the column-to-token mapping for a batch.
func (s *Service) UpdateMapping(ctx context.Context, batchID int64, mapping map[string]string) (*BatchResponse, error) {
	mappingJSON, _ := json.Marshal(mapping)
	batchRecord, err := s.repo.UpdateImportBatchMapping(ctx, batchID, mappingJSON, "mapped")
	if err != nil {
		return nil, fmt.Errorf("update mapping: %w", err)
	}
	return toBatchResponse(batchRecord), nil
}

// Delete deletes a batch and its file from MinIO.
func (s *Service) Delete(ctx context.Context, id int64) error {
	batchRecord, err := s.repo.GetImportBatchByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get batch: %w", err)
	}
	_ = s.storage.Delete(ctx, batchRecord.FilePath)
	return s.repo.DeleteImportBatch(ctx, id)
}

func toBatchResponse(b sqlcdb.ImportBatch) *BatchResponse {
	resp := &BatchResponse{
		ID:         b.ID,
		EventID:    b.EventID,
		FilePath:   b.FilePath,
		Status:     b.Status.String,
		RowsTotal:  int(b.RowsTotal.Int32),
		RowsOk:     int(b.RowsOk.Int32),
		RowsFailed: int(b.RowsFailed.Int32),
		CreatedAt:  b.CreatedAt.Time,
		UpdatedAt:  b.UpdatedAt.Time,
	}
	if len(b.Mapping) > 0 {
		var mapping map[string]string
		if err := json.Unmarshal(b.Mapping, &mapping); err == nil {
			resp.Mapping = mapping
		}
	}
	if len(b.Tokens) > 0 {
		var tokens []string
		if err := json.Unmarshal(b.Tokens, &tokens); err == nil {
			resp.Tokens = tokens
		}
	}
	return resp
}
