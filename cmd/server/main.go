package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Ozdal97/go-url-shortener/internal/config"
	"github.com/Ozdal97/go-url-shortener/internal/handler"
	"github.com/Ozdal97/go-url-shortener/internal/pkg/hashids"
	jwtpkg "github.com/Ozdal97/go-url-shortener/internal/pkg/jwt"
	"github.com/Ozdal97/go-url-shortener/internal/repository"
	"github.com/Ozdal97/go-url-shortener/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("config")
	}
	setupLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := repository.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("postgres")
	}
	defer pool.Close()

	rdbOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("redis url")
	}
	rdb := redis.NewClient(rdbOpts)
	defer rdb.Close()

	enc, err := hashids.New(cfg.HashIDSalt, cfg.HashIDMinLen)
	if err != nil {
		log.Fatal().Err(err).Msg("hashids")
	}

	jm := jwtpkg.New(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	users := repository.NewUserRepo(pool)
	links := repository.NewShortLinkRepo(pool)
	cache := repository.NewLinkCache(rdb, time.Hour)

	authSvc := service.NewAuthService(users, jm)
	linkSvc := service.NewLinkService(links, cache, enc)

	deps := handler.Deps{
		Auth:    handler.NewAuthHandler(authSvc),
		Link:    handler.NewLinkHandler(linkSvc),
		JWT:     jm,
		Limiter: handler.NewIPLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst),
	}

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler.NewRouter(deps),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("http server starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("listen")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
}

func setupLogger(level string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
}
