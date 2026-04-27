package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Ozdal97/go-url-shortener/internal/pkg/jwt"
)

type Deps struct {
	Auth    *AuthHandler
	Link    *LinkHandler
	JWT     *jwt.Manager
	Limiter *ipLimiter
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * 1000_000_000))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "PATCH"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))
	r.Use(d.Limiter.Middleware)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/register", d.Auth.Register)
		api.Post("/auth/login", d.Auth.Login)
		api.Post("/auth/refresh", d.Auth.Refresh)

		api.Group(func(p chi.Router) {
			p.Use(AuthMiddleware(d.JWT))
			p.Post("/links", d.Link.Create)
			p.Get("/links", d.Link.List)
			p.Delete("/links/{code}", d.Link.Delete)
		})
	})

	r.Get("/{code}", d.Link.Redirect)
	return r
}
