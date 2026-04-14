package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/chiragguptadtu/trigger/internal/store"
)

func (s *Server) handleListImportErrors(w http.ResponseWriter, r *http.Request) {
	errs, err := s.store.ListImportErrors(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	result := make([]ImportErrorResponse, 0, len(errs))
	for _, e := range errs {
		result = append(result, newImportErrorResponse(e))
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleListPermissions(w http.ResponseWriter, r *http.Request) {
	cmd, err := s.store.GetCommandBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		if errors.Is(store.Normalize(err), store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "command not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	perms, err := s.store.ListCommandPermissions(r.Context(), cmd.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	result := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		result = append(result, newPermissionResponse(p))
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGrantPermission(w http.ResponseWriter, r *http.Request) {
	cmd, err := s.store.GetCommandBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		if errors.Is(store.Normalize(err), store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "command not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var req struct {
		GranteeType string `json:"grantee_type"`
		GranteeID   string `json:"grantee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.GranteeType != "user" && req.GranteeType != "group" {
		writeError(w, http.StatusBadRequest, "grantee_type must be 'user' or 'group'")
		return
	}

	granteeID, err := uuid.Parse(req.GranteeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid grantee_id")
		return
	}

	perm, err := s.store.CreateCommandPermission(r.Context(), store.CreateCommandPermissionParams{
		CommandID:   cmd.ID,
		GranteeType: req.GranteeType,
		GranteeID:   granteeID,
	})
	if err != nil {
		if errors.Is(store.Normalize(err), store.ErrConflict) {
			writeError(w, http.StatusConflict, "permission already granted")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, newPermissionResponse(perm))
}

func (s *Server) handleRevokePermission(w http.ResponseWriter, r *http.Request) {
	cmd, err := s.store.GetCommandBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		if errors.Is(store.Normalize(err), store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "command not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	granteeType := r.PathValue("granteeType")
	if granteeType != "user" && granteeType != "group" {
		writeError(w, http.StatusBadRequest, "granteeType must be 'user' or 'group'")
		return
	}

	granteeID, err := uuid.Parse(r.PathValue("granteeID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid granteeID")
		return
	}

	if err := s.store.DeleteCommandPermission(r.Context(), store.DeleteCommandPermissionParams{
		CommandID:   cmd.ID,
		GranteeType: granteeType,
		GranteeID:   granteeID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

