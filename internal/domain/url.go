package domain

import (
	"time"
)

type URL struct {
	ID          int64                  `json:"id"`
	ShortCode   string                 `json:"short_code"`
	OriginalURL string                 `json:"original_url"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	ClickCount  int64                  `json:"click_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type Analytics struct {
	ID        int64     `json:"id"`
	ShortCode string    `json:"short_code"`
	ClickedAt time.Time `json:"clicked_at"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Referer   string    `json:"referer,omitempty"`
	Country   string    `json:"country,omitempty"`
}

type URLStats struct {
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	ClickCount  int64      `json:"click_count"`
	CreatedAt   time.Time  `json:"created_at"`
	LastClicked *time.Time `json:"last_clicked,omitempty"`
}
