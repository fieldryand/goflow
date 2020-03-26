package main

import (
	"encoding/json"
	"fmt"
	"github.com/fieldryand/goflow/core"
	"github.com/fieldryand/goflow/jobs"
	"net/http"
)

var taskState map[string]string

func submit(w http.ResponseWriter, req *http.Request) {
	example := jobs.ExampleJob()
	taskState = example.TaskState
	reads := make(chan core.ReadOp)
	go example.Run(reads)
	go func() {
		read := core.ReadOp{Resp: make(chan map[string]string)}
		reads <- read
		taskState = <-read.Resp
	}()
	fmt.Fprintf(w, "job submitted\n")
}

func status(w http.ResponseWriter, req *http.Request) {
	encoded, _ := json.Marshal(taskState)
	fmt.Fprintf(w, string(encoded)+"\n")
}

func main() {
	http.HandleFunc("/submit", submit)
	http.HandleFunc("/status", status)

	http.ListenAndServe(":8090", nil)
}
