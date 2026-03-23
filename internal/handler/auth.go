package handler

import (
	"encoding/json"
	"net/http"
)

type dummyLoginRequest struct {
	Role string `json:"role"`
}

func (h *Handler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Role != "admin" && req.Role != "user" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "role must be 'admin' or 'user'")
		return
	}

	token, err := h.auth.DummyLogin(req.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
