package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ozdal97/go-url-shortener/internal/domain"
)

type ShortLinkRepo struct {
	db *pgxpool.Pool
}

func NewShortLinkRepo(db *pgxpool.Pool) *ShortLinkRepo {
	return &ShortLinkRepo{db: db}
}

func (r *ShortLinkRepo) Insert(ctx context.Context, l *domain.ShortLink) error {
	const q = `
		INSERT INTO short_links (code, target_url, user_id, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	err := r.db.QueryRow(ctx, q, l.Code, l.TargetURL, l.UserID, l.ExpiresAt).
		Scan(&l.ID, &l.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *ShortLinkRepo) FindByCode(ctx context.Context, code string) (*domain.ShortLink, error) {
	const q = `
		SELECT id, code, target_url, user_id, clicks, expires_at, created_at
		FROM short_links WHERE code = $1`
	l := &domain.ShortLink{}
	err := r.db.QueryRow(ctx, q, code).
		Scan(&l.ID, &l.Code, &l.TargetURL, &l.UserID, &l.Clicks, &l.ExpiresAt, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *ShortLinkRepo) IncrementClicks(ctx context.Context, code string) error {
	const q = `UPDATE short_links SET clicks = clicks + 1 WHERE code = $1`
	_, err := r.db.Exec(ctx, q, code)
	return err
}

func (r *ShortLinkRepo) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]domain.ShortLink, error) {
	const q = `
		SELECT id, code, target_url, user_id, clicks, expires_at, created_at
		FROM short_links WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.ShortLink
	for rows.Next() {
		var l domain.ShortLink
		if err := rows.Scan(&l.ID, &l.Code, &l.TargetURL, &l.UserID, &l.Clicks, &l.ExpiresAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *ShortLinkRepo) DeleteByCode(ctx context.Context, userID int64, code string) error {
	const q = `DELETE FROM short_links WHERE code = $1 AND user_id = $2`
	tag, err := r.db.Exec(ctx, q, code, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
