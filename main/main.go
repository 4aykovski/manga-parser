package main

import (
	"log/slog"

	"github.com/4aykovski/manga-parser/manhwaclan"
)

func main() {
	parser := manhwaclan.New(slog.Default(), 10)
	parser.Parse("https://manhwaclan.com/manga/why-would-a-villainess-have-virtues/")
}
