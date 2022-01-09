package main

import "github.com/fieldryand/goflow"

func main() {
	gf := goflow.New(goflow.Options{StreamJobRuns: true, ShowExamples: true})
	gf.Use(goflow.DefaultLogger())
	gf.Run(":8181")
}
