//Package goflow implements a minimal workflow scheduler.
package goflow

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Starts the application webserver.
func Start(jobs map[string]*Job) {
	taskState := make(map[string]string)

	router := gin.Default()

	router.GET("/job/:name/submit", func(c *gin.Context) {
		name := c.Param("name")
		job := jobs[name]
		taskState = job.TaskState
		reads := make(chan readOp)
		go job.run(reads)
		go func() {
			read := readOp{Resp: make(chan map[string]string)}
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
