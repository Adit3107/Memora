package models

import "time"

type ContentType string

const (
	ContentTypeVideo    ContentType = "video"
	ContentTypeDocument ContentType = "document"
	ContentTypeArticle  ContentType = "article"
	ContentTypeImage    ContentType = "image"
)

type Content struct {
	ID           string      `json:"id"`
	UserID       string      `json:"user_id"`
	SpaceID      string      `json:"space_id"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	Type         ContentType `json:"type"`
	SourceURL    *string     `json:"source_url,omitempty"`
	ThumbnailURL *string     `json:"thumbnail_url,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
