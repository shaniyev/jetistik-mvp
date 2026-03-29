package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Repository defines data access for user operations.
type Repository interface {
	GetUserByID(ctx context.Context, id int64) (sqlcdb.User, error)
	UpdateUserProfile(ctx context.Context, params sqlcdb.UpdateUserProfileParams) (sqlcdb.UpdateUserProfileRow, error)
	ListTeacherStudents(ctx context.Context, teacherID int64) ([]sqlcdb.TeacherStudent, error)
	AddTeacherStudent(ctx context.Context, teacherID int64, studentIIN string) (sqlcdb.TeacherStudent, error)
	RemoveTeacherStudent(ctx context.Context, teacherID int64, studentIIN string) error
	ListCertificatesByUserID(ctx context.Context, userID int64) ([]sqlcdb.ListCertificatesByUserIDRow, error)
}

type pgRepository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new user repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{q: sqlcdb.New(pool)}
}

func (r *pgRepository) GetUserByID(ctx context.Context, id int64) (sqlcdb.User, error) {
	user, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return sqlcdb.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (r *pgRepository) UpdateUserProfile(ctx context.Context, params sqlcdb.UpdateUserProfileParams) (sqlcdb.UpdateUserProfileRow, error) {
	user, err := r.q.UpdateUserProfile(ctx, params)
	if err != nil {
		return sqlcdb.UpdateUserProfileRow{}, fmt.Errorf("update profile: %w", err)
	}
	return user, nil
}

func (r *pgRepository) ListTeacherStudents(ctx context.Context, teacherID int64) ([]sqlcdb.TeacherStudent, error) {
	students, err := r.q.ListTeacherStudents(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list teacher students: %w", err)
	}
	return students, nil
}

func (r *pgRepository) AddTeacherStudent(ctx context.Context, teacherID int64, studentIIN string) (sqlcdb.TeacherStudent, error) {
	ts, err := r.q.AddTeacherStudent(ctx, sqlcdb.AddTeacherStudentParams{
		TeacherID:  teacherID,
		StudentIin: studentIIN,
	})
	if err != nil {
		return sqlcdb.TeacherStudent{}, fmt.Errorf("add teacher student: %w", err)
	}
	return ts, nil
}

func (r *pgRepository) RemoveTeacherStudent(ctx context.Context, teacherID int64, studentIIN string) error {
	err := r.q.RemoveTeacherStudent(ctx, sqlcdb.RemoveTeacherStudentParams{
		TeacherID:  teacherID,
		StudentIin: studentIIN,
	})
	if err != nil {
		return fmt.Errorf("remove teacher student: %w", err)
	}
	return nil
}

func (r *pgRepository) ListCertificatesByUserID(ctx context.Context, userID int64) ([]sqlcdb.ListCertificatesByUserIDRow, error) {
	certs, err := r.q.ListCertificatesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list certificates by user: %w", err)
	}
	return certs, nil
}
