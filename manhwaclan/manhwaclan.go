package manhwaclan

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/4aykovski/manga-parser"
	"github.com/gocolly/colly"
)

type ManhwaClanParser struct {
	Collector *colly.Collector
	projects  []parser.Project
	errors    chan error
}

func New(log *slog.Logger, projectsForOnce int) *ManhwaClanParser {
	_ = make([]parser.Project, projectsForOnce)
	errors := make(chan error)

	collector := colly.NewCollector()

	collector.OnRequest(func(r *colly.Request) {
		log.Info("Visiting", r.URL.String())
	})

	collector.OnError(func(r *colly.Response, err error) {
		errors <- err
	})

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		project, _ := collectProjectInfo(e)
		fmt.Println(project)
	})

	return &ManhwaClanParser{Collector: collector}
}

func (p *ManhwaClanParser) Parse(projectUrl string) {

	p.Collector.Visit(projectUrl)
}

func (p *ManhwaClanParser) Errors() <-chan error {
	return p.errors
}

func collectProjectInfo(e *colly.HTMLElement) (parser.Project, error) {
	// get chapters list
	chaptersList := e.DOM.Find("ul.version-chap")
	if chaptersList.Nodes == nil {
		return parser.Project{}, fmt.Errorf("URL - %s: %w", e.Request.URL, parser.ErrChaptersNotFound)
	}

	// get chapters count
	chaptersCount := len(chaptersList.Find("li").Nodes)

	// get project name from meta
	projectName, ok := e.DOM.Find("meta[property='og:title']").Attr("content")
	if !ok {
		return parser.Project{}, fmt.Errorf("URL - %s: %w", e.Request.URL, parser.ErrTitleNameNotFound)
	}

	// get project description from meta
	description, ok := e.DOM.Find("meta[property='og:description']").Attr("content")
	if !ok {
		return parser.Project{}, fmt.Errorf("URL - %s: %w", e.Request.URL, parser.ErrDescriptionNotFound)
	}

	// get project tags
	var tags []string
	tagNodes := e.DOM.Find("div.genres-content").Find("a").Nodes
	for _, tagNode := range tagNodes {
		tags = append(tags, tagNode.FirstChild.Data)
	}

	// get project last updated at
	lastChapter := chaptersList.Find("li").First()

	return parser.Project{
		Name:          projectName,
		Url:           e.Request.URL.String(),
		Tags:          tags,
		ChaptersCount: chaptersCount,
		Chapters:      []parser.Chapter{},
		LastUpdatedAt: time.Now(),
		Description:   description,
	}, nil
}
