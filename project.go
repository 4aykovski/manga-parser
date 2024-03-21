package parser

import "time"

type Project struct {
	Name          string    `json:"name,omitempty"`
	Url           string    `json:"url,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
	ChaptersCount int       `json:"chapters_count,omitempty"`
	Chapters      []Chapter `json:"chapters,omitempty"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}
