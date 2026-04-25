package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Ozdal97/go-url-shortener/internal/domain"
	"github.com/Ozdal97/go-url-shortener/internal/pkg/hashids"
)

type memLinks struct {
	mu    sync.Mutex
	byCode map[string]*domain.ShortLink
}

func newMemLinks() *memLinks { return &memLinks{byCode: map[string]*domain.ShortLink{}} }

func (m *memLinks) Insert(_ context.Context, l *domain.ShortLink) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byCode[l.Code]; ok {
		return domain.ErrAlreadyExists
	}
	cp := *l
	m.byCode[l.Code] = &cp
	return nil
}

func (m *memLinks) FindByCode(_ context.Context, code string) (*domain.ShortLink, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.byCode[code]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *l
	return &cp, nil
}

func (m *memLinks) IncrementClicks(_ context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if l, ok := m.byCode[code]; ok {
		l.Clicks++
	}
	return nil
}

func (m *memLinks) ListByUser(_ context.Context, _ int64, _, _ int) ([]domain.ShortLink, error) {
	return nil, nil
}

func (m *memLinks) DeleteByCode(_ context.Context, _ int64, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byCode[code]; !ok {
		return domain.ErrNotFound
	}
	delete(m.byCode, code)
	return nil
}

type noopCache struct{}

func (noopCache) Get(context.Context, string) (string, error) { return "", domain.ErrNotFound }
func (noopCache) Set(context.Context, string, string) error   { return nil }
func (noopCache) Del(context.Context, string) error           { return nil }

func newLinkSvc(t *testing.T) *LinkService {
	t.Helper()
	enc, err := hashids.New("abc", 6)
	require.NoError(t, err)
	return NewLinkService(newMemLinks(), noopCache{}, enc)
}

func TestCreate_RejectsBadURL(t *testing.T) {
	svc := newLinkSvc(t)
	cases := []string{"", "not-a-url", "ftp://x", "http://"}
	for _, c := range cases {
		_, err := svc.Create(context.Background(), CreateInput{URL: c})
		require.Error(t, err, c)
	}
}

func TestCreate_AndResolve(t *testing.T) {
	svc := newLinkSvc(t)
	ctx := context.Background()
	l, err := svc.Create(ctx, CreateInput{URL: "https://golang.org"})
	require.NoError(t, err)
	require.NotEmpty(t, l.Code)

	got, err := svc.Resolve(ctx, l.Code)
	require.NoError(t, err)
	require.Equal(t, "https://golang.org", got)
}

func TestResolve_Expired(t *testing.T) {
	svc := newLinkSvc(t)
	ctx := context.Background()
	past := time.Now().Add(-time.Hour)
	_, err := svc.Create(ctx, CreateInput{URL: "https://x.com", ExpiresAt: &past})
	require.NoError(t, err)

	// İlk linkin code'unu öğrenmek için yeni bir create + resolve çevriminden
	// kaçınmak adına direkt store üzerinden alalım.
	links := svc.repo.(*memLinks)
	var code string
	for c := range links.byCode {
		code = c
	}
	_, err = svc.Resolve(ctx, code)
	require.True(t, errors.Is(err, domain.ErrExpired))
}
