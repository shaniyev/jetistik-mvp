package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"

	"jetistik/internal/platform/config"
	"jetistik/internal/sqlcdb"
	"jetistik/internal/storage"
)

type generateHandler struct {
	q       *sqlcdb.Queries
	storage *storage.Client
	cfg     *config.Config
	rdb     *redis.Client
}

// ProgressEvent is published to Redis pub/sub for SSE streaming.
type ProgressEvent struct {
	RowID    int64  `json:"row_id"`
	RowName  string `json:"row_name,omitempty"`
	RowIIN   string `json:"row_iin,omitempty"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Progress int    `json:"progress"`
	Total    int    `json:"total"`
	RowsOk   int   `json:"rows_ok"`
	RowsFailed int `json:"rows_failed"`
}

func progressChannel(batchID int64) string {
	return fmt.Sprintf("batch:progress:%d", batchID)
}

// HandleGenerateBatch processes the generate_batch task.
func (h *generateHandler) HandleGenerateBatch(ctx context.Context, t *asynq.Task) error {
	var payload GenerateBatchPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	batchID := payload.BatchID
	slog.Info("starting batch generation", "batch_id", batchID)

	// Load batch
	batch, err := h.q.GetImportBatchByID(ctx, batchID)
	if err != nil {
		return fmt.Errorf("get batch: %w", err)
	}

	// Load event
	event, err := h.q.GetEventByID(ctx, batch.EventID)
	if err != nil {
		return fmt.Errorf("get event: %w", err)
	}

	// Load template
	tmpl, err := h.q.GetTemplateByEventID(ctx, batch.EventID)
	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}

	// Download template PPTX from MinIO
	tmplReader, err := h.storage.Download(ctx, tmpl.FilePath)
	if err != nil {
		h.failBatch(ctx, batchID, "template download failed")
		return fmt.Errorf("download template: %w", err)
	}
	tmplData, err := io.ReadAll(tmplReader)
	tmplReader.Close()
	if err != nil {
		h.failBatch(ctx, batchID, "template read failed")
		return fmt.Errorf("read template: %w", err)
	}

	// Update batch status to generating
	h.q.UpdateImportBatchStatus(ctx, sqlcdb.UpdateImportBatchStatusParams{
		ID:         batchID,
		Status:     pgtype.Text{String: "generating", Valid: true},
		RowsOk:     pgtype.Int4{Int32: 0, Valid: true},
		RowsFailed: pgtype.Int4{Int32: 0, Valid: true},
		Report:     []byte("{}"),
	})

	// Parse mapping
	var mapping map[string]string
	if len(batch.Mapping) > 0 {
		_ = json.Unmarshal(batch.Mapping, &mapping)
	}

	// Load participant rows
	rows, err := h.q.ListParticipantRowsByBatch(ctx, batchID)
	if err != nil {
		h.failBatch(ctx, batchID, "failed to load rows")
		return fmt.Errorf("list rows: %w", err)
	}

	total := len(rows)
	rowsOk := 0
	rowsFailed := 0

	for i, row := range rows {
		err := h.processRow(ctx, row, batch, event, tmplData, mapping)

		status := "ok"
		errMsg := ""
		if err != nil {
			status = "failed"
			errMsg = err.Error()
			if len(errMsg) > 2000 {
				errMsg = errMsg[:2000]
			}
			rowsFailed++
			slog.Error("row generation failed", "row_id", row.ID, "error", err)
		} else {
			rowsOk++
		}

		// Update participant row status
		h.q.UpdateParticipantRowStatus(ctx, sqlcdb.UpdateParticipantRowStatusParams{
			ID:     row.ID,
			Status: pgtype.Text{String: status, Valid: true},
			Error:  pgtype.Text{String: errMsg, Valid: errMsg != ""},
		})

		// Update batch counters
		h.q.UpdateImportBatchStatus(ctx, sqlcdb.UpdateImportBatchStatusParams{
			ID:         batchID,
			Status:     pgtype.Text{String: "generating", Valid: true},
			RowsOk:     pgtype.Int4{Int32: int32(rowsOk), Valid: true},
			RowsFailed: pgtype.Int4{Int32: int32(rowsFailed), Valid: true},
			Report:     []byte("{}"),
		})

		// Publish progress event
		evt := ProgressEvent{
			RowID:      row.ID,
			RowName:    row.Name.String,
			RowIIN:     row.Iin.String,
			Status:     status,
			Error:      errMsg,
			Progress:   i + 1,
			Total:      total,
			RowsOk:     rowsOk,
			RowsFailed: rowsFailed,
		}
		h.publishProgress(ctx, batchID, evt)
	}

	// Final batch status
	finalStatus := "done"
	if rowsFailed > 0 && rowsOk > 0 {
		finalStatus = "done_with_errors"
	} else if rowsFailed > 0 && rowsOk == 0 {
		finalStatus = "failed"
	}

	report, _ := json.Marshal(map[string]int{"ok": rowsOk, "failed": rowsFailed})
	h.q.UpdateImportBatchStatus(ctx, sqlcdb.UpdateImportBatchStatusParams{
		ID:         batchID,
		Status:     pgtype.Text{String: finalStatus, Valid: true},
		RowsOk:     pgtype.Int4{Int32: int32(rowsOk), Valid: true},
		RowsFailed: pgtype.Int4{Int32: int32(rowsFailed), Valid: true},
		Report:     report,
	})

	// Publish completion event
	h.publishProgress(ctx, batchID, ProgressEvent{
		Status:     "complete",
		Progress:   total,
		Total:      total,
		RowsOk:     rowsOk,
		RowsFailed: rowsFailed,
	})

	slog.Info("batch generation complete", "batch_id", batchID, "ok", rowsOk, "failed", rowsFailed)
	return nil
}

func (h *generateHandler) processRow(
	ctx context.Context,
	row sqlcdb.ParticipantRow,
	batch sqlcdb.ImportBatch,
	event sqlcdb.Event,
	tmplData []byte,
	mapping map[string]string,
) error {
	// Build token values from mapping + payload
	var payload map[string]string
	if len(row.Payload) > 0 {
		_ = json.Unmarshal(row.Payload, &payload)
	}
	if payload == nil {
		payload = make(map[string]string)
	}

	tokenValues := make(map[string]string)
	for token, col := range mapping {
		if col == "" {
			continue
		}
		tokenValues[token] = payload[col]
	}
	// Fallback: ensure fname has a value
	if tokenValues["fname"] == "" {
		tokenValues["fname"] = row.Name.String
	}

	// Generate verification code
	code := uuid.New().String()
	code = strings.ReplaceAll(code, "-", "")

	// Generate QR code
	verifyURL := strings.TrimRight(h.cfg.PublicBaseURL, "/") + "/verify/" + code
	qrPNG, err := makeQRPNGBytes(verifyURL)
	if err != nil {
		return fmt.Errorf("generate qr: %w", err)
	}

	// Process PPTX: replace tokens + insert QR
	pptxData, err := replaceTokensInPPTX(tmplData, tokenValues, qrPNG)
	if err != nil {
		return fmt.Errorf("process pptx: %w", err)
	}

	// Convert to PDF via Gotenberg
	pdfData, err := convertPPTXToPDF(pptxData, h.cfg.GotenbergURL)
	if err != nil {
		return fmt.Errorf("convert to pdf: %w", err)
	}

	// Upload PDF to MinIO
	pdfPath := storage.CertificatePath(batch.EventID, code)
	_, err = h.storage.Upload(ctx, pdfPath, bytes.NewReader(pdfData), int64(len(pdfData)), "application/pdf")
	if err != nil {
		return fmt.Errorf("upload pdf: %w", err)
	}

	// Build certificate payload
	certPayload := make(map[string]interface{})
	for k, v := range payload {
		certPayload[k] = v
	}
	certPayload["event_title"] = event.Title
	if event.Date.Valid {
		certPayload["event_date"] = event.Date.Time.Format("2006-01-02")
	}
	certPayload["event_city"] = event.City.String

	// Load organization name
	if event.OrganizationID > 0 {
		org, err := h.q.GetOrganizationByID(ctx, event.OrganizationID)
		if err == nil {
			certPayload["organization_name"] = org.Name
		}
	}

	certPayloadJSON, _ := json.Marshal(certPayload)

	// Create certificate record
	_, err = h.q.CreateCertificate(ctx, sqlcdb.CreateCertificateParams{
		EventID:        batch.EventID,
		OrganizationID: pgtype.Int8{Int64: event.OrganizationID, Valid: event.OrganizationID > 0},
		Iin:            row.Iin,
		Name:           row.Name,
		Code:           code,
		PdfPath:        pgtype.Text{String: pdfPath, Valid: true},
		Status:         pgtype.Text{String: "valid", Valid: true},
		Payload:        certPayloadJSON,
	})
	if err != nil {
		return fmt.Errorf("create certificate: %w", err)
	}

	return nil
}

func (h *generateHandler) failBatch(ctx context.Context, batchID int64, errMsg string) {
	report, _ := json.Marshal(map[string]string{"error": errMsg})
	h.q.UpdateImportBatchStatus(ctx, sqlcdb.UpdateImportBatchStatusParams{
		ID:         batchID,
		Status:     pgtype.Text{String: "failed", Valid: true},
		RowsOk:     pgtype.Int4{Int32: 0, Valid: true},
		RowsFailed: pgtype.Int4{Int32: 0, Valid: true},
		Report:     report,
	})
}

func (h *generateHandler) publishProgress(ctx context.Context, batchID int64, evt ProgressEvent) {
	data, err := json.Marshal(evt)
	if err != nil {
		slog.Error("marshal progress event", "error", err)
		return
	}
	channel := progressChannel(batchID)
	if err := h.rdb.Publish(ctx, channel, string(data)).Err(); err != nil {
		slog.Error("publish progress", "error", err, "channel", channel)
	}
}
