package models

import "time"

type Url struct {
	Id             int64      `json:"id"`
	Username       string     `json:"username"`
	Short_code     string     `json:"short_code"`
	Original_url   string     `json:"original_url"`
	CreatedAt      time.Time  `json:"created_at"`
	ClickCount     int64      `json:"click_count"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
}

type StatsResponse struct {
	ShortCode      string     `json:"short_code"`
	OriginalURL    string     `json:"original_url"`
	CreatedAt      time.Time  `json:"created_at"`
	ClickCount     int64      `json:"click_count"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
}
