package middleware

import (
	"net/http"

	"jetistik/internal/platform/response"
)

// RequireRole creates middleware that checks whether the authenticated user
// has one of the allowed roles. Must be used after JWTAuth middleware.
func RequireRole(allowed ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		roleSet[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uc, ok := GetUser(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authentication required")
				return
			}

			if !roleSet[uc.Role] {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
