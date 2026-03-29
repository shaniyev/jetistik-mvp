package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"jetistik/internal/sqlcdb"
)

var (
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrUserNotActive       = errors.New("user account is not active")
	ErrUsernameExists      = errors.New("username already taken")
	ErrEmailExists         = errors.New("email already taken")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

// Service handles auth business logic.
type Service struct {
	repo       Repository
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewService creates a new auth service.
func NewService(repo Repository, secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Login authenticates a user and returns a token pair.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, string, error) {
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("login: %w", err)
	}

	if !user.IsActive.Bool {
		return nil, "", ErrUserNotActive
	}

	matches, needsRehash := VerifyPassword(req.Password, user.Password)
	if !matches {
		return nil, "", ErrInvalidCredentials
	}

	// Rehash Django PBKDF2 to bcrypt on successful login
	if needsRehash {
		newHash, err := HashPassword(req.Password)
		if err != nil {
			slog.Error("failed to rehash password", "user_id", user.ID, "error", err)
		} else {
			if err := s.repo.UpdateUserPassword(ctx, user.ID, newHash); err != nil {
				slog.Error("failed to update rehashed password", "user_id", user.ID, "error", err)
			} else {
				slog.Info("rehashed Django password to bcrypt", "user_id", user.ID)
			}
		}
	}

	return s.issueTokens(ctx, user)
}

// Register creates a new user and returns a token pair.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, string, error) {
	exists, err := s.repo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, "", fmt.Errorf("register: %w", err)
	}
	if exists {
		return nil, "", ErrUsernameExists
	}

	if req.Email != "" {
		emailExists, err := s.repo.EmailExists(ctx, req.Email)
		if err != nil {
			return nil, "", fmt.Errorf("register: %w", err)
		}
		if emailExists {
			return nil, "", ErrEmailExists
		}
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, "", fmt.Errorf("register: %w", err)
	}

	lang := req.Language
	if lang == "" {
		lang = "kz"
	}

	user, err := s.repo.CreateUser(ctx, sqlcdb.CreateUserParams{
		Username: req.Username,
		Email:    pgtype.Text{String: req.Email, Valid: req.Email != ""},
		Password: hashedPassword,
		Iin:      pgtype.Text{String: req.IIN, Valid: req.IIN != ""},
		Role:     req.Role,
		Language: pgtype.Text{String: lang, Valid: true},
	})
	if err != nil {
		return nil, "", fmt.Errorf("register: %w", err)
	}

	return s.issueTokens(ctx, user)
}

// RegisterOrg creates a new staff user and organization.
func (s *Service) RegisterOrg(ctx context.Context, req RegisterOrgRequest) (*AuthResponse, string, error) {
	exists, err := s.repo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, "", fmt.Errorf("register org: %w", err)
	}
	if exists {
		return nil, "", ErrUsernameExists
	}

	if req.Email != "" {
		emailExists, err := s.repo.EmailExists(ctx, req.Email)
		if err != nil {
			return nil, "", fmt.Errorf("register org: %w", err)
		}
		if emailExists {
			return nil, "", ErrEmailExists
		}
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, "", fmt.Errorf("register org: %w", err)
	}

	lang := req.Language
	if lang == "" {
		lang = "kz"
	}

	// Create staff user
	user, err := s.repo.CreateUser(ctx, sqlcdb.CreateUserParams{
		Username: req.Username,
		Email:    pgtype.Text{String: req.Email, Valid: req.Email != ""},
		Password: hashedPassword,
		Role:     "staff",
		Language: pgtype.Text{String: lang, Valid: true},
	})
	if err != nil {
		return nil, "", fmt.Errorf("register org: %w", err)
	}

	// Create organization
	org, err := s.repo.CreateOrganization(ctx, sqlcdb.CreateOrganizationParams{
		Name: req.OrgName,
	})
	if err != nil {
		return nil, "", fmt.Errorf("register org: create org: %w", err)
	}

	// Link user to organization as owner
	if err := s.repo.AddOrganizationMember(ctx, sqlcdb.AddOrganizationMemberParams{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           pgtype.Text{String: "owner", Valid: true},
	}); err != nil {
		return nil, "", fmt.Errorf("register org: add member: %w", err)
	}

	return s.issueTokens(ctx, user)
}

// Refresh validates a refresh token and issues a new token pair.
func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*AuthResponse, string, error) {
	hash := HashRefreshToken(rawRefreshToken)

	storedToken, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", ErrInvalidRefreshToken
		}
		return nil, "", fmt.Errorf("refresh: %w", err)
	}

	// Delete the used refresh token (rotation)
	if err := s.repo.DeleteRefreshToken(ctx, hash); err != nil {
		slog.Error("failed to delete used refresh token", "error", err)
	}

	user, err := s.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, "", fmt.Errorf("refresh: %w", err)
	}

	if !user.IsActive.Bool {
		return nil, "", ErrUserNotActive
	}

	return s.issueTokens(ctx, user)
}

// Logout deletes the refresh token.
func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}
	hash := HashRefreshToken(rawRefreshToken)
	return s.repo.DeleteRefreshToken(ctx, hash)
}

// issueTokens creates a new access + refresh token pair for a user.
func (s *Service) issueTokens(ctx context.Context, user sqlcdb.User) (*AuthResponse, string, error) {
	accessToken, err := GenerateAccessToken(user.ID, user.Username, user.Role, s.secret, s.accessTTL)
	if err != nil {
		return nil, "", fmt.Errorf("issue tokens: %w", err)
	}

	rawRefresh, err := GenerateRefreshToken()
	if err != nil {
		return nil, "", fmt.Errorf("issue tokens: %w", err)
	}

	refreshHash := HashRefreshToken(rawRefresh)
	_, err = s.repo.CreateRefreshToken(ctx, sqlcdb.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(s.refreshTTL), Valid: true},
	})
	if err != nil {
		return nil, "", fmt.Errorf("issue tokens: %w", err)
	}

	resp := &AuthResponse{
		AccessToken: accessToken,
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email.String,
			IIN:       maskIIN(user.Iin.String),
			Role:      user.Role,
			Language:  user.Language.String,
			CreatedAt: user.CreatedAt.Time,
		},
	}

	return resp, rawRefresh, nil
}

// maskIIN masks an IIN for display: 990512345678 -> 9905****5678
func maskIIN(iin string) string {
	if len(iin) != 12 {
		return iin
	}
	return iin[:4] + "****" + iin[8:]
}
