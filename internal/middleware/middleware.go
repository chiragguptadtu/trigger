package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/chiragguptadtu/trigger/internal/auth"
)

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type contextKey string

const claimsKey contextKey = "auth_claims"

// WithClaims stores claims in the context (used by tests and Authenticate).
func WithClaims(ctx context.Context, c *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

// ClaimsFromContext retrieves claims stored by Authenticate.
func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*auth.Claims)
	return c, ok && c != nil
}

// Authenticate validates the Bearer JWT and injects claims into the context.
// Returns 401 if the header is missing, malformed, or the token is invalid.
// Short-circuits if claims are already present in the context (test seam — safe
// because the contextKey type is unexported and cannot be set outside this package).
func Authenticate(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := ClaimsFromContext(r.Context()); ok {
				next.ServeHTTP(w, r)
				return
			}
			hdr := r.Header.Get("Authorization")
			if !strings.HasPrefix(hdr, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "missing or malformed authorization header")
				return
			}
			token := strings.TrimPrefix(hdr, "Bearer ")
			claims, err := auth.ValidateToken(token, secret)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

// RequireAdmin allows only admin users through. Must be used after Authenticate.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		if !claims.IsAdmin {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
