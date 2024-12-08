package main

import (
	"context"

	"github.com/fieldryand/goflow/v3"
	"github.com/philippgille/gokv/file"
)

func main() {
	storeOptions := file.Options{}
	store, err := file.NewStore(storeOptions)
	if err != nil {
		panic(err)
	}
	defer store.Close()

	ctx := context.Background()

	options := goflow.Options{
		Store:        store,
		UIPath:       "ui/",
		ShowExamples: true,
		WithSeconds:  true,
	}
	gf := goflow.New(options)
	gf.RunWithWebserver(ctx, ":8181")
}
