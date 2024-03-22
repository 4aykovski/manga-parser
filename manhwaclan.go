package parser

import (
	"fmt"
	"log/slog"

	"github.com/gocolly/colly"
)

type ManhwaClanParser struct {
	Collector *colly.Collector
	projects  []Project
	errors    chan error
}

func NewManhwaClanParser(log slog.Logger, projectsForOnce int) *ManhwaClanParser {
	projects := make([]Project, projectsForOnce)
	errors := make(chan error)

	collector := colly.NewCollector()

	collector.OnRequest(func(r *colly.Request) {
		log.Info("Visiting", r.URL.String())
	})

	collector.OnError(func(r *colly.Response, err error) {
		errors <- err
	})

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		chaptersList := e.DOM.Find("ul.version-chap")
		if chaptersList.Nodes == nil {
			errors <- fmt.Errorf("URL - %s: %w", e.Request.URL, ErrChaptersNotFound)
			return
		}

		chaptersQuantity := len(chaptersList.Find("li").Nodes)

		titleNameSelection := e.DOM.Find("meta[property='og:title']")
		titleName, ok := titleNameSelection.Attr("content")
		if !ok {
			errors <- fmt.Errorf("URL - %s: %w", e.Request.URL, ErrTitleNameNotFound)
			return
		}

		output = append(output, *title.New(titleName, e.Request.URL.String(), chaptersQuantity))
	})

	return &ManhwaClanParser{Collector: collector}
}

func (p *ManhwaClanParser) Parse() {

}
