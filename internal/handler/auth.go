package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/chiragguptadtu/trigger/internal/auth"
	"github.com/chiragguptadtu/trigger/internal/store"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := s.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(store.Normalize(err), store.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !user.IsActive {
		writeError(w, http.StatusUnauthorized, "account disabled")
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	ttl := time.Duration(s.config.TokenTTL) * time.Second
	token, err := auth.GenerateToken(user.ID.String(), user.IsAdmin, s.config.JWTSecret, ttl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
