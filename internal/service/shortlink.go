package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/Ozdal97/go-url-shortener/internal/domain"
	"github.com/Ozdal97/go-url-shortener/internal/pkg/hashids"
)

type LinkStore interface {
	Insert(ctx context.Context, l *domain.ShortLink) error
	FindByCode(ctx context.Context, code string) (*domain.ShortLink, error)
	IncrementClicks(ctx context.Context, code string) error
	ListByUser(ctx context.Context, userID int64, limit, offset int) ([]domain.ShortLink, error)
	DeleteByCode(ctx context.Context, userID int64, code string) error
}

type Cache interface {
	Get(ctx context.Context, code string) (string, error)
	Set(ctx context.Context, code, target string) error
	Del(ctx context.Context, code string) error
}

type LinkService struct {
	repo  LinkStore
	cache Cache
	enc   *hashids.Encoder
	now   func() time.Time
}

func NewLinkService(repo LinkStore, cache Cache, enc *hashids.Encoder) *LinkService {
	return &LinkService{repo: repo, cache: cache, enc: enc, now: time.Now}
}

type CreateInput struct {
	URL       string
	UserID    *int64
	ExpiresAt *time.Time
}

func (s *LinkService) Create(ctx context.Context, in CreateInput) (*domain.ShortLink, error) {
	if err := validateURL(in.URL); err != nil {
		return nil, err
	}
	const maxAttempts = 5
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		code, err := s.enc.Random()
		if err != nil {
			return nil, err
		}
		l := &domain.ShortLink{
			Code:      code,
			TargetURL: in.URL,
			UserID:    in.UserID,
			ExpiresAt: in.ExpiresAt,
		}
		if err := s.repo.Insert(ctx, l); err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				lastErr = err
				continue
			}
			return nil, err
		}
		return l, nil
	}
	return nil, lastErr
}

func (s *LinkService) Resolve(ctx context.Context, code string) (string, error) {
	if v, err := s.cache.Get(ctx, code); err == nil {
		go func() {
			bg, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = s.repo.IncrementClicks(bg, code)
		}()
		return v, nil
	}

	l, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return "", err
	}
	if l.IsExpired(s.now()) {
		return "", domain.ErrExpired
	}
	_ = s.cache.Set(ctx, code, l.TargetURL)
	go func() {
		bg, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.repo.IncrementClicks(bg, code)
	}()
	return l.TargetURL, nil
}

func (s *LinkService) ListMine(ctx context.Context, userID int64, limit, offset int) ([]domain.ShortLink, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *LinkService) Delete(ctx context.Context, userID int64, code string) error {
	if err := s.repo.DeleteByCode(ctx, userID, code); err != nil {
		return err
	}
	_ = s.cache.Del(ctx, code)
	return nil
}

func validateURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return domain.ErrInvalidInput
	}
	u, err := url.Parse(raw)
	if err != nil {
		return domain.ErrInvalidInput
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return domain.ErrInvalidInput
	}
	if u.Host == "" {
		return domain.ErrInvalidInput
	}
	return nil
}
