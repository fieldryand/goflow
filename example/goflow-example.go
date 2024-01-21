package main

import (
	"github.com/fieldryand/goflow/v2"
)

func main() {
	options := goflow.Options{
		Streaming:    true,
		ShowExamples: true,
		WithSeconds:  true,
	}
	gf := goflow.New(options)
	gf.Use(goflow.DefaultLogger())
	gf.Run(":8181")
}
