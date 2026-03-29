package middleware

import (
	"context"
	"net/http"
	"strings"

	"jetistik/internal/auth"
	"jetistik/internal/platform/response"
)

type userContextKey struct{}

// UserClaims holds the authenticated user's JWT claims in the request context.
type UserClaims struct {
	UserID   int64
	Username string
	Role     string
}

// JWTAuth creates middleware that extracts and validates JWT access tokens.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				response.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization header is required")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "authorization header must be Bearer <token>")
				return
			}

			claims, err := auth.ParseAccessToken(parts[1], secret)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired access token")
				return
			}

			uc := UserClaims{
				UserID:   claims.UserID,
				Username: claims.Username,
				Role:     claims.Role,
			}

			ctx := context.WithValue(r.Context(), userContextKey{}, uc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUser extracts the authenticated user claims from the request context.
func GetUser(ctx context.Context) (UserClaims, bool) {
	uc, ok := ctx.Value(userContextKey{}).(UserClaims)
	return uc, ok
}

// MustGetUser extracts user claims or panics. Use only after JWTAuth middleware.
func MustGetUser(ctx context.Context) UserClaims {
	uc, ok := GetUser(ctx)
	if !ok {
		panic("middleware: MustGetUser called without JWTAuth middleware")
	}
	return uc
}
