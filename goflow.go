package main

import (
	"fmt"
	"github.com/fieldryand/goflow/jobs"
	"time"
)

func main() {
	stat := make(chan string, 1)
	example := jobs.ExampleJob

	// Submit the job.
	go example.Run(stat)

	// Concurrently, get the status of the tasks.
	// They should  all be "None."
	go example.Status()

	// Wait one second.
	time.Sleep(time.Duration(1) * time.Second)

	// Get the status again. The two tasks upstream
	// of "sleep 2" should be "Success."
	go example.Status()

	// Print the job status.
	jobStatus := <-stat
	fmt.Println("Example job finished:", jobStatus)
}
