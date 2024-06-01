package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/4aykovski/manga-parser/manhwaclan"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))
	parser := manhwaclan.New(log, 10)
	urls := []string{
		"https://manhwaclan.com/manga/one-punch-man",
		"https://manhwaclan.com/manga/one-step-forward-to-the-flower-path/",
		"https://manhwaclan.com/manga/the-lost-cinderella/",
	}
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, url := range urls {
		go func(url string) {
			defer wg.Done()
			parser.Parse(url)
		}(url)
	}
	wg.Wait()

	parser.Close()

	errors := parser.Errors()
	for err := range errors {
		log.Error("Error", err)
	}

	projects := parser.Projects()
	for _, project := range projects {
		fmt.Println(project)
		log.Info("Project", slog.String("name", project.Name), slog.String("url", project.Url))
	}
}
