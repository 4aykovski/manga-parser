package parser

import "time"

type Project struct {
	Name          string    `json:"name"`
	Url           string    `json:"url"`
	Tags          []string  `json:"tags,omitempty"`
	ChaptersCount int       `json:"chapters_count"`
	Chapters      []Chapter `json:"chapters,omitempty"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	Description   string    `json:"description,omitempty"`
	Authors       []string  `json:"authors,omitempty"`
}
