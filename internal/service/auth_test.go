package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Ozdal97/go-url-shortener/internal/domain"
	"github.com/Ozdal97/go-url-shortener/internal/pkg/jwt"
)

type memUsers struct {
	byEmail map[string]*domain.User
	nextID  int64
}

func newMemUsers() *memUsers {
	return &memUsers{byEmail: map[string]*domain.User{}}
}

func (m *memUsers) Create(_ context.Context, u *domain.User) error {
	if _, ok := m.byEmail[u.Email]; ok {
		return domain.ErrAlreadyExists
	}
	m.nextID++
	u.ID = m.nextID
	u.CreatedAt = time.Now()
	cp := *u
	m.byEmail[u.Email] = &cp
	return nil
}

func (m *memUsers) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func newSvc(t *testing.T) (*AuthService, *memUsers) {
	t.Helper()
	users := newMemUsers()
	j := jwt.New("supersecret-1234567890", 15*time.Minute, time.Hour)
	return NewAuthService(users, j), users
}

func TestRegister_RejectsShortPassword(t *testing.T) {
	svc, _ := newSvc(t)
	_, err := svc.Register(context.Background(), "a@b.com", "short")
	require.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestRegisterLogin_HappyPath(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()

	u, err := svc.Register(ctx, "Foo@Bar.com", "verysecret")
	require.NoError(t, err)
	require.Equal(t, "foo@bar.com", u.Email)

	tok, err := svc.Login(ctx, "foo@bar.com", "verysecret")
	require.NoError(t, err)
	require.NotEmpty(t, tok.AccessToken)
	require.NotEmpty(t, tok.RefreshToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	_, err := svc.Register(ctx, "x@y.com", "verysecret")
	require.NoError(t, err)

	_, err = svc.Login(ctx, "x@y.com", "wrongpass")
	require.ErrorIs(t, err, domain.ErrInvalidCredential)
}

func TestRefresh_RejectsAccessToken(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	_, err := svc.Register(ctx, "z@z.com", "verysecret")
	require.NoError(t, err)
	pair, err := svc.Login(ctx, "z@z.com", "verysecret")
	require.NoError(t, err)

	_, err = svc.Refresh(ctx, pair.AccessToken)
	require.ErrorIs(t, err, domain.ErrUnauthorized)

	pair2, err := svc.Refresh(ctx, pair.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, pair2.AccessToken)
}
