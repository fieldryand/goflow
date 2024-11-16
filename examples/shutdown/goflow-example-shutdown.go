package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/fieldryand/goflow/v3"
)

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())
	options := goflow.Options{
		UIPath:       "ui/",
		ShowExamples: true,
		WithSeconds:  true,
	}
	gf := goflow.New(options)
	go gf.RunWithWebserver(ctx, ":8181")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
		log.Println("interrupt")
		cancelCtx()
		time.Sleep(1 * time.Second)
		return
	}
}
