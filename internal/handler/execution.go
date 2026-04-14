package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/chiragguptadtu/trigger/internal/middleware"
	"github.com/chiragguptadtu/trigger/internal/store"
)

type triggerRequest struct {
	Inputs map[string]any `json:"inputs"`
}

func (s *Server) handleTriggerExecution(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

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

	var req triggerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Inputs == nil {
		req.Inputs = map[string]any{}
	}

	// Validate inputs against the command's schema.
	inputs, err := s.store.ListCommandInputs(r.Context(), cmd.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if msg := validateInputs(inputs, req.Inputs); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	inputsJSON, err := json.Marshal(req.Inputs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user id")
		return
	}

	exec_, err := s.store.CreateExecution(r.Context(), store.CreateExecutionParams{
		CommandID:   cmd.ID,
		TriggeredBy: userID,
		Inputs:      inputsJSON,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if s.enqueuer != nil {
		if err := s.enqueuer.Enqueue(r.Context(), exec_.ID.String()); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to enqueue job")
			return
		}
	}

	writeJSON(w, http.StatusAccepted, newExecutionResponse(exec_))
}

func (s *Server) handleGetExecution(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid execution id")
		return
	}

	exec_, err := s.store.GetExecutionByID(r.Context(), id)
	if err != nil {
		if errors.Is(store.Normalize(err), store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "execution not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, newExecutionResponse(exec_))
}

func (s *Server) handleListExecutions(w http.ResponseWriter, r *http.Request) {
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

	executions, err := s.store.ListExecutionsForCommand(r.Context(), store.ListExecutionsForCommandParams{
		CommandID: cmd.ID,
		Limit:     50,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	result := make([]ExecutionResponse, 0, len(executions))
	for _, e := range executions {
		result = append(result, newExecutionListResponse(e))
	}
	writeJSON(w, http.StatusOK, result)
}

// validateInputs checks required presence and closed-input option membership.
// Returns an error message or "" if valid.
func validateInputs(schema []store.CommandInput, provided map[string]any) string {
	for _, input := range schema {
		val, exists := provided[input.Name]

		if input.Required && !exists {
			return fmt.Sprintf("missing required input: %s", input.Name)
		}
		if !exists {
			continue
		}

		if input.Type == "closed" && input.Options != nil {
			var options []string
			if err := json.Unmarshal(input.Options, &options); err != nil {
				continue
			}
			if input.Multi {
				arr, ok := val.([]any)
				if !ok {
					return fmt.Sprintf("input %s must be an array", input.Name)
				}
				for _, item := range arr {
					strVal, ok := item.(string)
					if !ok {
						return fmt.Sprintf("input %s: all values must be strings", input.Name)
					}
					if !containsStr(options, strVal) {
						return fmt.Sprintf("invalid value for %s: %q is not one of %v", input.Name, strVal, options)
					}
				}
			} else {
				strVal, ok := val.(string)
				if !ok {
					return fmt.Sprintf("input %s must be a string", input.Name)
				}
				if !containsStr(options, strVal) {
					return fmt.Sprintf("invalid value for %s: must be one of %v", input.Name, options)
				}
			}
		}
	}
	return ""
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
