package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"jetistik/internal/sqlcdb"
)

// Service handles user business logic.
type Service struct {
	repo Repository
}

// NewService creates a new user service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetProfile returns the user's profile.
func (s *Service) GetProfile(ctx context.Context, userID int64) (*ProfileResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return toProfileResponse(user), nil
}

// UpdateProfile updates the user's profile fields.
func (s *Service) UpdateProfile(ctx context.Context, userID int64, req UpdateProfileRequest) (*ProfileResponse, error) {
	params := sqlcdb.UpdateUserProfileParams{
		ID: userID,
	}
	if req.Email != nil {
		params.Email = pgtype.Text{String: *req.Email, Valid: true}
	}
	if req.IIN != nil {
		params.Iin = pgtype.Text{String: *req.IIN, Valid: true}
	}
	if req.Language != nil {
		params.Language = pgtype.Text{String: *req.Language, Valid: true}
	}

	row, err := s.repo.UpdateUserProfile(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return toProfileResponseFromRow(row), nil
}

// ListStudents returns the teacher's linked students.
func (s *Service) ListStudents(ctx context.Context, teacherID int64) ([]TeacherStudentResponse, error) {
	students, err := s.repo.ListTeacherStudents(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}

	result := make([]TeacherStudentResponse, len(students))
	for i, ts := range students {
		result[i] = TeacherStudentResponse{
			ID:         ts.ID,
			TeacherID:  ts.TeacherID,
			StudentIIN: ts.StudentIin,
			CreatedAt:  ts.CreatedAt.Time,
		}
	}
	return result, nil
}

// AddStudent links a student IIN to the teacher.
func (s *Service) AddStudent(ctx context.Context, teacherID int64, studentIIN string) (*TeacherStudentResponse, error) {
	ts, err := s.repo.AddTeacherStudent(ctx, teacherID, studentIIN)
	if err != nil {
		return nil, fmt.Errorf("add student: %w", err)
	}
	return &TeacherStudentResponse{
		ID:         ts.ID,
		TeacherID:  ts.TeacherID,
		StudentIIN: ts.StudentIin,
		CreatedAt:  ts.CreatedAt.Time,
	}, nil
}

// RemoveStudent unlinks a student IIN from the teacher.
func (s *Service) RemoveStudent(ctx context.Context, teacherID int64, studentIIN string) error {
	return s.repo.RemoveTeacherStudent(ctx, teacherID, studentIIN)
}

func toProfileResponse(u sqlcdb.User) *ProfileResponse {
	iin := u.Iin.String
	if len(iin) == 12 {
		iin = iin[:4] + "****" + iin[8:]
	}
	return &ProfileResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email.String,
		IIN:       iin,
		Role:      u.Role,
		Language:  u.Language.String,
		CreatedAt: u.CreatedAt.Time,
	}
}

func toProfileResponseFromRow(u sqlcdb.UpdateUserProfileRow) *ProfileResponse {
	iin := u.Iin.String
	if len(iin) == 12 {
		iin = iin[:4] + "****" + iin[8:]
	}
	return &ProfileResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email.String,
		IIN:       iin,
		Role:      u.Role,
		Language:  u.Language.String,
		CreatedAt: u.CreatedAt.Time,
	}
}
