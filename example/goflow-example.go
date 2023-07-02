package main

import "github.com/fieldryand/goflow/v2"

func main() {
	options := goflow.Options{
		UIPath:       "ui/",
		Streaming:    true,
		ShowExamples: true,
	}
	gf := goflow.New(options)
	gf.Use(goflow.DefaultLogger())
	gf.Run(":8181")
}
