package main

import (
	"encoding/json"
	"github.com/fieldryand/goflow/core"
	"github.com/gin-gonic/gin"
	"net/http"
)

//go:generate go run jobs/gen.go

var taskState map[string]string

func main() {
	router := gin.Default()

	router.GET("/job/:name/submit", func(c *gin.Context) {
		name := c.Param("name")
		job := flow(name)()
		taskState = job.TaskState
		reads := make(chan core.ReadOp)
		go job.Run(reads)
		go func() {
			read := core.ReadOp{Resp: make(chan map[string]string)}
			reads <- read
			taskState = <-read.Resp
		}()
		c.String(http.StatusOK, "job submitted\n")
	})

	router.GET("status", func(c *gin.Context) {
		encoded, _ := json.Marshal(taskState)
		c.String(http.StatusOK, string(encoded)+"\n")
	})

	router.Run(":8090")
}
