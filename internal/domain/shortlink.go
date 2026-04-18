package domain

import "time"

type ShortLink struct {
	ID        int64
	Code      string
	TargetURL string
	UserID    *int64
	Clicks    int64
	ExpiresAt *time.Time
	CreatedAt time.Time
}

func (s *ShortLink) IsExpired(now time.Time) bool {
	return s.ExpiresAt != nil && now.After(*s.ExpiresAt)
}
