package middleware

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/chiragguptadtu/trigger/internal/store"
)

// RequireCommandAccess allows the request through if the authenticated user is
// an admin OR has an explicit permission grant on the command identified by the
// {slug} path value.
func RequireCommandAccess(q *store.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "not authenticated")
				return
			}

			// Admins bypass all command-level permissions.
			if claims.IsAdmin {
				next.ServeHTTP(w, r)
				return
			}

			slug := r.PathValue("slug")
			cmd, err := q.GetCommandBySlug(r.Context(), slug)
			if err != nil {
				if errors.Is(store.Normalize(err), store.ErrNotFound) {
					writeError(w, http.StatusNotFound, "command not found")
					return
				}
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid user id in token")
				return
			}

			ok, err = q.UserCanAccessCommand(r.Context(), store.UserCanAccessCommandParams{
				UserID:    userID,
				CommandID: cmd.ID,
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			if !ok {
				writeError(w, http.StatusForbidden, "access denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
