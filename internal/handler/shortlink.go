package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Ozdal97/go-url-shortener/internal/domain"
	"github.com/Ozdal97/go-url-shortener/internal/service"
)

type LinkHandler struct {
	svc *service.LinkService
}

func NewLinkHandler(svc *service.LinkService) *LinkHandler {
	return &LinkHandler{svc: svc}
}

type createReq struct {
	URL       string     `json:"url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type linkResp struct {
	Code      string     `json:"code"`
	URL       string     `json:"url"`
	ShortURL  string     `json:"short_url"`
	Clicks    int64      `json:"clicks"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func toResp(l *domain.ShortLink, base string) linkResp {
	return linkResp{
		Code:      l.Code,
		URL:       l.TargetURL,
		ShortURL:  base + "/" + l.Code,
		Clicks:    l.Clicks,
		ExpiresAt: l.ExpiresAt,
		CreatedAt: l.CreatedAt,
	}
}

func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	uid, _ := userIDFrom(r)
	in := service.CreateInput{URL: req.URL, UserID: &uid, ExpiresAt: req.ExpiresAt}
	l, err := h.svc.Create(r.Context(), in)
	if err != nil {
		s, m := mapError(err)
		writeError(w, s, m)
		return
	}
	writeJSON(w, http.StatusCreated, toResp(l, baseURL(r)))
}

func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	target, err := h.svc.Resolve(r.Context(), code)
	if err != nil {
		s, m := mapError(err)
		writeError(w, s, m)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *LinkHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFrom(r)
	limit := atoiQuery(r, "limit", 20)
	offset := atoiQuery(r, "offset", 0)
	links, err := h.svc.ListMine(r.Context(), uid, limit, offset)
	if err != nil {
		s, m := mapError(err)
		writeError(w, s, m)
		return
	}
	out := make([]linkResp, 0, len(links))
	base := baseURL(r)
	for i := range links {
		out = append(out, toResp(&links[i], base))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *LinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFrom(r)
	code := chi.URLParam(r, "code")
	if err := h.svc.Delete(r.Context(), uid, code); err != nil {
		s, m := mapError(err)
		writeError(w, s, m)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if v := r.Header.Get("X-Forwarded-Proto"); v != "" {
		scheme = v
	}
	return scheme + "://" + r.Host
}

func atoiQuery(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	var n int
	for _, c := range v {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}
