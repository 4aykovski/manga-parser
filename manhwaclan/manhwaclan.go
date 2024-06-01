// Package manhwaclan contains Parser for https://manhwaclan.com/ and can be used only with this domain.
package manhwaclan

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"

	"github.com/4aykovski/manga-parser"
)

type Parser struct {
	Collector *colly.Collector
	projects  *[]parser.Project
	errors    chan error
	mutex     *sync.Mutex
}

// New creates new manhwaclan parser.
// Defines Collector.OnRequest, Collector.OnError, Collector.OnHtml for manhwaclan parsing
func New(log *slog.Logger, projectsForOnce int) *Parser {
	projects := make([]parser.Project, 0, projectsForOnce)
	errorsChan := make(chan error, projectsForOnce)
	var mu sync.Mutex

	collector := colly.NewCollector(
		colly.AllowedDomains("manhwaclan.com"),
	)

	collector.OnRequest(func(r *colly.Request) {
		log.Info("Visiting", slog.String("url", r.URL.String()))
	})

	collector.OnError(func(r *colly.Response, err error) {
		errorsChan <- err
	})

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		project, err := collectProjectInfo(e)
		if err != nil {
			errorsChan <- err
			return
		}

		mu.Lock()
		projects = append(projects, project)
		mu.Unlock()
	})

	return &Parser{
		Collector: collector,
		projects:  &projects,
		errors:    errorsChan,
		mutex:     &mu,
	}
}

// Projects returns list of parsed projects
func (p *Parser) Projects() []parser.Project {
	return *p.projects
}

// Close closes errors channel
func (p *Parser) Close() {
	close(p.errors)
}

// Parse parses manhwaclan project by project url
// If any errors was occurred it will be sent to errors channel that is returned by Errors method
func (p *Parser) Parse(projectUrl string) {
	p.Collector.Visit(projectUrl)
}

// Errors returns errors channel with errors that occurred during parsing
func (p *Parser) Errors() <-chan error {
	return p.errors
}

// collectProjectInfo collects project info from HTML
func collectProjectInfo(e *colly.HTMLElement) (parser.Project, error) {
	// get chapters list
	chaptersList := e.DOM.Find("ul.version-chap")
	if chaptersList.Nodes == nil {
		return parser.Project{}, fmt.Errorf("URL - %s: %w", e.Request.URL, parser.ErrChaptersNotFound)
	}

	// get project name from meta
	projectName, ok := e.DOM.Find("meta[property='og:title']").Attr("content")
	if !ok {
		return parser.Project{}, fmt.Errorf("URL - %s: %w", e.Request.URL, parser.ErrProjectNameNotFound)
	}

	// collect chapters info
	chapters, err := collectChaptersInfo(chaptersList, projectName)
	if err != nil {
		return parser.Project{}, fmt.Errorf("URL - %s: %w - %w", e.Request.URL, parser.ErrCantParseChapters, err)
	}

	// get chapters count
	chaptersCount := len(chaptersList.Find("li").Nodes)

	// get project description from meta
	description, ok := e.DOM.Find("meta[property='og:description']").Attr("content")
	if !ok {
		return parser.Project{}, fmt.Errorf("URL - %s: %w", e.Request.URL, parser.ErrDescriptionNotFound)
	}

	// get project tags
	var tags []string
	tagNodes := e.DOM.Find("div.genres-content").Find("-a").Nodes
	for _, tagNode := range tagNodes {
		tags = append(tags, tagNode.FirstChild.Data)
	}

	// get project last updated at
	lastUpdateDate, err := getLastUpdateDate(chaptersList)
	if err != nil {
		return parser.Project{}, fmt.Errorf("URL - %s: %w: %s", e.Request.URL, parser.ErrCantParseLastUpdateDate, err)
	}

	// get authors
	var authors []string
	authorNodes := e.DOM.Find("div.author-content").Find("a").Nodes
	for _, authorNode := range authorNodes {
		authors = append(authors, authorNode.FirstChild.Data)
	}

	return parser.Project{
		Name:          projectName,
		Url:           e.Request.URL.String(),
		Tags:          tags,
		ChaptersCount: chaptersCount,
		Chapters:      chapters,
		LastUpdatedAt: lastUpdateDate,
		Description:   description,
		Authors:       authors,
	}, nil
}

// collectChaptersInfo collects chapters info from HTML chapters list
func collectChaptersInfo(chaptersList *goquery.Selection, projectName string) ([]parser.Chapter, error) {
	chapters := make([]parser.Chapter, 0, chaptersList.Find("li").Length())

	errs := make(chan error, chaptersList.Find("li").Length())

	chaptersList.Find("li").Each(func(i int, s *goquery.Selection) {
		chapterUrl := s.Find("a").AttrOr("href", "")

		chapterReleaseDateSelection := s.
			Find("span.chapter-release-date").
			First()

		chapterUploadedAt, err := getUploadDate(chapterReleaseDateSelection)
		if err != nil {
			errs <- err
			return
		}

		chapterNumber := strings.Split(s.Find("a").Text(), " ")[1]

		chapter := parser.Chapter{
			ProjectName: projectName,
			Name:        "",
			Url:         chapterUrl,
			Number:      chapterNumber,
			UploadedAt:  chapterUploadedAt,
		}

		chapters = append(chapters, chapter)
	})

	close(errs)

	// check if there was any errors and chapter wasn't added
	for err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return chapters, nil
}

// getLastUpdateDate gets last update date from HTML chapters list
func getLastUpdateDate(chaptersList *goquery.Selection) (time.Time, error) {
	lastChapterReleaseDateSelection := chaptersList.
		Find("li").
		First().
		Find("span.chapter-release-date").
		First()

	lastUpdateDate, err := getUploadDate(lastChapterReleaseDateSelection)
	if err != nil {
		return time.Time{}, fmt.Errorf("can't get last update date: %w", err)
	}

	return lastUpdateDate, nil
}

// getUploadDate gets upload date from goquery Selection.
// There are two ways of html on manhwaclan
// 1. markup when last chapter was just released. This markup contains "a" tag with "title" attribute that contains
// information about how long ago chapter was released in format "N hours/days ago"
// 2. standard markup when last chapter was released a long time ago. This markup contains "i" tag with text inside.
// This text contains information about how long ago chapter was released in format "MMM DD, YYYY"
func getUploadDate(chapterSelection *goquery.Selection) (time.Time, error) {
	lastUpdateDate := time.Time{}
	if chapterSelection.Find("a").Nodes != nil {
		linkSelection := chapterSelection.Find("a")
		titleAttr, ok := linkSelection.Attr("title")
		if !ok {
			return time.Time{}, errors.New("can't get 'title' attribute")
		}

		titleFormat := strings.Split(titleAttr, " ")
		if len(titleFormat) > 3 {
			return time.Time{}, errors.New("'title' attr too long")
		}

		switch titleFormat[1] {
		case "hours":
		case "hour":
			hours, err := strconv.Atoi(titleFormat[0])
			if err != nil {
				return time.Time{}, fmt.Errorf("can't convert: %w", err)
			}
			lastUpdateDate = time.Now().Add(-time.Duration(hours) * time.Hour)
		case "days":
		case "day":
			days, err := strconv.Atoi(titleFormat[0])
			if err != nil {
				return time.Time{}, fmt.Errorf("can't convert: %w", err)
			}
			lastUpdateDate = time.Now().Add(-time.Duration(days) * time.Hour * 24)
		case "min":
		case "mins":
			mins, err := strconv.Atoi(titleFormat[0])
			if err != nil {
				return time.Time{}, fmt.Errorf("can't convert: %w", err)
			}
			lastUpdateDate = time.Now().Add(-time.Duration(mins) * time.Minute)
		default:
			return time.Time{}, errors.New("bat 'title' argument")
		}
	} else if chapterSelection.Find("i").Nodes != nil {
		lastUpdateText := chapterSelection.Find("i").Text()

		layout := "January 2, 2006"

		var err error
		lastUpdateDate, err = time.Parse(layout, lastUpdateText)
		if err != nil {
			return time.Time{}, fmt.Errorf("can't convert: %w", err)
		}
	} else {
		return time.Time{}, errors.New("can't find nodes")
	}

	return lastUpdateDate, nil
}
