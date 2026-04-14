package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/chiragguptadtu/trigger/internal/middleware"
	"github.com/chiragguptadtu/trigger/internal/store"
)

// handleListCommands returns all commands the caller can access.
// Admins receive every active command; operators receive only granted commands.
func (s *Server) handleListCommands(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var cmds []store.Command
	var err error
	if claims.IsAdmin {
		cmds, err = s.store.ListAllCommands(r.Context())
	} else {
		uid, parseErr := uuid.Parse(claims.UserID)
		if parseErr != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		cmds, err = s.store.ListCommandsForUser(r.Context(), uid)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := make([]CommandResponse, len(cmds))
	for i, c := range cmds {
		resp[i] = newCommandResponse(c, nil)
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleGetCommand returns a single command with its inputs.
// Access is enforced by RequireCommandAccess middleware.
func (s *Server) handleGetCommand(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	cmd, err := s.store.GetCommandBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(store.Normalize(err), store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "command not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	inputs, err := s.store.ListCommandInputs(r.Context(), cmd.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, newCommandResponse(cmd, inputs))
}
