package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"jetistik/internal/sqlcdb"
)

// Repository defines data access for auth operations.
type Repository interface {
	CreateUser(ctx context.Context, params sqlcdb.CreateUserParams) (sqlcdb.User, error)
	GetUserByUsername(ctx context.Context, username string) (sqlcdb.User, error)
	GetUserByID(ctx context.Context, id int64) (sqlcdb.User, error)
	UpdateUserPassword(ctx context.Context, id int64, password string) error
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateRefreshToken(ctx context.Context, params sqlcdb.CreateRefreshTokenParams) (sqlcdb.RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, hash string) (sqlcdb.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, hash string) error
	DeleteRefreshTokensByUser(ctx context.Context, userID int64) error
}

type pgRepository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new auth repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{q: sqlcdb.New(pool)}
}

func (r *pgRepository) CreateUser(ctx context.Context, params sqlcdb.CreateUserParams) (sqlcdb.User, error) {
	user, err := r.q.CreateUser(ctx, params)
	if err != nil {
		return sqlcdb.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *pgRepository) GetUserByUsername(ctx context.Context, username string) (sqlcdb.User, error) {
	user, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		return sqlcdb.User{}, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (r *pgRepository) GetUserByID(ctx context.Context, id int64) (sqlcdb.User, error) {
	user, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return sqlcdb.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *pgRepository) UpdateUserPassword(ctx context.Context, id int64, password string) error {
	err := r.q.UpdateUserPassword(ctx, sqlcdb.UpdateUserPasswordParams{
		ID:       id,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

func (r *pgRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	exists, err := r.q.UsernameExists(ctx, username)
	if err != nil {
		return false, fmt.Errorf("check username exists: %w", err)
	}
	return exists, nil
}

func (r *pgRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	exists, err := r.q.EmailExists(ctx, pgtype.Text{String: email, Valid: email != ""})
	if err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}
	return exists, nil
}

func (r *pgRepository) CreateRefreshToken(ctx context.Context, params sqlcdb.CreateRefreshTokenParams) (sqlcdb.RefreshToken, error) {
	token, err := r.q.CreateRefreshToken(ctx, params)
	if err != nil {
		return sqlcdb.RefreshToken{}, fmt.Errorf("create refresh token: %w", err)
	}
	return token, nil
}

func (r *pgRepository) GetRefreshTokenByHash(ctx context.Context, hash string) (sqlcdb.RefreshToken, error) {
	token, err := r.q.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return sqlcdb.RefreshToken{}, fmt.Errorf("get refresh token: %w", err)
	}
	return token, nil
}

func (r *pgRepository) DeleteRefreshToken(ctx context.Context, hash string) error {
	err := r.q.DeleteRefreshToken(ctx, hash)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

func (r *pgRepository) DeleteRefreshTokensByUser(ctx context.Context, userID int64) error {
	err := r.q.DeleteRefreshTokensByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("delete user refresh tokens: %w", err)
	}
	return nil
}
