package certificate

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"jetistik/internal/sqlcdb"
	"jetistik/internal/storage"
)

var (
	ErrNotFound = errors.New("certificate not found")
)

// Service handles certificate business logic.
type Service struct {
	repo    Repository
	storage *storage.Client
	baseURL string
}

// NewService creates a new certificate service.
func NewService(repo Repository, storage *storage.Client, baseURL string) *Service {
	return &Service{repo: repo, storage: storage, baseURL: baseURL}
}

// GetByID returns a certificate by its ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*CertificateResponse, error) {
	cert, err := s.repo.GetCertificateByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get certificate: %w", err)
	}
	return toCertResponse(cert), nil
}

// GetByCode returns a certificate by its verification code.
func (s *Service) GetByCode(ctx context.Context, code string) (*CertificateResponse, error) {
	cert, err := s.repo.GetCertificateByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get certificate by code: %w", err)
	}
	return toCertResponse(cert), nil
}

// Verify looks up a certificate by code and returns verification info with event/org details.
func (s *Service) Verify(ctx context.Context, code string) (*VerifyResponse, error) {
	cert, err := s.repo.GetCertificateByCodeWithDetails(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &VerifyResponse{Valid: false, Code: code, Status: "not_found"}, nil
		}
		return nil, fmt.Errorf("verify: %w", err)
	}
	var payload map[string]interface{}
	if len(cert.Payload) > 0 {
		json.Unmarshal(cert.Payload, &payload)
	}

	return &VerifyResponse{
		Valid:         cert.Status.String == "valid",
		Code:          cert.Code,
		Name:          cert.Name.String,
		IIN:           cert.Iin.String,
		Status:        cert.Status.String,
		RevokedReason: cert.RevokedReason.String,
		EventTitle:    cert.EventTitle,
		OrgName:       cert.OrgName.String,
		Payload:       payload,
		CreatedAt:     cert.CreatedAt.Time,
	}, nil
}

// ListByEvent returns paginated certificates for an event.
func (s *Service) ListByEvent(ctx context.Context, eventID int64, page, perPage int) ([]CertificateResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	certs, err := s.repo.ListCertificatesByEvent(ctx, eventID, int32(perPage), int32(offset))
	if err != nil {
		return nil, 0, fmt.Errorf("list certificates: %w", err)
	}
	total, err := s.repo.CountCertificatesByEvent(ctx, eventID)
	if err != nil {
		return nil, 0, fmt.Errorf("count certificates: %w", err)
	}

	result := make([]CertificateResponse, len(certs))
	for i, c := range certs {
		result[i] = *toCertResponse(c)
	}
	return result, total, nil
}

// Revoke revokes a certificate with a reason.
func (s *Service) Revoke(ctx context.Context, id int64, reason string) (*CertificateResponse, error) {
	cert, err := s.repo.UpdateCertificateStatus(ctx, id, "revoked", reason)
	if err != nil {
		return nil, fmt.Errorf("revoke: %w", err)
	}
	return toCertResponse(cert), nil
}

// Unrevoke restores a revoked certificate.
func (s *Service) Unrevoke(ctx context.Context, id int64) (*CertificateResponse, error) {
	cert, err := s.repo.UpdateCertificateStatus(ctx, id, "valid", "")
	if err != nil {
		return nil, fmt.Errorf("unrevoke: %w", err)
	}
	return toCertResponse(cert), nil
}

// UpdateStatus updates a certificate's status.
func (s *Service) UpdateStatus(ctx context.Context, id int64, status string) (*CertificateResponse, error) {
	cert, err := s.repo.UpdateCertificateStatus(ctx, id, status, "")
	if err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}
	return toCertResponse(cert), nil
}

// Delete deletes a certificate and its PDF from MinIO.
func (s *Service) Delete(ctx context.Context, id int64) error {
	cert, err := s.repo.GetCertificateByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get certificate: %w", err)
	}
	if cert.PdfPath.String != "" {
		_ = s.storage.Delete(ctx, cert.PdfPath.String)
	}
	return s.repo.DeleteCertificate(ctx, id)
}

// DownloadURL returns a presigned download URL for the certificate PDF.
func (s *Service) DownloadURL(ctx context.Context, id int64) (string, error) {
	cert, err := s.repo.GetCertificateByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("get certificate: %w", err)
	}
	if cert.PdfPath.String == "" {
		return "", fmt.Errorf("certificate has no PDF")
	}
	url, err := s.storage.PresignedURL(ctx, cert.PdfPath.String, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("presigned url: %w", err)
	}
	return url, nil
}

// DownloadURLByCode returns a presigned download URL for the certificate PDF by code.
func (s *Service) DownloadURLByCode(ctx context.Context, code string) (string, error) {
	cert, err := s.repo.GetCertificateByCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("get certificate: %w", err)
	}
	if cert.PdfPath.String == "" {
		return "", fmt.Errorf("certificate has no PDF")
	}
	url, err := s.storage.PresignedURL(ctx, cert.PdfPath.String, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("presigned url: %w", err)
	}
	return url, nil
}

// SearchByIIN returns certificates matching an IIN.
func (s *Service) SearchByIIN(ctx context.Context, iin string) ([]SearchResult, error) {
	rows, err := s.repo.SearchCertificatesByIIN(ctx, iin)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	results := make([]SearchResult, len(rows))
	for i, r := range rows {
		iinMasked := r.Iin.String
		if len(iinMasked) == 12 {
			iinMasked = iinMasked[:4] + "****" + iinMasked[8:]
		}
		results[i] = SearchResult{
			ID:         r.ID,
			EventID:    r.EventID,
			IIN:        iinMasked,
			Name:       r.Name.String,
			Code:       r.Code,
			Status:     r.Status.String,
			EventTitle: r.EventTitle,
			OrgName:    r.OrgName.String,
			CreatedAt:  r.CreatedAt.Time,
		}
	}
	return results, nil
}

// UpdateFields updates name/iin of a certificate.
func (s *Service) UpdateFields(ctx context.Context, id int64, name, iin *string) (*CertificateResponse, error) {
	params := sqlcdb.UpdateCertificateFieldsParams{ID: id}
	if name != nil {
		params.Name = pgtype.Text{String: *name, Valid: true}
	}
	if iin != nil {
		params.Iin = pgtype.Text{String: *iin, Valid: true}
	}
	cert, err := s.repo.UpdateCertificateFields(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update fields: %w", err)
	}
	return toCertResponse(cert), nil
}

// DownloadFile returns a reader for the file at the given path.
func (s *Service) DownloadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.storage.Download(ctx, path)
}

// DownloadZipByIIN streams a ZIP archive of all valid certificate PDFs for the given IIN.
func (s *Service) DownloadZipByIIN(ctx context.Context, iin string, w io.Writer) (string, error) {
	rows, err := s.repo.ListCertificatesByIIN(ctx, iin)
	if err != nil {
		return "", fmt.Errorf("list by iin: %w", err)
	}

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, r := range rows {
		pdfPath := r.PdfPath.String
		if pdfPath == "" {
			continue
		}
		reader, err := s.storage.Download(ctx, pdfPath)
		if err != nil {
			continue // skip missing files
		}
		name := r.Name.String
		if name == "" {
			name = "certificate"
		}
		fileName := fmt.Sprintf("%s_%s.pdf", name, r.Code[:8])
		entry, err := zipWriter.Create(fileName)
		if err != nil {
			reader.Close()
			continue
		}
		io.Copy(entry, reader)
		reader.Close()
	}

	maskedIIN := iin
	if len(iin) >= 6 {
		maskedIIN = iin[:4] + "****" + iin[len(iin)-2:]
	}
	return fmt.Sprintf("certificates_%s.zip", maskedIIN), nil
}

func toCertResponse(cert sqlcdb.Certificate) *CertificateResponse {
	resp := &CertificateResponse{
		ID:            cert.ID,
		EventID:       cert.EventID,
		Name:          cert.Name.String,
		Code:          cert.Code,
		PdfPath:       cert.PdfPath.String,
		Status:        cert.Status.String,
		RevokedReason: cert.RevokedReason.String,
		CreatedAt:     cert.CreatedAt.Time,
		UpdatedAt:     cert.UpdatedAt.Time,
	}
	if cert.OrganizationID.Valid {
		id := cert.OrganizationID.Int64
		resp.OrganizationID = &id
	}
	iin := cert.Iin.String
	if len(iin) == 12 {
		iin = iin[:4] + "****" + iin[8:]
	}
	resp.IIN = iin
	if len(cert.Payload) > 0 {
		var payload map[string]interface{}
		if err := json.Unmarshal(cert.Payload, &payload); err == nil {
			resp.Payload = payload
		}
	}
	return resp
}
