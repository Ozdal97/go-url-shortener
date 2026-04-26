package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Ozdal97/go-url-shortener/internal/domain"
)

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict, "already exists"
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, "invalid input"
	case errors.Is(err, domain.ErrInvalidCredential):
		return http.StatusUnauthorized, "invalid credentials"
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, domain.ErrExpired):
		return http.StatusGone, "link expired"
	default:
		return http.StatusInternalServerError, "internal error"
	}
}
