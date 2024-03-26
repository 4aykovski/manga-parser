package parser

import "errors"

var (
	ErrChaptersNotFound        = errors.New("chapters not found")
	ErrProjectNameNotFound     = errors.New("title name not found")
	ErrDescriptionNotFound     = errors.New("description not found")
	ErrCantParseLastUpdateDate = errors.New("can't parse last update date")
)
