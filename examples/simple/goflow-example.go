package main

import (
	"github.com/fieldryand/goflow/v3"
)

func main() {
	options := goflow.Options{
		UIPath:       "ui/",
		ShowExamples: true,
		WithSeconds:  true,
	}
	gf := goflow.New(options)
	gf.Run(":8181")
}
