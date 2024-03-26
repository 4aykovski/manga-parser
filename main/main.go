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
		"https://manhwaclan.com/manga/why-would-a-villainess-have-virtues/",
		"https://manhwaclan.com/manga/the-world-without-my-sister-who-everyone-loved/",
		"https://manhwaclan.com/manga/apollos-heart/",
		"https://manhwaclan.com/manga/muscle-joseon/",
		"https://manhwaclan.com/manga/one-punch-man/",
		"https://manhwaclan.com/manga/beauty-and-the-beasts/",
		"https://manhwaclan.com/manga/divorce-is-the-condition/",
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
	fmt.Println("123")
}
