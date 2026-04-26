package handler

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"

	"github.com/Ozdal97/go-url-shortener/internal/pkg/jwt"
)

type ctxKey string

const userCtxKey ctxKey = "uid"

func AuthMiddleware(j *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			tok := strings.TrimPrefix(h, "Bearer ")
			claims, err := j.Parse(tok)
			if err != nil || claims.Type != "access" {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func userIDFrom(r *http.Request) (int64, bool) {
	v, ok := r.Context().Value(userCtxKey).(int64)
	return v, ok
}

type ipLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

func NewIPLimiter(rps, burst int) *ipLimiter {
	return &ipLimiter{
		visitors: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if v, ok := l.visitors[ip]; ok {
		return v
	}
	lim := rate.NewLimiter(l.rps, l.burst)
	l.visitors[ip] = lim
	return lim
}

func (l *ipLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !l.get(ip).Allow() {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.Index(v, ","); i > 0 {
			return strings.TrimSpace(v[:i])
		}
		return strings.TrimSpace(v)
	}
	return r.RemoteAddr
}
