package parser

import "time"

type Chapter struct {
	ProjectName string    `json:"project_name,omitempty"`
	Name        string    `json:"name,omitempty"`
	Url         string    `json:"url,omitempty"`
	Number      int       `json:"number,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
}
