package certificate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

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

// Verify looks up a certificate by code and returns verification info.
func (s *Service) Verify(ctx context.Context, code string) (*VerifyResponse, error) {
	cert, err := s.repo.GetCertificateByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &VerifyResponse{Valid: false, Code: code, Status: "not_found"}, nil
		}
		return nil, fmt.Errorf("verify: %w", err)
	}
	return &VerifyResponse{
		Valid:         cert.Status.String == "valid",
		Code:          cert.Code,
		Name:          cert.Name.String,
		Status:        cert.Status.String,
		RevokedReason: cert.RevokedReason.String,
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
