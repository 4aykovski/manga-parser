package parser

import "errors"

var (
	ErrChaptersNotFound    = errors.New("chapters not found")
	ErrTitleNameNotFound   = errors.New("title name not found")
	ErrDescriptionNotFound = errors.New("description not found")
)
