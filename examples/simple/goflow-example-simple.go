package main

import (
	"context"

	"github.com/fieldryand/goflow/v3"
)

func main() {
	ctx := context.Background()
	options := goflow.Options{
		UIPath:       "ui/",
		ShowExamples: true,
		WithSeconds:  true,
	}
	gf := goflow.New(options)
	gf.Run(ctx, ":8181")
}
