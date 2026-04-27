package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ozdal97/go-url-shortener/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type credReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req credReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	u, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		s, m := mapError(err)
		writeError(w, s, m)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": u.ID, "email": u.Email})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req credReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	pair, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		s, m := mapError(err)
		writeError(w, s, m)
		return
	}
	writeJSON(w, http.StatusOK, tokenResp{pair.AccessToken, pair.RefreshToken})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	pair, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		s, m := mapError(err)
		writeError(w, s, m)
		return
	}
	writeJSON(w, http.StatusOK, tokenResp{pair.AccessToken, pair.RefreshToken})
}
