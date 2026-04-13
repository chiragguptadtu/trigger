package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"trigger/internal/crypto"
	"trigger/internal/middleware"
	"trigger/internal/store"
)

type updateConfigRequest struct {
	Value       string `json:"value"       validate:"required"`
	Description string `json:"description"`
}

type createConfigRequest struct {
	updateConfigRequest
	Key string `json:"key" validate:"required"`
}

func (s *Server) handleCreateConfig(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	var req createConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Key == "" || req.Value == "" {
		writeError(w, http.StatusBadRequest, "key and value are required")
		return
	}

	encrypted, err := crypto.Encrypt(s.config.EncryptionKey, req.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encryption failed")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user id")
		return
	}

	entry, err := s.store.CreateConfigEntry(r.Context(), store.CreateConfigEntryParams{
		Key:            req.Key,
		ValueEncrypted: encrypted,
		Description:    req.Description,
		CreatedBy:      userID,
	})
	if err != nil {
		if errors.Is(store.Normalize(err), store.ErrConflict) {
			writeError(w, http.StatusConflict, "key already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, newConfigResponse(entry))
}

func (s *Server) handleListConfig(w http.ResponseWriter, r *http.Request) {
	rows, err := s.store.ListConfigEntries(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := make([]ConfigResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, newConfigResponseFromRow(row))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req updateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Value == "" {
		writeError(w, http.StatusBadRequest, "value is required")
		return
	}

	encrypted, err := crypto.Encrypt(s.config.EncryptionKey, req.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encryption failed")
		return
	}

	entry, err := s.store.UpdateConfigEntry(r.Context(), store.UpdateConfigEntryParams{
		Key:            key,
		ValueEncrypted: encrypted,
		Description:    req.Description,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, newConfigResponse(entry))
}

func (s *Server) handleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if err := s.store.DeleteConfigEntry(r.Context(), key); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
