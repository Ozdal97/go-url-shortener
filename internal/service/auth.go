package service

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/Ozdal97/go-url-shortener/internal/domain"
	"github.com/Ozdal97/go-url-shortener/internal/pkg/jwt"
)

type UserStore interface {
	Create(ctx context.Context, u *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type AuthService struct {
	users UserStore
	jwt   *jwt.Manager
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func NewAuthService(users UserStore, j *jwt.Manager) *AuthService {
	return &AuthService{users: users, jwt: j}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || len(password) < 8 {
		return nil, domain.ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &domain.User{Email: email, Password: string(hash)}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	u.Password = ""
	return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrInvalidCredential
	}
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredential
	}
	return s.issuePair(u.ID)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.jwt.Parse(refreshToken)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if claims.Type != "refresh" {
		return nil, domain.ErrUnauthorized
	}
	return s.issuePair(claims.UserID)
}

func (s *AuthService) issuePair(userID int64) (*TokenPair, error) {
	at, err := s.jwt.IssueAccess(userID)
	if err != nil {
		return nil, err
	}
	rt, err := s.jwt.IssueRefresh(userID)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: at, RefreshToken: rt}, nil
}
